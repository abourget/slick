package hipchatv2

import "github.com/jmcvetta/napping"

type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	MentionName string `json:"mention_name"`
	Title       string `json:"title"`
	PhotoURL    string `json:"photo_url"`
	JID         string `json:"xmpp_jid"`
	//LastActive  time.Time `json:"last_active"`
}

func GetUsers(authToken string) (users []User, err error) {
	sess := NewSession(authToken)

	url := "https://api.hipchat.com/v2/user"
	params := napping.Params{"expand": "items"}
	result := struct {
		Items []User `json:"items"`
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
