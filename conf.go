package slick

import (
	"errors"
	"os"
)

type SlackConfig struct {
	Username        string
	Password        string
	Resource        string
	Nickname        string
	JoinChannels    []string `json:"join_channels"`
	GeneralChannel  string   `json:"general_channel"`
	ApiToken        string   `json:"api_token"`
	WebBaseURL      string   `json:"web_base_url"`
}

type LevelConfig struct {
	Path string
}

type ChatPluginConfig struct {
	// Whether to handle the bot's own messages
	EchoMessages bool

	// Whether to handle messages that are not destined to me
	OnlyMentions bool
}

type Config struct {
	Slack SlackConfig `json:"Slack"`
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
