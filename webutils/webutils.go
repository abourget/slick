package webutils

import (
	"fmt"
	"net/http"

	"github.com/abourget/ahipbot"
	"github.com/abourget/ahipbot/hipchatv2"
	"github.com/gorilla/mux"
)

type Utils struct {
	bot *ahipbot.Bot
}

func (utils *Utils) Config() *ahipbot.PluginConfig {
	return nil
}
func (utils *Utils) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		return &Utils{bot: bot}
	})
}

func (utils *Utils) WebPluginSetup(router *mux.Router) {
	router.HandleFunc("/send_notif", utils.handleNotif)
}

func (utils *Utils) handleNotif(w http.ResponseWriter, r *http.Request) {
	hipchatv2.SendNotification(utils.bot.Config.HipchatApiToken, "DevOps", "gray", "text", "Hey that's great!", false)

	http.Error(w, "OK", 200)
}

// func handleGetUsers(w http.ResponseWriter, r *http.Request) {
// 	users, err := hipchatv2.GetUsers(bot.Config.HipchatApiToken)
// 	if err != nil {
// 		webReportError(w, "Error fetching users", err)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	enc := json.NewEncoder(w)
// 	out := struct {
// 		Users []hipchatv2.User `json:"users"`
// 	}{
// 		Users: users,
// 	}

// 	err = enc.Encode(out)
// 	if err != nil {
// 		webReportError(w, "Error encoding JSON", err)
// 		return
// 	}
// 	return
// }

// func handleGetRooms(w http.ResponseWriter, r *http.Request) {
// 	rooms, err := hipchatv2.GetRooms(bot.Config.HipchatApiToken)
// 	if err != nil {
// 		webReportError(w, "Error fetching rooms", err)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	enc := json.NewEncoder(w)
// 	out := struct {
// 		Rooms []hipchatv2.Room `json:"rooms"`
// 	}{
// 		Rooms: rooms,
// 	}

// 	err = enc.Encode(out)
// 	if err != nil {
// 		webReportError(w, "Error encoding JSON", err)
// 		return
// 	}
// 	return
// }

func webReportError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%s\n\n%s\n", msg, err)))
}
