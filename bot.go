package plotbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jmcvetta/napping"
	"github.com/plotly/hipchat"
	"github.com/plotly/plotbot/hipchatv2"
)

const (
	ConfDomain = "conf.hipchat.com"
)

type Bot struct {
	configFile string
	Config     HipchatConfig

	// Hipchat XMPP client
	client *hipchat.Client

	// Conversations handling
	conversations     []*Conversation
	addConversationCh chan *Conversation
	delConversationCh chan *Conversation

	// replySink sends messages to Hipchat rooms or users
	disconnected chan bool
	replySink    chan *BotReply
	Users        []User
	Rooms        []Room

	redisConfig RedisConfig
	// RedisPool holds a connection to Redis.  NOTE: Prefix all your keys with "plotbot:" please.
	RedisPool *redis.Pool

	// Features from the heart
	WebServer WebServer
	Rewarder  Rewarder
	Mood      Mood
}

func NewHipbot(configFile string) *Bot {
	bot := &Bot{
		configFile:        configFile,
		replySink:         make(chan *BotReply, 10),
		addConversationCh: make(chan *Conversation, 100),
		delConversationCh: make(chan *Conversation, 100),
	}

	return bot
}

func (bot *Bot) Run() {
	bot.loadBaseConfig()
	bot.setupStorage()

	// Init all plugins
	enabledPlugins := make([]string, 0)
	for _, plugin := range registeredPlugins {
		pluginType := reflect.TypeOf(plugin)
		if pluginType.Kind() == reflect.Ptr {
			pluginType = pluginType.Elem()
		}
		typeList := make([]string, 0)
		if _, ok := plugin.(Rewarder); ok {
			typeList = append(typeList, "Rewarder")
		}
		if _, ok := plugin.(ChatPlugin); ok {
			typeList = append(typeList, "ChatPlugin")
		}
		if _, ok := plugin.(WebServer); ok {
			typeList = append(typeList, "WebServer")
		}
		if _, ok := plugin.(WebPlugin); ok {
			typeList = append(typeList, "WebPlugin")
		}

		log.Printf("Plugin %s implements %s", pluginType.String(),
			strings.Join(typeList, ", "))
		enabledPlugins = append(enabledPlugins, strings.Replace(pluginType.String(), ".", "_", -1))
	}

	InitRewarder(bot)
	InitChatPlugins(bot)
	InitWebServer(bot, enabledPlugins)
	InitWebPlugins(bot)

	if bot.WebServer != nil {
		go bot.WebServer.ServeWebRequests()
	}

	//time.Sleep(5000 * time.Second)
	//return

	for {
		log.Println("Connecting client...")
		err := bot.connectClient()
		if err != nil {
			log.Println("  `- Failed: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		bot.setupHandlers()

		select {
		case <-bot.disconnected:
			log.Println("Disconnected...")
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (bot *Bot) ListenFor(conv *Conversation) error {
	conv.Bot = bot

	err := conv.checkParams()
	if err != nil {
		log.Println("Bot.ListenFor(): Invalid Conversation: ", err)
		return err
	}

	conv.setupChannels()

	if conv.isManaged() {
		go conv.launchManager()
	}

	bot.addConversationCh <- conv

	return nil
}

func (bot *Bot) Reply(msg *Message, reply string) {
	log.Println("Replying:", reply)
	bot.replySink <- msg.Reply(reply)
}

// ReplyMention replies with a @mention named prefixed, when replying in public. When replying in private, nothing is added.
func (bot *Bot) ReplyMention(msg *Message, reply string) {
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

func (bot *Bot) ReplyPrivately(msg *Message, reply string) {
	log.Println("Replying privately:", reply)
	bot.replySink <- msg.ReplyPrivately(reply)
}

func (bot *Bot) Notify(room, color, format, msg string, notify bool) (*napping.Request, error) {
	log.Println("Notifying room ", room, ": ", msg)
	return hipchatv2.SendNotification(bot.Config.HipchatApiToken, bot.GetRoomId(room), color, format, msg, notify)
}

func (bot *Bot) SendToRoom(room string, message string) {
	log.Println("Sending to room ", room, ": ", message)

	room = CanonicalRoom(room)

	reply := &BotReply{
		To:      room,
		Message: message,
	}
	bot.replySink <- reply
	return
}

// GetRoomdId returns the numeric room ID as string for a given XMPP room JID
func (bot *Bot) GetRoomId(room string) string {
	roomName := CanonicalRoom(room)
	for _, room := range bot.Rooms {
		if roomName == room.JID {
			return fmt.Sprintf("%v", room.ID)
		}
	}
	return room
}

func (bot *Bot) connectClient() (err error) {
	resource := bot.Config.Resource
	if resource == "" {
		resource = "bot"
	}

	bot.client, err = hipchat.NewClient(
		bot.Config.Username, bot.Config.Password, resource)
	if err != nil {
		log.Println("Error in connectClient(): ", err)
		return
	}

	for _, room := range bot.Config.Rooms {
		bot.client.Join(CanonicalRoom(room), bot.Config.Nickname)
	}

	return
}

func (bot *Bot) setupHandlers() {
	bot.client.Status("chat")
	bot.disconnected = make(chan bool)
	go bot.client.KeepAlive()
	go bot.replyHandler()
	go bot.messageHandler()
	go bot.usersPolling()
	go bot.roomsPolling()
	log.Println("Bot ready")
}

func (bot *Bot) loadBaseConfig() {
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

func (bot *Bot) setupStorage() error {
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

func (bot *Bot) removeConversation(conv *Conversation) {
	for i, element := range bot.conversations {
		if element == conv {
			// following: https://code.google.com/p/go-wiki/wiki/SliceTricks
			copy(bot.conversations[i:], bot.conversations[i+1:])
			bot.conversations[len(bot.conversations)-1] = nil
			bot.conversations = bot.conversations[:len(bot.conversations)-1]
			return
		}
	}

	return
	bot.conversations = append(bot.conversations, conv)

}

func (bot *Bot) messageHandler() {
	msgs := bot.client.Messages()
	for {
		select {
		case <-bot.disconnected:
			return

		case conv := <-bot.addConversationCh:
			bot.conversations = append(bot.conversations, conv)

		case conv := <-bot.delConversationCh:
			bot.removeConversation(conv)

		case rawMsg := <-msgs:
			fromUser := bot.GetUser(rawMsg.From)

			if fromUser == nil {
				log.Printf("Message from unknown user, skipping: %#v\n", rawMsg)
				continue
			}

			msg := &Message{Message: rawMsg}
			msg.FromUser = fromUser
			msg.FromRoom = bot.GetRoom(rawMsg.From)
			msg.applyMentionsMe(bot)
			msg.applyFromMe(bot)

			log.Printf("Incoming message: %s\n", msg)

			for _, conv := range bot.conversations {
				filterFunc := defaultFilterFunc
				if conv.FilterFunc != nil {
					filterFunc = conv.FilterFunc
				}

				if filterFunc(conv, msg) {
					conv.HandlerFunc(conv, msg)
				}
			}
		}

		// Always flush conversations deletions between messages, so a
		// Close()'d Conversation never processes another message.
		select {
		case conv := <-bot.delConversationCh:
			bot.removeConversation(conv)
		default:
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
			hcUsers, err := hipchatv2.GetUsers(bot.Config.HipchatApiToken)

			if err != nil {
				log.Printf("GetUsers error: %s", err)
				timeout = time.After(5 * time.Second)
				continue
			}

			users := []User{}
			for _, user := range hcUsers {
				users = append(users, UserFromHipchatv2(user))
			}
			bot.Users = users
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
			hcRooms, err := hipchatv2.GetRooms(bot.Config.HipchatApiToken)

			if err != nil {
				log.Printf("GetRooms error: %s", err)
				timeout = time.After(5 * time.Second)
				continue
			}

			rooms := []Room{}
			for _, room := range hcRooms {
				rooms = append(rooms, RoomFromHipchatv2(room))
			}

			bot.Rooms = rooms
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
