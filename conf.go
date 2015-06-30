package slick

import (
	"errors"
	"os"
)

type SlackConfig struct {
	Username       string
	Password       string
	Nickname       string
	JoinChannels   []string `json:"join_channels"`
	GeneralChannel string   `json:"general_channel"`
	TeamDomain     string   `json:"team_domain"`
	TeamID         string   `json:"team_id"`
	ApiToken       string   `json:"api_token"`
	WebBaseURL     string   `json:"web_base_url"`
	Debug          bool
}

type LevelDBConfig struct {
	Path string
}

type ChatPluginConfig struct {
	// Whether to handle the bot's own messages
	EchoMessages bool

	// Whether to handle messages that are not destined to me
	OnlyMentions bool
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
