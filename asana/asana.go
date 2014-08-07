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
	Id        int64
	Name      string
	Assignee  User
	Completed bool
	Followers []User
	Notes     string
	Projects  []Project
	Hearted   bool
	Hearts    []User
	NumHearts int `json:"num_hearts"`
	Tags      []Tag
	Workspace Workspace
}

type Story struct {
	Id        int64
	Text      string
	Type      string
	CreatedBy User   `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	Id         int64
	Name       string
	Email      string
	Photo      *Photo
	Workspaces []Workspace
}

type Workspace struct {
	Id   int64
	Name string
}

type Project struct {
	Id   int64
	Name string
}

type Photo struct {
	Image21  string `json:"image_21x21"`
	Image27  string `json:"image_27x27"`
	Image36  string `json:"image_36x36"`
	Image60  string `json:"image_60x60"`
	Image128 string `json:"image_128x128"`
}

type Tag struct {
	Id   int64
	Name string
}

func (t *Tag) StringId() string {
	return fmt.Sprintf("%v", t.Id)
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
	var data struct {
		Data []Task
	}
	workspace := asana.workspace

	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/tasks?assignee=%s",
		workspace, userId)

	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	return data.Data, err
}

func (asana *Client) GetTasksByTag(tagId string) ([]Task, error) {
	var data struct {
		Data []Task
	}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/tags/%s/tasks",
		tagId)

	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	return data.Data, err
}

func (asana *Client) GetTaskStories(taskId int64) ([]Story, error) {
	var data struct {
		Data []Story
	}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/tasks/%v/stories", taskId)
	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	return data.Data, err
}

func (asana *Client) GetUser(userId int64) (*User, error) {
	var data struct {
		Data User
	}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/users/%v", userId)
	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	return &data.Data, err
}

func (asana *Client) GetTaskById(taskId int64) (*Task, error) {
	var data struct {
		Data Task
	}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/tasks/%v", taskId)
	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	fmt.Println("The tasks content: ", string(body))

	return &data.Data, err
}

func (asana *Client) GetTagsOnTask(tagId int64) ([]Tag, error) {
	var data struct {
		Data []Tag
	}

	url := fmt.Sprintf("https://app.asana.com/api/1.0/tasks/%v/tags",
		tagId)

	body, err := asana.request("GET", url)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &data)

	return data.Data, err
}
