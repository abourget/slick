package main

import (
	"time"
	"fmt"
	"log"
	"github.com/bpostlethwaite/ahipbot/asana"
	"io/ioutil"
	"strings"
	"strconv"
	"os"
)


var stormTaskFile = "seenStormTasks.txt"

func StormWatch(asanaClient *asana.AsanaClient) {

	stormId := "15014460778242"
	var stormedTasks []string
	var tasks []asana.Task
	var err error
	var taskAlreadyStormed bool
	var taskId string
	for {

		if bot.stormMode {
			return
		}
		stormedTasks, err = readTasks()


		tasks, err = asanaClient.GetTasksByTag(stormId)
		if err != nil {
			log.Println("Storm Check Error: ", err)
		}

		for i := 0; i < len(tasks); i++ {
			taskId = strconv.FormatInt(tasks[i].Id, 10)
			taskAlreadyStormed = stringInSlice(taskId, stormedTasks)

			if !taskAlreadyStormed {
				bot.stormMode = true
				writeTask(taskId)
				break
			}

		}

		time.Sleep(5 * time.Second)
	}

}


func readTasks() (stormedTasks []string, err error) {
	content, err := ioutil.ReadFile(stormTaskFile)
	if err != nil {
		log.Fatalln("ERROR reading seenStorms:", err)
		return nil, err
	}
	stormedTasks = strings.Split(string(content), "\n")

	for idx, task := range stormedTasks {
		stormedTasks[idx] = strings.TrimSpace(task)
	}


	fmt.Println(stormedTasks)
	return stormedTasks, nil
}

func writeTask(taskId string) error {
	file, err := os.OpenFile(stormTaskFile, os.O_RDWR|os.O_APPEND, 0660);

	if err != nil {
		return err
	}

	_, err = file.WriteString(taskId + "\n")

	if err != nil {
		return err
	}

	return nil
}


func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		fmt.Println(a, b)
		if b == a {
			return true
		}
	}
	return false
}
