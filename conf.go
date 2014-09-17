package plotbot

import (
	"errors"
	"os"
)

type HipchatConfig struct {
	Username        string
	Password        string
	Resource        string
	Nickname        string
	Mention         string
	Rooms           []string
	HipchatApiToken string `json:"hipchat_api_token"`
	WebBaseURL      string `json:"web_base_url"`
}

type RedisConfig struct {
	Host string
}

type ChatPluginConfig struct {
	// Whether to handle the bot's own messages
	EchoMessages bool

	// Whether to handle messages that are not destined to me
	OnlyMentions bool
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
