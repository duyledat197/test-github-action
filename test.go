package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	DATE_FMT = "20060102"
	Comma    = ","
	Bracket  = "[]"
)

var (
	WorkStateList = []string{"In progress", "Dev complete", "Ready for Review", "Waiting for demo"}
)

type JiraClient struct {
	UserName  string
	UserToken string

	RgIssueKey     *regexp.Regexp
	RepoBaseBranch map[string]string
	ReleaseDay     time.Weekday

	ConcurentLimit int
}

type Response struct {
	Issues []struct {
		Key    string `json:"key,omitempty"`
		Fields struct {
			IssueLinks []struct {
				OutwardIssue struct {
					Key    string `json:"key,omitempty"`
					Fields struct {
						Status struct {
							Name string `json:"name,omitempty"`
						} `json:"status,omitempty"`
					} `json:"fields,omitempty"`
				} `json:"outwardIssue,omitempty"`
			} `json:"issuelinks,omitempty"`
		} `json:"fields,omitempty"`
	} `json:"issues,omitempty"`
}

type Ticket struct {
	ID     string
	Status string
}

func (a *JiraClient) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return a.call(url, req)
}

func (a *JiraClient) call(url string, req *http.Request) ([]byte, error) {
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(a.UserName, a.UserToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}
	return body, nil
}

func (a *JiraClient) getIssueTickets(releaseDate string) ([]*Ticket, string, error) {
	tmpl := url.QueryEscape(fmt.Sprintf("project = LT AND summary ~ %v AND issuetype = Release", releaseDate))
	resp, err := a.get(fmt.Sprintf("https://manabie.atlassian.net/rest/api/3/search?jql=%s", tmpl))
	if err != nil {
		return nil, "", fmt.Errorf("error when fetching issues from search endpoint: %w", err)
	}

	var result Response

	if err := json.Unmarshal(resp, &result); err != nil {
		log.Println(err)
		return nil, "", err
	}

	if len(result.Issues) == 0 {
		return nil, "", fmt.Errorf("issue list is empty")
	}

	var tickets []*Ticket

	for _, v := range result.Issues[0].Fields.IssueLinks {
		tickets = append(tickets, &Ticket{
			ID:     v.OutwardIssue.Key,
			Status: v.OutwardIssue.Fields.Status.Name,
		})
	}

	return tickets, result.Issues[0].Key, nil
}

func isTicketInWorkState(status string) bool {
	for _, v := range WorkStateList {
		if status == v {
			return true
		}
	}
	return false
}

func main() {
	jiraUserFlag := flag.String("user", "", "JIRA user name, eg: devops@manabie.com")
	jiraTokenFlag := flag.String("token", "", "JIRA user token")
	releaseDateFlag := flag.String("releaseDate", "", "Release date in yyyymmdd fmt, eg: 20160101")
	commitTicketIDsFlag := flag.String("commitTicketIDs", "", "List commit tickets, eg: [LT-8586,LT-1234]")

	flag.Parse()

	commitTicketIDs := strings.Split(strings.Trim(*commitTicketIDsFlag, Bracket), Comma)
	a := &JiraClient{
		UserName:  *jiraUserFlag,
		UserToken: *jiraTokenFlag,
	}
	_, err := time.Parse(DATE_FMT, *releaseDateFlag)
	if err != nil {
		panic(err)
	}

	jiraTickets, mainTicketID, err := a.getIssueTickets(*releaseDateFlag)

	if err != nil {
		panic(err)
	}

	ticketMap := make(map[string]*Ticket)

	for _, v := range jiraTickets {
		ticketMap[v.ID] = v
	}

	var messages []string
	var statusMessages []string

	for _, v := range commitTicketIDs {
		ticket, ok := ticketMap[v]
		if !ok {
			messages = append(messages, v)
		} else {
			if !isTicketInWorkState(ticket.Status) {
				statusMessages = append(statusMessages, v)
			}
		}

	}

	var msg string

	log.Println(messages)
	log.Println(statusMessages)

	if len(messages) > 0 {
		msg = strings.Join(messages, ", ")
		msg += fmt.Sprintf(" isn't noted in release ticket (%s)", mainTicketID)
	}

	if len(statusMessages) > 0 {
		sMsg := strings.Join(statusMessages, ", ")
		sMsg = fmt.Sprintf(" isn't in work state (status must in %s)", sMsg)
		msg += fmt.Sprintf("\n%v", sMsg)
	}

	if len(msg) > 0 {
		f, err := os.Create(".env")
		if err != nil {
			panic(err)
		}
		if _, err := f.Write([]byte(msg)); err != nil {
			panic(err)
		}

		defer f.Close()
	}
}
