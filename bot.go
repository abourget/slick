package slick

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/syndtr/goleveldb/leveldb"
)

type Bot struct {
	// Global bot configuration
	configFile string
	Config     SlackConfig

	// Slack connectivity
	Slack    *slack.Client
	rtm      *slack.RTM
	Users    map[string]slack.User
	Channels map[string]slack.Channel
	Myself   slack.UserDetails

	// Internal handling
	listeners     []*Listener
	addListenerCh chan *Listener
	delListenerCh chan *Listener
	outgoingMsgCh chan *slack.OutgoingMessage

	// Storage
	LevelDBConfig LevelDBConfig
	DB            *leveldb.DB

	// Other features
	WebServer WebServer
	Mood      Mood
}

func New(configFile string) *Bot {
	bot := &Bot{
		configFile:    configFile,
		outgoingMsgCh: make(chan *slack.OutgoingMessage, 100),
		addListenerCh: make(chan *Listener, 100),
		delListenerCh: make(chan *Listener, 100),

		Users:    make(map[string]slack.User),
		Channels: make(map[string]slack.Channel),
	}

	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	return bot
}

func (bot *Bot) Run() {
	bot.loadBaseConfig()

	// Write PID
	err := bot.writePID()
	if err != nil {
		log.Fatal("Couldn't write PID file:", err)
	}

	db, err := leveldb.OpenFile(bot.LevelDBConfig.Path, nil)
	if err != nil {
		log.Fatal("Could not initialize Leveldb key/value store")
	}
	defer func() {
		log.Fatal("Database is closing")
		db.Close()
	}()
	bot.DB = db

	// Init all plugins
	enabledPlugins := make([]string, 0)
	for _, plugin := range registeredPlugins {
		pluginType := reflect.TypeOf(plugin)
		if pluginType.Kind() == reflect.Ptr {
			pluginType = pluginType.Elem()
		}
		typeList := make([]string, 0)
		if _, ok := plugin.(PluginInitializer); ok {
			typeList = append(typeList, "Plugin")
		}
		if _, ok := plugin.(WebServer); ok {
			typeList = append(typeList, "WebServer")
		}
		if _, ok := plugin.(WebServerAuth); ok {
			typeList = append(typeList, "WebServerAuth")
		}
		if _, ok := plugin.(WebPlugin); ok {
			typeList = append(typeList, "WebPlugin")
		}

		log.Printf("Plugin %s implements %s", pluginType.String(),
			strings.Join(typeList, ", "))
		enabledPlugins = append(enabledPlugins, strings.Replace(pluginType.String(), ".", "_", -1))
	}

	initChatPlugins(bot)
	initWebServer(bot, enabledPlugins)
	initWebPlugins(bot)

	if bot.WebServer != nil {
		go bot.WebServer.RunServer()
	}

	bot.Slack = slack.New(bot.Config.ApiToken)
	bot.Slack.SetDebug(bot.Config.Debug)

	rtm := bot.Slack.NewRTM()
	bot.rtm = rtm

	bot.setupHandlers()

	bot.rtm.ManageConnection()
}

func (bot *Bot) writePID() error {
	var serverConf struct {
		Server struct {
			Pidfile string `json:"pid_file"`
		}
	}

	err := bot.LoadConfig(&serverConf)
	if err != nil {
		return err
	}

	if serverConf.Server.Pidfile == "" {
		return nil
	}

	pid := os.Getpid()
	pidb := []byte(strconv.Itoa(pid))
	return ioutil.WriteFile(serverConf.Server.Pidfile, pidb, 0755)
}

// ListenFor registers a listener for messages and events. There are two main
// handling functions on a Listener: MessageHandlerFunc and EventHandlerFunc.
// MessageHandlerFunc is filtered by a bunch of other properties of the Listener,
// whereas EventHandlerFunc will receive all events unfiltered, but with
// *slick.Message instead of a raw *slack.MessageEvent (it's in there anyway),
// which adds a bunch of useful methods to it.
//
// Explore the Listener for more details.
func (bot *Bot) ListenFor(listen *Listener) error {
	listen.Bot = bot

	err := listen.checkParams()
	if err != nil {
		log.Println("Bot.ListenFor(): Invalid Listener: ", err)
		return err
	}

	listen.setupChannels()

	if listen.isManaged() {
		go listen.launchManager()
	}

	bot.addListenerCh <- listen

	return nil
}

func (bot *Bot) Notify(room, color, format, msg string, notify bool) error {
	log.Println("DEPRECATED. Please use the Slack API with .PostMessage")
	// bot.api.PostMessage(room, msg, slack.PostMessageParameters{
	// 	Attachments: []slack.Attachment{
	// 		{
	// 			Color: color,
	// 			Text: msg,
	// 		},
	// 	},
	// })
	return nil
}

