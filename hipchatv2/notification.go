package hipchatv2

import (
	"fmt"
	"github.com/jmcvetta/napping"
)

func SendNotification(authToken, room, color, format, msg string, notify bool) error {
	url := fmt.Sprintf("https://api.hipchat.com/v2/room/%s/notification", room)
	
	sess := NewSession(authToken)

	payload := struct {
		Color  string `json:"color"`          // yellow, green, red, purple, gray
		Format string `json:"message_format"` // "html" or "text"
		Notify bool   `json:"notify"`
		Msg    string `json:"message"`
	}{
		Color:  color,
		Format: format,
		Msg:    msg,
		Notify: notify,
	}
	e := ApiError{}

	req := napping.Request{
		Url:     url,
		Method:  "POST",
		Payload: payload,
		Error:   e,


	}

	_, err := sess.Send(&req)

	return err
}
