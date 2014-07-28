package asana

import (
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"fmt"
)

type AsanaClient struct {
	client *http.Client
	workspace string
	key       string
}


type Task struct {
	Id   int64
	Name string
}

type taskData struct {
	Data []Task
}


func NewClient(key, workspace string) (*AsanaClient, error) {

	asana := &AsanaClient{
		workspace: workspace,
		key: key,
		client: &http.Client{},
	}

	return asana, nil
}

func (asana *AsanaClient) SetWorkspace (workspace string) {
	asana.workspace = workspace
}

func (asana *AsanaClient) request(method string, uri string) ([]byte, error) {

	log.Println("ASANA REQ", uri)

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create request %s %s", method, uri, err)
	}
	log.Println(asana.key)

	req.SetBasicAuth(asana.key, "")

	res, err := asana.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Asana request error %s %s %s", method, uri, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}


func (asana *AsanaClient) GetTasksByAssignee(userId string) ([]Task, error) {
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


func (asana *AsanaClient) GetTasksByTag(tagId string) ([]Task, error) {
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
