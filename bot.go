package plotbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/plotly/plotbot/hipchatv2"
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
	disconnected chan bool
	replySink    chan *BotReply
	Users        []User
	Rooms        []Room

	redisConfig RedisConfig
	// RedisPool holds a connection to Redis.  NOTE: Prefix all your keys with "plotbot:" please.
	RedisPool *redis.Pool
}

func NewHipbot(configFile string) *Bot {
	bot := &Bot{
		configFile: configFile,
		replySink:  make(chan *BotReply, 10),
	}

	return bot
}

func (bot *Bot) Run() {
	bot.LoadBaseConfig()
	bot.SetupStorage()

	// Web related
	LoadPlugins(bot)
	LoadWebHandler(bot)

	//time.Sleep(5000 * time.Second)
	//return

	for {
		log.Println("Connecting client...")
		err := bot.ConnectClient()
		if err != nil {
			log.Println("  `- Failed: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		bot.SetupHandlers()

		select {
		case <-bot.disconnected:
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

// ReplyMention replies with a @mention named prefixed, when replying in public. When replying in private, nothing is added.
func (bot *Bot) ReplyMention(msg *BotMessage, reply string) {
	if msg.IsPrivate() {
		bot.Reply(msg, reply)
	} else {
		prefix := ""
		if msg.FromUser != nil {
			prefix = fmt.Sprintf("@%s ", msg.FromUser.MentionName)
		}
		bot.Reply(msg, fmt.Sprintf("%s%s", prefix, reply))
	}
}

func (bot *Bot) ReplyTo(jid, reply string) {
	log.Println("Replying:", reply)
	rep := &BotReply{
		To:      jid,
		Message: reply,
	}
	bot.replySink <- rep
}

func (bot *Bot) ReplyPrivate(msg *BotMessage, reply string) {
	log.Println("Replying privately:", reply)
	bot.replySink <- msg.ReplyPrivate(reply)
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

func (bot *Bot) ConnectClient() (err error) {
	resource := bot.Config.Resource
	if resource == "" {
		resource = "bot"
	}

	bot.client, err = hipchat.NewClient(
		bot.Config.Username, bot.Config.Password, resource)
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

func (bot *Bot) SetupHandlers() {
	bot.client.Status("chat")
	bot.disconnected = make(chan bool)
	go bot.client.KeepAlive()
	go bot.replyHandler()
	go bot.messageHandler()
	go bot.usersPolling()
	go bot.roomsPolling()
	log.Println("Bot ready")
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
		log.Fatalln("LoadConfig(): Error reading config:", err)
		return
	}
	err = json.Unmarshal(content, &config)

	if err != nil {
		log.Println("LoadConfig(): Error unmarshaling JSON", err)
	}
	return
}

func (bot *Bot) replyHandler() {
	for {
		select {
		case <-bot.disconnected:
			return
		case reply := <-bot.replySink:
			if reply != nil {
				//log.Println("REPLYING", reply.To, reply.Message)
				bot.client.Say(reply.To, bot.Config.Nickname, reply.Message)
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

func (bot *Bot) messageHandler() {
	msgs := bot.client.Messages()
	for {
		select {
		case <-bot.disconnected:
			return
		case msg := <-msgs:
			botMsg := &BotMessage{Message: msg}
			botMsg.FromUser = bot.GetUser(msg.From)
			botMsg.FromRoom = bot.GetRoom(msg.From)
			if botMsg.FromUser == nil {
				log.Printf("Message from unknown user, skipping: %#v\n", botMsg)
				continue
			}

			log.Printf("Incoming message: %s\n", botMsg)

			atMention := "@" + bot.Config.Mention
			fromMyself := strings.HasPrefix(botMsg.FromNick(), bot.Config.Nickname)
			mentionColon := bot.Config.Mention + ":"
			mentionComma := bot.Config.Mention + ","

			if strings.Contains(msg.Body, atMention) || strings.HasPrefix(msg.Body, mentionColon) || strings.HasPrefix(msg.Body, mentionComma) || botMsg.IsPrivate() {
				botMsg.BotMentioned = true
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
}

// Disconnect, you can call many times, checks closed channel first.
func (bot *Bot) Disconnect() {
	select {
	case _, ok := <-bot.disconnected:
		if ok {
			close(bot.disconnected)
		}
	default:
	}
}

func (bot *Bot) usersPolling() {
	timeout := time.After(0)
	for {
		select {
		case <-bot.disconnected:
			return
		case <-timeout:
			// FIXME: Disregard error?! wwoooaah!
			hcUsers, _ := hipchatv2.GetUsers(bot.Config.HipchatApiToken)
			users := []User{}
			for _, hcu := range hcUsers {
				users = append(users, User{hcu})
			}
			bot.Users = users
			//log.Printf("Users: %#v\n", users)
		}
		timeout = time.After(3 * time.Minute)
	}
}

func (bot *Bot) roomsPolling() {
	timeout := time.After(0)
	for {
		select {
		case <-bot.disconnected:
			return
		case <-timeout:
			hcRooms, _ := hipchatv2.GetRooms(bot.Config.HipchatApiToken)
			rooms := []Room{}
			for _, hcu := range hcRooms {
				rooms = append(rooms, Room{hcu})
			}
			bot.Rooms = rooms
			//log.Printf("Rooms: %#v\n", rooms)
		}
		timeout = time.After(3 * time.Minute)
	}
}

// GetUser returns a User by JID, ID, Name or Email
func (bot *Bot) GetUser(find string) *User {
	if strings.Contains(find, "/") {
		parts := strings.Split(find, "/")
		jid := parts[0]
		resource := parts[1]
		if strings.Contains(jid, "@chat.hipchat.com") {
			find = jid
		} else {
			find = resource
		}
	}
	for _, user := range bot.Users {
		//log.Printf("Hmmmm, %#v\n", user)
		if user.Email == find || user.JID == find || strconv.FormatInt(user.ID, 10) == find || user.Name == find {
			return &user
		}
	}
	return nil
}

// GetUser returns a User by JID, ID or Name
func (bot *Bot) GetRoom(q string) *Room {
	if strings.Contains(q, "@chat.hipchat.com") {
		return nil
	}
	if strings.Contains(q, "/") {
		q = strings.Split(q, "/")[0]
		// assert: strings.Contains(jid, "@conf.hipchat.com")
	}
	for _, room := range bot.Rooms {
		//log.Printf("Hmmmm, %#v\n", room)
		if room.JID == q || strconv.FormatInt(room.ID, 10) == q || room.Name == q {
			return &room
		}
	}
	return nil
}