func (bot *Bot) SendToChannel(channelName string, message string) {
	channel := bot.GetChannelByName(channelName)

	if channel == nil {
		log.Printf("Couldn't send message, channel %q not found: %q\n", channelName, message)
		return
	}
	log.Printf("Sending to channel %q: %q\n", channelName, message)

	bot.SendOutgoingMessage(message, channel.ID)

	return
}

func (bot *Bot) setupHandlers() {
	go bot.replyHandler()
	go bot.messageHandler()
	log.Println("Bot ready")
}

func (bot *Bot) cacheUsers(users []slack.User) {
	bot.Users = make(map[string]slack.User)
	for _, user := range users {
		bot.Users[user.ID] = user
	}
}

func (bot *Bot) cacheChannels(channels []slack.Channel, groups []slack.Group) {
	bot.Channels = make(map[string]slack.Channel)
	for _, channel := range channels {
		bot.Channels[channel.ID] = channel
	}

	for _, group := range groups {
		c := slack.Channel{}
		c.ID = group.ID
		c.Name = group.Name
		c.IsChannel = false
		c.Creator = group.Creator
		c.IsArchived = group.IsArchived
		c.IsGeneral = group.IsGeneral
		c.IsGroup = true
		c.Members = group.Members
		c.Topic = group.Topic
		c.Purpose = group.Purpose
		c.IsMember = true
		c.NumMembers = group.NumMembers
		bot.Channels[group.ID] = c
	}
}

func (bot *Bot) loadBaseConfig() {
	if err := checkPermission(bot.configFile); err != nil {
		log.Fatal("ERROR Checking Permissions: ", err)
	}

	var config1 struct {
		Slack SlackConfig
	}
	err := bot.LoadConfig(&config1)
	if err != nil {
		log.Fatalln("Error loading Slack config section:", err)
	} else {
		bot.Config = config1.Slack
	}

	var config2 struct {
		LevelDB LevelDBConfig
	}
	err = bot.LoadConfig(&config2)
	if err != nil {
		log.Fatalln("Error loading LevelDB config section:", err)
	} else {
		bot.LevelDBConfig = config2.LevelDB
	}
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
		outMsg := <-bot.outgoingMsgCh
		if outMsg == nil {
			continue
		}

		bot.rtm.SendMessage(outMsg)

		time.Sleep(50 * time.Millisecond)
	}
}

// SendOutgoingMessage returns a *slack.OutgoingMessage and schedules it for departure.
func (bot *Bot) SendOutgoingMessage(text string, to string) *slack.OutgoingMessage {
	outMsg := bot.rtm.NewOutgoingMessage(text, to)
	bot.outgoingMsgCh <- outMsg
	return outMsg
}

func (bot *Bot) removeListener(listen *Listener) {
	for i, element := range bot.listeners {
		if element == listen {
			// following: https://code.google.com/p/go-wiki/wiki/SliceTricks
			copy(bot.listeners[i:], bot.listeners[i+1:])
			bot.listeners[len(bot.listeners)-1] = nil
			bot.listeners = bot.listeners[:len(bot.listeners)-1]
			return
		}
	}

	return
}

func (bot *Bot) messageHandler() {
	for {
	nextMessages:
		select {
		case listen := <-bot.addListenerCh:
			bot.listeners = append(bot.listeners, listen)

		case listen := <-bot.delListenerCh:
			bot.removeListener(listen)

		case event := <-bot.rtm.IncomingEvents:
			bot.handleRTMEvent(&event)
		}

		// Always flush listeners deletions between messages, so a
		// Close()'d Listener never processes another message.
		for {
			select {
			case listen := <-bot.delListenerCh:
				bot.removeListener(listen)
			default:
				goto nextMessages
			}
		}
	}
}

