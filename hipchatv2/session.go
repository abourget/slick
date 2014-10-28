package hipchatv2

import (
	"fmt"
	"net/http"

	"github.com/jmcvetta/napping"
)

type Session struct {
	authToken string
	session   *napping.Session
}

func NewSession(authToken string) *Session {
	napSess := napping.Session{Header: &http.Header{}}
	napSess.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	sess := Session{authToken: authToken, session: &napSess}

	return &sess
}

func (sess *Session) Send(req *napping.Request) (res *napping.Response, err error) {
	res, err = sess.session.Send(req)
	if err != nil {
		content := "[unavailable]"
		if res != nil {
			content = res.RawText()
		}
		return nil, fmt.Errorf("Error while sending request: %s %#v", err, content)
	}

	if res.Status() >= 300 {
		return nil, fmt.Errorf("Error from server status=%s: %s", res.Status(), res.Error)
	}

	return res, nil
}
