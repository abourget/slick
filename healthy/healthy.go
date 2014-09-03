package healthy

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/abourget/ahipbot"
)

// Hipbot Plugin
type Healthy struct {
	urls   []string
	config *ahipbot.PluginConfig
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		healthy := &Healthy{
			config: &ahipbot.PluginConfig{
				EchoMessages: false,
				OnlyMentions: true,
			},
		}

		var conf struct {
			HealthCheck struct {
				Urls []string
			}
		}
		bot.LoadConfig(&conf)

		healthy.urls = conf.HealthCheck.Urls
		return healthy
	})
}

// Configuration
func (healthy *Healthy) Config() *ahipbot.PluginConfig {
	return healthy.config
}

// Handler
func (healthy *Healthy) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
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
