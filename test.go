package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	DATE_FMT = "20060102"
)

type ReleaseAssistant struct {
	UserName  string
	UserToken string

	RgIssueKey     *regexp.Regexp
	RepoBaseBranch map[string]string
	ReleaseDay     time.Weekday

	ConcurentLimit int
}

type Data struct {
	Issues []struct {
		IssueLinks []struct {
			OutwardIssue struct {
				Key    string `json:"key,omitempty"`
				Status struct {
					Name string `json:"name,omitempty"`
				} `json:"status,omitempty"`
			} `json:"outwardIssue,omitempty"`
		} `json:"issuelinks,omitempty"`
	} `json:"issues"`
}

func (a *ReleaseAssistant) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return a.call(url, req)
}

func (a *ReleaseAssistant) call(url string, req *http.Request) ([]byte, error) {
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

func (a *ReleaseAssistant) searchRelease(when time.Time) (string, string, error) {
	log.Println("searchRelease", when)

	releaseDate := when.Format(DATE_FMT)
	tmpl := url.QueryEscape(`project = LT AND summary ~ "` + releaseDate + `" AND issuetype = Release`)
	resp, err := a.get(fmt.Sprintf("https://manabie.atlassian.net/rest/api/3/search?jql=%s", tmpl))
	if err != nil {
		log.Println(err)
		return "", "", fmt.Errorf("error when fetching issues from search endpoint: %w", err)
	}

	var result Data

	if err := json.Unmarshal(resp, &result); err != nil {
		log.Println(err)
		return "", "", err
	}
	log.Println(result)
	return "", "", nil
}

func main() {
	jiraUserFlag := flag.String("user", "", "JIRA user name, eg: devops@manabie.com")
	jiraTokenFlag := flag.String("token", "", "JIRA user token")
	releaseDateFlag := flag.String("releaseDate", "", "Release date in yyyymmdd fmt, eg: 20160101")

	flag.Parse()
	log.Println("jiraUserFlag", *jiraUserFlag)
	log.Println("jiraTokenFlag", *jiraTokenFlag)

	a := &ReleaseAssistant{
		UserName:  *jiraUserFlag,
		UserToken: *jiraTokenFlag,

		RgIssueKey: regexp.MustCompile(`LT-[0-9]{1,6}`),
		RepoBaseBranch: map[string]string{
			"manabie-com/backend":             "develop",
			"manabie-com/school-portal-admin": "develop",
			"manabie-com/student-app":         "develop",
			"manabie-com/eibanam":             "develop",
		},
		ReleaseDay:     time.Thursday,
		ConcurentLimit: 5,
	}
	if releaseDay, err := time.Parse(DATE_FMT, *releaseDateFlag); err == nil {
		log.Println(a.searchRelease(releaseDay))
		a.searchRelease(releaseDay)
	} else {
		log.Println(err)
	}

}
