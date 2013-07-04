package main

import (
	"code.google.com/p/gcfg"
	"errors"
	"log"
	"os"
)

type Hipchat struct {
	Username string
	Password string
	Nickname string
	Resource string
	Rooms    []string
}

type HealthCheck struct {
	Url []string
}

type Config struct {
	Hipchat     Hipchat
	HealthCheck HealthCheck
}

func checkPermission(file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Mode()&0077 != 0 {
		return errors.New("Config file is permitted to group/other. Do chmod 600 " + file)
	}
	return nil
}

func loadConfig(file string) Config {
	if err := checkPermission(file); err != nil {
		log.Fatal(err)
	}

	var config Config
	err := gcfg.ReadFileInto(&config, file)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
