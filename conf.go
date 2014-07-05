package main

import (
	"errors"
	"os"
)

type HipchatConfig struct {
	Username string
	Password string
	Nickname string
	Mention  string
	Rooms    []string
}

type Config struct {
	Hipchat HipchatConfig `json:"Hipchat"`
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
