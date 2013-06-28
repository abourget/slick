package main

import (
	"code.google.com/p/gcfg"
	"log"
	"os"
	"errors"
)

type Config struct {
	Hipchat struct {
		Server string
		Port uint
		Id string
		Password string
		Rooms []string
	}
}

func checkPermission(file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Mode() & 0077 != 0 {
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
