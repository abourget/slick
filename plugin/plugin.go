package plugin
import (
	"github.com/tkawachi/hipchat"
)

type HandleReply struct {
	To      string
	Message string
}

type Plugin interface {
	Handle(*hipchat.Message) *HandleReply
}
