package plotbot

import (
	"testing"
	"time"
)

func TestConversationCheckParams(t *testing.T) {
	c := Conversation{
		ListenUntil:    time.Now(),
		ListenDuration: 120 * time.Second,
	}
	err := c.checkParams()
	if err == nil {
		t.Error("checkParams shouldn't be nil")
	}
}

func TestDefeaultFilter(t *testing.T) {
}
