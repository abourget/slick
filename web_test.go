package ahipbot

import (
	"testing"
)

func TestConfigureWebapp(t *testing.T) {
	conf := &WebappConfig{ClientID: "boo", ClientSecret: "mama", SessionAuthKey: "123", SessionEncryptKey: "456"}
	configureWebapp(conf)
}
