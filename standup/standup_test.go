package standup

import (
	"fmt"
	"testing"
)

func TestStandup(t *testing.T) {
	su := &Standup{}
	fmt.Println("Standup:", su)
}

func TestRegexpMatch(t *testing.T) {
	input := `!blocking this is good
!yesterday thank you
`
	match := sectionRegexp.FindAllStringSubmatchIndex(input, -1)
	res := extractSectionAndText(input, match)

	if res[0].name != "blocking" {
		t.Error("res[0].name should be blocking")
		t.Error("boo", res)
	}
	if res[1].name != "yesterday" {
		t.Error("res[1].name should be yesterday")
	}
	if res[1].text != "thank you" {
		t.Error("res[1].text should be 'thank you'")
	}
}
