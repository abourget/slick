package healthy

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/plotly/plotbot"
)

// Hipbot Plugin
type Healthy struct {
	urls   []string
	config *plotbot.ChatPluginConfig
}

func init() {
	plotbot.RegisterPlugin(&Healthy{})
}

func (healthy *Healthy) InitChatPlugin(bot *plotbot.Bot) {
	healthy.config = &plotbot.ChatPluginConfig{
		EchoMessages: false,
		OnlyMentions: true,
	}

	var conf struct {
		HealthCheck struct {
			Urls []string
		}
	}
	bot.LoadConfig(&conf)

	healthy.urls = conf.HealthCheck.Urls
}

// Configuration
func (healthy *Healthy) ChatConfig() *plotbot.ChatPluginConfig {
	return healthy.config
}

// Handler
func (healthy *Healthy) ChatHandler(bot *plotbot.Bot, msg *plotbot.Message) {
	if msg.ContainsAny([]string{"health", "healthy?", "health_check"}) {
		log.Println("Health check. Requested by", msg.From)
		bot.Reply(msg, healthy.CheckAll())
	}
}

func (healthy *Healthy) CheckAll() string {
	result := make(map[string]bool)
	failed := make([]string, 0)
	for _, url := range healthy.urls {
		ok := check(url)
		result[url] = ok
		if !ok {
			failed = append(failed, url)
		}
	}
	if len(failed) == 0 {
		return "All green (For " +
			strings.Join(healthy.urls, ", ") + ")"
	} else {
		return "WARN!! Something wrong with " +
			strings.Join(failed, ", ")
	}
}

func check(url string) bool {
	res, err := http.Get(url)
	if err != nil {
		return false
	}
	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return false
	}
	if res.StatusCode/100 != 2 {
		return false
	}
	return true
}
