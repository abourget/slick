package inout

import (
	"github.com/tkawachi/hipchat"
	"github.com/tkawachi/hipbot/plugin"
	"strings"
)

type InOut struct{
	IsIn map[string]bool
}

func New() *InOut {
	inout := new(InOut)
	inout.IsIn = make(map[string]bool)
	return inout
}

func (inout *InOut) Who() string {
	ins := []string{}
	for k, v := range inout.IsIn {
		if v {
			ins = append(ins, k)
		}
	}
	if len(ins) > 0 {
		return strings.Join(ins, ", ") + " are working!"
	} else {
		return "No one is working"
	}
}

func (inout *InOut) Handle(msg *hipchat.Message) *plugin.HandleReply {
	switch msg.Body {
	case "in":
		inout.IsIn[msg.FromNick()] = true
	case "out":
		inout.IsIn[msg.FromNick()] = false
	}
	switch msg.Body {
	case "in", "out", "who":
		return &plugin.HandleReply {
			To: msg.From,
			Message: inout.Who(),
		}
	}
	return nil
}

