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
	AcceptList = []string{"In progress", "Dev complete", "Ready for Review", "Waiting for demo"}
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

func (a *JiraClient) getIssueTickets(releaseDate string) ([]*Ticket, error) {
	tmpl := url.QueryEscape(fmt.Sprintf("project = LT AND summary ~ %v AND issuetype = Release", releaseDate))
	resp, err := a.get(fmt.Sprintf("https://manabie.atlassian.net/rest/api/3/search?jql=%s", tmpl))
	if err != nil {
		return nil, fmt.Errorf("error when fetching issues from search endpoint: %w", err)
	}

	var result Response

	if err := json.Unmarshal(resp, &result); err != nil {
		log.Println(err)
		return nil, err
	}

	if len(result.Issues) == 0 {
		return nil, fmt.Errorf("issue list is empty")
	}

	var tickets []*Ticket

	for _, v := range result.Issues[0].Fields.IssueLinks {
		tickets = append(tickets, &Ticket{
			ID:     v.OutwardIssue.Key,
			Status: v.OutwardIssue.Fields.Status.Name,
		})
	}

	return tickets, nil
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

	jiraTickets, err := a.getIssueTickets(*releaseDateFlag)

	if err != nil {
		panic(err)
	}

	ticketMap := make(map[string]*Ticket)

	for _, v := range jiraTickets {
		ticketMap[v.ID] = v
	}

	var message []string

	for _, v := range commitTicketIDs {
		_, ok := ticketMap[v]
		if !ok {
			message = append(message, fmt.Sprintf("%s is not noted in release ticket", v))
		}
	}

	b, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	log.Println(string(b))
	os.Setenv("MESSAGES", string(b))
}
