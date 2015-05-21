package healthy

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/abourget/slick"
)

// Hipbot Plugin
type Healthy struct {
	urls   []string
	config *slick.ChatPluginConfig
}

func init() {
	slick.RegisterPlugin(&Healthy{})
}

func (healthy *Healthy) InitChatPlugin(bot *slick.Bot) {
	healthy.config = &slick.ChatPluginConfig{
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

	bot.ListenFor(&slick.Conversation{
		ContainsAny: []string{"health", "healthy?", "health_check"},
		HandlerFunc: healthy.ChatHandler,
	})
}

// Handler
func (healthy *Healthy) ChatHandler(conv *slick.Conversation, msg *slick.Message) {
	log.Println("Health check. Requested by", msg.FromUser.Name)
	conv.Reply(msg, healthy.CheckAll())
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
