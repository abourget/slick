package healthy

import (
	"github.com/tkawachi/hipbot/plugin"
	"github.com/tkawachi/hipchat"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Healthy struct {
	urls []string
}

func New(urls []string) *Healthy {
	healthy := new(Healthy)
	healthy.urls = urls
	return healthy
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

func (healthy *Healthy) Handle(msg *hipchat.Message) *plugin.HandleReply {
	switch msg.Body {
	case "health", "healthy?", "health check", "Health check",
		"Health Check", "health_check", "healthcheck":
		log.Println("Health check. Requested by", msg.From)
		return &plugin.HandleReply{
			To:      msg.From,
			Message: healthy.CheckAll(),
		}
	}
	return nil
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
