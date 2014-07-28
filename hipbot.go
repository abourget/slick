package ahipbot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/abourget/ahipbot/hipchatv2"
	"github.com/garyburd/redigo/redis"
	"github.com/tkawachi/hipchat"
)

const (
	ConfDomain = "conf.hipchat.com"
)

type Hipbot struct {
	configFile string
	config     HipchatConfig

	// Hipchat XMPP client
	client *hipchat.Client

	// replySink sends messages to Hipchat rooms or users
	replySink chan *BotReply
	Users     []hipchatv2.User

	redisConfig RedisConfig
	RedisPool   *redis.Pool
}

func NewHipbot(configFile string) *Hipbot {
	bot := &Hipbot{}
	bot.replySink = make(chan *BotReply)
	bot.configFile = configFile
	return bot
}

func (bot *Hipbot) Reply(msg *BotMessage, reply string) {
	log.Println("Replying:", reply)
	bot.replySink <- msg.Reply(reply)
}

func (bot *Hipbot) ConnectClient() (err error) {
	bot.client, err = hipchat.NewClient(
		bot.config.Username, bot.config.Password, "bot")
	if err != nil {
		return
	}

	for _, room := range bot.config.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		bot.client.Join(room, bot.config.Nickname)
	}

	return
}

func (bot *Hipbot) SetupHandlers() chan bool {
	bot.client.Status("chat")
	disconnect := make(chan bool)
	go bot.client.KeepAlive()
	go bot.replyHandler(disconnect)
	go bot.messageHandler(disconnect)
	go bot.disconnectHandler(disconnect)
	go bot.usersPolling(disconnect)
	log.Println("Bot ready")
	return disconnect
}

func (bot *Hipbot) LoadBaseConfig() {
	if err := checkPermission(bot.configFile); err != nil {
		log.Fatal(err)
	}

	var config1 struct {
		Hipchat HipchatConfig
	}
	err := bot.LoadConfig(&config1)
	if err != nil {
		log.Fatalln("Error loading Hipchat config section: ", err)
	} else {
		bot.config = config1.Hipchat
	}

	var config2 struct {
		Redis RedisConfig
	}
	err = bot.LoadConfig(&config2)
	if err != nil {
		log.Fatalln("Error loading Redis config section: ", err)
	} else {
		bot.redisConfig = config2.Redis
	}
}

func (bot *Hipbot) SetupStorage() error {
	server := bot.redisConfig.Host
	if server == "" {
		server = "127.0.0.1:6379"
	}
	bot.RedisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			// if _, err := c.Do("AUTH", password); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// Test the config
	conn := bot.RedisPool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		log.Println("ERROR: PING to redis failed! ", err)
	} else {
		log.Println("Redis ping successful")
	}
	return err
}

func (bot *Hipbot) LoadConfig(config interface{}) (err error) {
	content, err := ioutil.ReadFile(bot.configFile)
	if err != nil {
		log.Fatalln("ERROR reading config:", err)
		return
	}
	err = json.Unmarshal(content, &config)
	return
}

func (bot *Hipbot) replyHandler(disconnect chan bool) {
	for {
		reply := <-bot.replySink
		if reply != nil {
			bot.client.Say(reply.To, bot.config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bot *Hipbot) messageHandler(disconnect chan bool) {
	msgs := bot.client.Messages()
	for {
		msg := <-msgs
		botMsg := &BotMessage{Message: msg}
		log.Println("MESSAGE", msg, bot.config.Username, msg.To)

		atMention := "@" + bot.config.Mention
		toMyself := strings.HasPrefix(msg.To, bot.config.Username)
		if strings.Contains(msg.Body, atMention) || strings.HasPrefix(msg.Body, bot.config.Mention) || toMyself {
			botMsg.BotMentioned = true
			log.Printf("Message to me from %s: %s\n", msg.From, msg.Body)
		}

		for _, p := range loadedPlugins {
			pluginConf := p.Config()

			fromMyself := strings.HasPrefix(botMsg.FromNick(), bot.config.Nickname)
			if !pluginConf.EchoMessages && fromMyself {
				continue
			}
			if pluginConf.OnlyMentions && !botMsg.BotMentioned {
				continue
			}

			p.Handle(bot, botMsg)
		}
	}
}

func (bot *Hipbot) disconnectHandler(disconnect chan bool) {
	select {
	case <-disconnect:
		return
	}
	close(disconnect)
}

func (bot *Hipbot) usersPolling(disconnect chan bool) {
	timeout := time.After(0)
	for {
		select {
		case <-disconnect:
			return
		case <-timeout:
			users, err := hipchatv2.GetUsers(bot.config.HipchatApiToken)
			bot.Users = users
			log.Println("Boo, ", users, err)
		}
		timeout = time.After(3 * time.Minute)
		reply := <-bot.replySink
		if reply != nil {
			bot.client.Say(reply.To, bot.config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// GetUser returns a hipchatv2.User by JID, ID, Name or Email
func (bot *Hipbot) GetUser(find string) *hipchatv2.User {
	if strings.Contains(find, "/") {
		find = strings.Split(find, "/")[0]
	}
	for _, user := range bot.Users {
		log.Printf("Hmmmm, %#v\n", user)
		if user.Email == find || user.JID == find || strconv.FormatInt(user.ID, 10) == find || user.Name == find {
			return &user
		}
	}
	return nil
}
