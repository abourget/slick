package hipchatv2

import (
	"fmt"

	"github.com/jmcvetta/napping"
)

func SetTopic(authToken, room, topic string) (*napping.Request, error) {
	url := fmt.Sprintf("https://api.hipchat.com/v2/room/%s/topic", room)

	sess := NewSession(authToken)

	payload := struct {
		Topic string `json:"topic"`
	}{topic}

	req := napping.Request{
		Url:     url,
		Method:  "PUT",
		Payload: payload,
	}

	_, err := sess.Send(&req)

	return &req, err

}
