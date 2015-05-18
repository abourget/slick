package standup

import "testing"

func TestStandupDataString(t *testing.T) {

	datastr := standupData{
		Yesterday: "a",
		Today:     "b",
		Blocking:  "c",
	}.String()

	str := `Yesterday: a
Today: b
Blocking: c
`

	if datastr != str {
		t.Error("expected '" + datastr + "'" + " to be '" + str + "'")
	}
}
