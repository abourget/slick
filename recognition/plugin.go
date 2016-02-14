package recognition

import (
	"log"

	"github.com/abourget/slick"
	"github.com/boltdb/bolt"
)

type Plugin struct {
	bot    *slick.Bot
	config Config
	store  Store
}

func init() {
	slick.RegisterPlugin(&Plugin{})
}

func (p *Plugin) InitPlugin(bot *slick.Bot) {
	p.bot = bot

	err := bot.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatalln("Couldn't create the `recognition` bucket")
	}

	var conf struct {
		Recognition Config
	}
	bot.LoadConfig(&conf)
	p.config = conf.Recognition

	p.store = &boltStore{db: bot.DB}

	p.listenRecognize()
	p.listenUpvotes()
}
