package hipchatv2

import "github.com/jmcvetta/napping"

type Room struct {
	ID                int64  `json:"id"`
	JID               string `json:"xmpp_jid"`
	Name              string `json:"name"`
	Topic             string `json:"topic"`
	Privacy           string `json:"privacy"`
	IsArchived        bool   `json:"is_archived"`
	IsGuestAccessible bool   `json:"is_guest_accessible"`
	GuestAccessURL    string `json:"guest_access_url"`
}

func GetRooms(authToken string) (rooms []Room, err error) {
	sess := NewSession(authToken)

	url := "https://api.hipchat.com/v2/room"
	params := napping.Params{"expand": "items"}
	result := struct {
		Items []Room `json:"items"`
	}{}

	req := napping.Request{
		Url:    url,
		Method: "GET",
		Params: &params,
		Result: &result,
	}

	_, err = sess.Send(&req)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
