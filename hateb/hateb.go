package hateb

import (
	"errors"
	"github.com/tkawachi/hipbot/plugin"
	"github.com/tkawachi/hipchat"
	//"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"
)

// はてなブックマークのランキングから記事を紹介する
type Hateb struct {
	ranking []string
	idx     int
	fetchAt time.Time
}

func New() *Hateb {
	h := new(Hateb)
	h.ranking = make([]string, 0)
	h.idx = 0
	return h
}

func (h *Hateb) Handle(msg *hipchat.Message) *plugin.HandleReply {
	if msg.Body == "hateb" {
		reply := h.NextMessage()
		if reply != nil {
			return &plugin.HandleReply{
				To:      msg.From,
				Message: *reply,
			}
		}
	}
	return nil
}

func (h *Hateb) NextMessage() *string {
	if time.Since(h.fetchAt) > 6*time.Hour {
		ranking, err := GetRanking()
		if err != nil {
			log.Println(err)
			h.ranking = []string{}
		} else {
			h.ranking = ranking
			h.fetchAt = time.Now()
		}
	}
	if len(h.ranking) == 0 {
		return nil
	}
	h.idx %= len(h.ranking)
	reply := h.ranking[h.idx]
	h.idx += 1
	return &reply
}

// はてぶランキングを取得して、チャットのメッセージに流せる形に整形する
func GetRanking() ([]string, error) {
	url, err := GetItUrl()
	if err != nil {
		return nil, err
	}
	log.Println("Getting IT Ranking")
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	msgs := doc.Find("h3>.entry-link").Map(func(_ int, sel *goquery.Selection) string {
		link, exists := sel.Attr("href")
		msg := sel.Text()
		if exists {
			msg += " " + link
		}
		return msg
	})
	return msgs, nil
}

const (
	HatebUrl      = "http://b.hatena.ne.jp"
	RankingTopUrl = HatebUrl + "/ranking"
)

// テクノロジーのURLをゲット
func GetItUrl() (string, error) {
	log.Println("Getting IT URL")
	doc, err := goquery.NewDocument(RankingTopUrl)
	if err != nil {
		return "", err
	}
	link, exists := doc.Find("a.it").Attr("href")
	if !exists {
		return "", errors.New("href does not exist")
	}
	if !strings.Contains(link, "://") {
		if link[0] != '/' {
			link = "/" + link
		}
		link = HatebUrl + link
	}
	return link, nil
}
