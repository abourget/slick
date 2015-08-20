package github

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const searchIssueURL = "https://api.github.com/search/issues"

type SearchQuery struct {
	Repo        string
	Labels      []string
	ClosedSince string
}

type SearchResponse struct {
	TotalCount int64 `json:"total_count"`
	Items      []IssueItem
}

type GHUser struct {
	Login string
}

type IssueItem struct {
	Url       string
	Title     string
	ID        int
	Number    int
	Milestone string
	Assignee  GHUser
	State     string
	Events    []IssueEvent
}

type IssueEvent struct {
	ID    int
	Actor GHUser
	Event string
}

type Conf struct {
	Authtoken      string
	Repos          []string
	Github2Hipchat map[string]string
}

type Client struct {
	Conf Conf
}

func (query *SearchQuery) Url() (url string) {

	url = searchIssueURL

	url += "?q="

	if query.Repo != "" {
		url += "+repo:" + query.Repo
	}

	if len(query.Labels) > 0 {
		for _, value := range query.Labels {
			url += "+label:" + value
		}
	}

	if query.ClosedSince != "" {
		url += "+closed:>" + query.ClosedSince
	}

	return
}

func (issue *IssueItem) LastClosedBy() string {
	for i := len(issue.Events) - 1; i >= 0; i-- {
		event := issue.Events[i]
		if event.Event == "closed" {
			return event.Actor.Login
		}
	}
	return ""
}

func (ghclient *Client) Get(url string) (body []byte, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(ghclient.Conf.Authtoken, "x-oauth-basic")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	body, err = ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	res.Body.Close()

	return
}

func (ghclient *Client) DoSearchQuery(query SearchQuery) ([]IssueItem, error) {

	url := query.Url()
	body, err := ghclient.Get(url)
	if err != nil {
		return nil, err
	}

	payload := SearchResponse{}
	json.Unmarshal(body, &payload)

	return payload.Items, nil
}

func (ghclient *Client) DoEventQuery(issueList []IssueItem, repo string, issueChan chan IssueItem) {

	defer close(issueChan)

	for _, issue := range issueList {

		url := "https://api.github.com/repos/" + repo + "/issues/" + strconv.Itoa(issue.Number) + "/events"
		body, err := ghclient.Get(url)
		if err != nil {
			log.Print(err)
		}

		events := make([]IssueEvent, 0)
		json.Unmarshal(body, &events)

		issue.Events = events
		issueChan <- issue

		time.Sleep(500 * time.Millisecond)

	}

}
