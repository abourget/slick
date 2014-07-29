package storm

import (
	"log"
	"io/ioutil"
	"strings"
	"os"
)


var stormTaskFile = "seenStormTasks.txt"
var asanaLink = "https://app.asana.com/0/7221799638526/"

func readTasks() []string {
	content, err := ioutil.ReadFile(stormTaskFile)
	if err != nil {
		log.Println("Storm: using blank seenStormTasks.txt file, ", err)
		content = []byte("")
	}
	stormedTasks := strings.Split(string(content), "\n")

	for idx, task := range stormedTasks {
		stormedTasks[idx] = strings.TrimSpace(task)
	}

	return stormedTasks
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
		if b == a {
			return true
		}
	}
	return false
}
