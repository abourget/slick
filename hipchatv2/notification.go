package hipchatv2

import (
	"fmt"
	"net/http"

	"github.com/jmcvetta/napping"
)

func SendNotification(authToken, room, color, format, msg string, notify bool) error {
	url := fmt.Sprintf("https://api.hipchat.com/v2/room/%s/notification", room)

	sess := napping.Session{}

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
	e := struct {
		Error struct {
			Message string
			Code    string
			Type    string
		}
	}{}
	req := napping.Request{
		Url:     url,
		Method:  "POST",
		Payload: payload,
		Error:   e,
		Header:  &http.Header{},
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	res, err := sess.Send(&req)
	if err != nil {
		return fmt.Errorf("Error while sending request: %s", err)
	}

	if res.Status() >= 300 {
		return fmt.Errorf("Error from server: %s, message: %s", e.Error.Type, e.Error.Message)
	}

	return nil
}
