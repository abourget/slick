package main

import (
	"os"
	"fmt"
)

func configFile() string {
	return os.Getenv("HOME") + "/.hipbot"
}

func main () {
	config := loadConfig(configFile())
	fmt.Println(config, len(config.Hipchat.Rooms))
}
