package asana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	workspace  string
	key        string
}

type Task struct {
	Id   int64
	Name string
}

type taskData struct {
	Data []Task
}

func NewClient(key, workspace string) *Client {
	return &Client{
		workspace:  workspace,
		key:        key,
		httpClient: &http.Client{},
	}
}

func (asana *Client) SetWorkspace(workspace string) {
	asana.workspace = workspace
}

func (asana *Client) request(method string, uri string) ([]byte, error) {

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create request %s %s", method, uri, err)
	}
	req.SetBasicAuth(asana.key, "")

	res, err := asana.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Asana request error %s %s %s", method, uri, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (asana *Client) GetTasksByAssignee(userId string) ([]Task, error) {
	td := taskData{}
	workspace := asana.workspace

	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/tasks?assignee=%s",
		workspace, userId)

	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &td)

	return td.Data, err
}

func (asana *Client) GetTasksByTag(tagId string) ([]Task, error) {
	td := taskData{}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/tags/%s/tasks",
		tagId)

	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &td)

	return td.Data, err
}
