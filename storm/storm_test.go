package storm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/abourget/ahipbot/asana"
)

func TestTakerTemplate(t *testing.T) {
	buf := bytes.NewBuffer([]byte(""))
	user := &asana.User{
		Name: "Bob",
		Photo: &asana.Photo{
			Image128: "big image",
		},
	}
	data := struct {
		User      *asana.User
		ForcePush string
	}{
		User:      user,
		ForcePush: "force_push.jpg",
	}
	takerTpl.Execute(buf, data)

	str := buf.String()

	//t.Log(str)

	if !strings.Contains(str, "force_push.jpg") {
		t.Error("Doesn't contain the force_push image")
	}

	if !strings.Contains(str, "big%20image") {
		t.Error("Doesn't contain the 'big%20image'")
	}

	if !strings.Contains(str, "Bob") {
		t.Error("Doesn't contain 'Bob'")
	}

}

func TestTakerTemplateNoPhoto(t *testing.T) {
	buf := bytes.NewBuffer([]byte(""))
	user := &asana.User{
		Name: "Bob",
	}
	data := tplData{
		"User":      user,
		"ForcePush": "force_push.jpg",
	}
	takerTpl.Execute(buf, data)

	//str := buf.String()
	//t.Log(str)
}

func TestSomething(t *testing.T) {
	v := map[string]interface{}{"Something": 123, "SomeOther": "string"}
	t.Log(v)
}
