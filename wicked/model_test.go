package wicked

import "testing"

func TestMeetingNextSubjectID(t *testing.T) {
	m := &Meeting{}
	s1 := &Subject{ID: "1"}
	s2 := &Subject{ID: "2"}
	sFoo := &Subject{ID: "foo"}

	m.Subjects = []*Subject{s1, s2, sFoo}

	if m.NextSubjectID() != "3" {
		t.Error("NextSubjectId should be '3'")
	}
}