func (bot *Bot) handleRTMEvent(event *slack.RTMEvent) {
	var msg *Message

	switch ev := event.Data.(type) {
	/**
	 * Connection handling...
	 */
	case *slack.LatencyReport:
		log.Printf("Current latency: %v\n", ev)
	case *slack.RTMError:
		log.Printf("Error: %d - %s\n", ev.Code, ev.Msg)

	case *slack.ConnectedEvent:
		log.Printf("Bot connected, connection_count=%d\n", ev.ConnectionCount)
		bot.Myself = *ev.Info.User
		bot.cacheUsers(ev.Info.Users)
		bot.cacheChannels(ev.Info.Channels, ev.Info.Groups)

		for _, channelName := range bot.Config.JoinChannels {
			channel := bot.GetChannelByName(channelName)
			if channel != nil && !channel.IsMember {
				bot.Slack.JoinChannel(channel.ID)
			}
		}

	case *slack.DisconnectedEvent:
		log.Println("Bot disconnected")

	case *slack.ConnectingEvent:
		log.Printf("Bot connecting, connection_count=%d, attempt=%d\n", ev.ConnectionCount, ev.Attempt)

	case *slack.HelloEvent:
		fmt.Println("Got a HELLO from websocket")

	/**
	 * Message dispatch and handling
	 */
	case *slack.MessageEvent:
		fmt.Printf("Message: %v\n", ev)
		msg = &Message{
			Msg:           &ev.Msg,
			SubMessage:    ev.SubMessage,
			bot:           bot,
		}

		user, ok := bot.Users[ev.User]
		if ok {
			msg.FromUser = &user
		}
		channel, ok := bot.Channels[ev.Channel]
		if ok {
			msg.FromChannel = &channel
		}

		msg.applyMentionsMe(bot)
		msg.applyFromMe(bot)

		log.Printf("Incoming message: %s\n", msg)

	case *slack.PresenceChangeEvent:
		user := bot.Users[ev.User]
		log.Printf("User %q is now %q\n", user.Name, ev.Presence)
		user.Presence = ev.Presence

	// TODO: manage im_open, im_close, and im_created ?

	/**
	 * User changes
	 */
	case *slack.UserChangeEvent:
		bot.Users[ev.User.ID] = ev.User

	/**
	 * Handle channel changes
	 */
	case *slack.ChannelRenameEvent:
		channel := bot.Channels[ev.Channel.ID]
		channel.Name = ev.Channel.Name

	case *slack.ChannelJoinedEvent:
		bot.Channels[ev.Channel.ID] = ev.Channel

	case *slack.ChannelCreatedEvent:
		c := slack.Channel{}
		c.ID = ev.Channel.ID
		c.Name = ev.Channel.Name
		c.Creator = ev.Channel.Creator
		bot.Channels[ev.Channel.ID] = c
		// NICE: poll the API to get a full Channel object ? many
		// things are missing here

	case *slack.ChannelDeletedEvent:
		delete(bot.Channels, ev.Channel)

	case *slack.ChannelArchiveEvent:
		channel := bot.Channels[ev.Channel]
		channel.IsArchived = true

	case *slack.ChannelUnarchiveEvent:
		channel := bot.Channels[ev.Channel]
		channel.IsArchived = false

	/**
	 * Handle group changes
	 */
	case *slack.GroupRenameEvent:
		group := bot.Channels[ev.Group.ID]
		group.Name = ev.Group.Name

	case *slack.GroupJoinedEvent:
		bot.Channels[ev.Channel.ID] = ev.Channel

	case *slack.GroupCreatedEvent:
		c := slack.Channel{}
		c.ID = ev.Channel.ID
		c.Name = ev.Channel.Name
		c.Creator = ev.Channel.Creator
		bot.Channels[ev.Channel.ID] = c

		// NICE: poll the API to get a full Group object ? many
		// things are missing here

	case *slack.GroupCloseEvent:
		// TODO: when a group is "closed"... does that mean removed ?
		// TODO: how do we even manage groups ?!?!
		delete(bot.Channels, ev.Channel)

	case *slack.GroupArchiveEvent:
		group := bot.Channels[ev.Channel]
		group.IsArchived = true

	case *slack.GroupUnarchiveEvent:
		group := bot.Channels[ev.Channel]
		group.IsArchived = false

	default:
		log.Printf("Unexpected: %#v\n", ev)
	}

	// Dispatch listeners
	for _, listen := range bot.listeners {
		if msg == nil {
			if listen.EventHandlerFunc != nil {
				listen.EventHandlerFunc(listen, event.Data)
			}
		} else {
			if listen.MessageHandlerFunc != nil {
				if defaultFilterFunc(listen, msg) {
					listen.MessageHandlerFunc(listen, msg)
				}
			} else {
				listen.EventHandlerFunc(listen, msg)
			}
		}
	}

}

// Disconnect the websocket.
func (bot *Bot) Disconnect() {
	// FIXME: implement a Reconnect() method.. calling the RTM method of the same name.
	// QUERYME: do we need that, really ?
	bot.rtm.Disconnect()
}

// GetUser returns a *slack.User by ID, Name, RealName or Email
func (bot *Bot) GetUser(find string) *slack.User {
	for _, user := range bot.Users {
		//log.Printf("Hmmmm, %#v\n", user)
		if user.Profile.Email == find || user.ID == find || user.Name == find || user.RealName == find {
			return &user
		}
	}
	return nil
}

// GetChannelByName returns a *slack.Channel by Name
func (bot *Bot) GetChannelByName(name string) *slack.Channel {
	name = strings.TrimLeft(name, "#")
	for _, channel := range bot.Channels {
		if channel.Name == name {
			return &channel
		}
	}
	return nil
}
