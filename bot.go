package ahipbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/abourget/ahipbot/hipchatv2"
	"github.com/garyburd/redigo/redis"
	"github.com/jmcvetta/napping"
	"github.com/tkawachi/hipchat"
)

const (
	ConfDomain = "conf.hipchat.com"
)

type Bot struct {
	configFile string
	Config     HipchatConfig

	// Hipchat XMPP client
	client *hipchat.Client

	// replySink sends messages to Hipchat rooms or users
	replySink chan *BotReply
	Users     []hipchatv2.User
	Rooms     []hipchatv2.Room

	redisConfig RedisConfig
	// RedisPool holds a connection to Redis.  NOTE: Prefix all your keys with "plotbot:" please.
	RedisPool *redis.Pool
}

func NewHipbot(configFile string) *Bot {
	bot := &Bot{}
	bot.replySink = make(chan *BotReply)
	bot.configFile = configFile
	return bot
}

func (bot *Bot) Run() {
	bot.LoadBaseConfig()
	bot.SetupStorage()

	// Web related
	LoadPlugins(bot)
	LoadWebHandler(bot)

	// Bot related
	for {
		log.Println("Connecting client...")
		err := bot.ConnectClient()
		if err != nil {
			log.Println("  `- Failed: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		disconnect := bot.SetupHandlers()

		select {
		case <-disconnect:
			log.Println("Disconnected...")
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (bot *Bot) Reply(msg *BotMessage, reply string) {
	log.Println("Replying:", reply)
	bot.replySink <- msg.Reply(reply)
}

func (bot *Bot) Notify(room, color, format, msg string, notify bool) (*napping.Request, error) {
	log.Println("Notifying room ", room, ": ", msg)
	return hipchatv2.SendNotification(bot.Config.HipchatApiToken, bot.GetRoomId(room), color, format, msg, notify)
}

func (bot *Bot) SendToRoom(room string, message string) {
	log.Println("Sending to room ", room, ": ", message)

	room = canonicalRoom(room)

	reply := &BotReply{
		To:      room,
		Message: message,
	}
	bot.replySink <- reply
	return
}

// GetRoomdId returns the numeric room ID as string for a given XMPP room JID
func (bot *Bot) GetRoomId(room string) string {
	roomName := canonicalRoom(room)
	for _, room := range bot.Rooms {
		if roomName == room.JID {
			return fmt.Sprintf("%v", room.ID)
		}
	}
	return room
}

func canonicalRoom(room string) string {
	if !strings.Contains(room, "@") {
		room += "@conf.hipchat.com"
	}
	return room
}

func baseRoom(room string) string {
	if strings.Contains(room, "@") {
		return strings.Split(room, "@")[0]
	}
	return room
}

func (bot *Bot) ConnectClient() (err error) {
	bot.client, err = hipchat.NewClient(
		bot.Config.Username, bot.Config.Password, "bot")
	if err != nil {
		log.Println("Error in ConnectClient()")
		return
	}

	for _, room := range bot.Config.Rooms {
		if !strings.Contains(room, "@") {
			room = room + "@" + ConfDomain
		}
		bot.client.Join(room, bot.Config.Nickname)
	}

	return
}

func (bot *Bot) SetupHandlers() chan bool {
	bot.client.Status("chat")
	disconnect := make(chan bool)
	go bot.client.KeepAlive()
	go bot.replyHandler(disconnect)
	go bot.messageHandler(disconnect)
	go bot.disconnectHandler(disconnect)
	go bot.usersPolling(disconnect)
	go bot.roomsPolling(disconnect)
	log.Println("Bot ready")
	return disconnect
}

func (bot *Bot) LoadBaseConfig() {
	if err := checkPermission(bot.configFile); err != nil {
		log.Fatal("ERROR Checking Permissions: ", err)
	}

	var config1 struct {
		Hipchat HipchatConfig
	}
	err := bot.LoadConfig(&config1)
	if err != nil {
		log.Fatalln("Error loading Hipchat config section: ", err)
	} else {
		bot.Config = config1.Hipchat
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

func (bot *Bot) SetupStorage() error {
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
			if _, err := c.Do("SELECT", "3"); err != nil {
				c.Close()
				return nil, err
			}
			// if _, err := c.Do("AUTH", password); err != nil {
			//  c.Close()
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

func (bot *Bot) LoadConfig(config interface{}) (err error) {
	content, err := ioutil.ReadFile(bot.configFile)
	if err != nil {
		log.Fatalln("ERROR reading config:", err)
		return
	}
	err = json.Unmarshal(content, &config)
	return
}

func (bot *Bot) replyHandler(disconnect chan bool) {
	for {
		reply := <-bot.replySink
		if reply != nil {
			log.Println("REPLYING", reply.To, reply.Message)
			bot.client.Say(reply.To, bot.Config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bot *Bot) messageHandler(disconnect chan bool) {
	msgs := bot.client.Messages()
	for {
		msg := <-msgs
		botMsg := &BotMessage{Message: msg}
		//log.Println("MESSAGE", msg, bot.Config.Username, msg.To)

		atMention := "@" + bot.Config.Mention
		toMyself := strings.HasPrefix(msg.To, bot.Config.Username)
		fromMyself := strings.HasPrefix(botMsg.FromNick(), bot.Config.Nickname)

		if strings.Contains(msg.Body, atMention) || strings.HasPrefix(msg.Body, bot.Config.Mention) || toMyself {
			botMsg.BotMentioned = true
			//log.Printf("Message to me from %s: %s\n", msg.From, msg.Body)
		}
		for _, p := range loadedPlugins {
			pluginConf := p.Config()

			if pluginConf == nil {
				continue
			}

			if !pluginConf.EchoMessages && fromMyself {
				//log.Printf("no echo but I just messaged myself")
				continue
			}
			if pluginConf.OnlyMentions && !botMsg.BotMentioned {
				//log.Printf("only mentions but not BotMentioned")
				continue
			}

			p.Handle(bot, botMsg)
		}
	}
}

func (bot *Bot) disconnectHandler(disconnect chan bool) {
	select {
	case <-disconnect:
		return
	}
	close(disconnect)
}

func (bot *Bot) usersPolling(disconnect chan bool) {
	timeout := time.After(0)
	for {
		select {
		case <-disconnect:
			return
		case <-timeout:
			// FIXME: Disregard error?! wwoooaah!
			users, _ := hipchatv2.GetUsers(bot.Config.HipchatApiToken)
			bot.Users = users
		}
		timeout = time.After(3 * time.Minute)
		reply := <-bot.replySink
		if reply != nil {
			bot.client.Say(reply.To, bot.Config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bot *Bot) roomsPolling(disconnect chan bool) {
	timeout := time.After(0)
	for {
		select {
		case <-disconnect:
			return
		case <-timeout:
			// FIXME: Disregard error?! wwoooaah!
			rooms, _ := hipchatv2.GetRooms(bot.Config.HipchatApiToken)
			bot.Rooms = rooms
		}
		timeout = time.After(3 * time.Minute)
		reply := <-bot.replySink
		if reply != nil {
			bot.client.Say(reply.To, bot.Config.Nickname, reply.Message)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// GetUser returns a hipchatv2.User by JID, ID, Name or Email
func (bot *Bot) GetUser(find string) *hipchatv2.User {
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
