package main

import (
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"io"
	"fmt"
)

type Asana struct {
	client *http.Client
	config *AsanaConfig
	key    string
}

type AsanaConfigSection struct {
	Asana AsanaConfig
}

type AsanaConfig struct {
	ApiKey            string `json:"apikey"`
	OrganizationId    string `json:"organizationID"`
	TeamId            string `json:"teamId"`
}

type Task struct {
	Id   int64
	Name string
}

type taskData struct {
	Data []Task
}


func launchAsana() {
	var conf AsanaConfigSection
	bot.LoadConfig(&conf)

	asana = &Asana{
		client: &http.Client{},
		config: &conf.Asana,
		key: conf.Asana.ApiKey,
	}



	log.Println(asana.config.ApiKey)

	data, err := asana.getAssigneeTasks()

	if err == nil {
		log.Fatal("getAssigneeError", err)
	}
	log.Println("DATA", data)

}

func (asana *Asana) request(method string, uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return req, fmt.Errorf("Cannot create request %s %s", method, uri, err)
	}
	req.SetBasicAuth(asana.key, "")
	return req, nil
}


func (asana *Asana) getAssigneeTasks() ([]Task, error) {
	td := taskData{}
	workspace := asana.config.OrganizationId
	userId := "7940485108532"

	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/tasks?assignee=%s",
		workspace, userId)

	log.Println(url)

	req, err := asana.request("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := asana.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return td.Data, fmt.Errorf("error reading body: %s", err)
	}
	err = json.Unmarshal(body, &td)

	log.Println(td.Data)
	return td.Data, err
}
