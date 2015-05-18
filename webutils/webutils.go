package webutils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/abourget/slick"
	"github.com/abourget/slick/hipchatv2"
)

type Utils struct {
	bot *slick.Bot
}

func init() {
	slick.RegisterPlugin(&Utils{})
}

func (utils *Utils) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	utils.bot = bot
	privRouter.HandleFunc("/send_notif", utils.handleNotif)
	privRouter.HandleFunc("/send_message", utils.handleSendMessage)
	privRouter.HandleFunc("/hipchat/rooms", utils.handleGetRooms)
	privRouter.HandleFunc("/hipchat/users", utils.handleGetUsers)
}

func (utils *Utils) handleNotif(w http.ResponseWriter, r *http.Request) {
	hipchatv2.SendNotification(utils.bot.Config.HipchatApiToken, "DevOps", "gray", "text", "Hey that's great!", false)

	http.Error(w, "OK", 200)
}

func (utils *Utils) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Room    string
		Message string
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Couldn't decode message", 500)
		return
	}

	room := utils.bot.GetRoom(data.Room)
	if room == nil {
		http.Error(w, "No such room", 400)
		return
	}

	utils.bot.SendToRoom(room.JID, data.Message)

	http.Error(w, "OK", 200)
}

func (utils *Utils) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := hipchatv2.GetUsers(utils.bot.Config.HipchatApiToken)
	if err != nil {
		webReportError(w, "Error fetching users", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Users []hipchatv2.User `json:"users"`
	}{
		Users: users,
	}

	err = enc.Encode(out)
	if err != nil {
		webReportError(w, "Error encoding JSON", err)
		return
	}
	return
}

func (utils *Utils) handleGetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := hipchatv2.GetRooms(utils.bot.Config.HipchatApiToken)
	if err != nil {
		webReportError(w, "Error fetching rooms", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	out := struct {
		Rooms []hipchatv2.Room `json:"rooms"`
	}{
		Rooms: rooms,
	}

	err = enc.Encode(out)
	if err != nil {
		webReportError(w, "Error encoding JSON", err)
		return
	}
	return
}

func webReportError(w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%s\n\n%s\n", msg, err)))
}
