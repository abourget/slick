package toxin

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/plotly/plotbot"
)

type Meeting struct {
	ID             string
	CreatedBy      *User
	Room           string
	Goal           string
	CurrentSubject *Subject
	Subjects       []*Subject
	StartTime      time.Time
	EndTime        time.Time
	Logs           []*Message
	Actions        []*Action
	Refs           []*Reference
	Participants   []*User
}

func NewMeeting(id string, user *plotbot.User, goal, room string) *Meeting {
	meeting := &Meeting{}
	meeting.ID = id
	meeting.Room = room
	meeting.Goal = goal
	meeting.StartTime = time.Now()

	newUser := meeting.ImportUser(user)
	meeting.CreatedBy = newUser

	return meeting
}

func (meeting *Meeting) ImportUser(user *plotbot.User) *User {
	fromEmail := user.Email

	for _, user := range meeting.Participants {
		if user.Email == fromEmail {
			return user
		}
	}

	newUser := &User{
		Email:    user.Email,
		Fullname: user.Name,
		PhotoURL: user.PhotoURL,
	}

	meeting.Participants = append(meeting.Participants, newUser)

	return newUser
}

var addActionMatcher = regexp.MustCompile(`(?mi)(a#([a-z]+) )?(.*)`)

func (meeting *Meeting) AddAction(user *User, subject *Subject, text string) *Action {
	// Analyze beginning of text, to see if there's a tag, and take the time
	match := addActionMatcher.FindStringSubmatch(text)
	if match == nil {
		return nil
	}

	id := ""
	if len(match[2]) > 0 {
		id = match[2]
	} else {
		id = meeting.NextActionID()
	}

	action := &Action{
		ID:      id,
		AddedBy: user,
		Text:    match[3],
	}

	meeting.Actions = append(meeting.Actions, action)
	subject.Actions = append(subject.Actions, action)

	return action
}

func (meeting *Meeting) GetActionByID(id string) *Action {
	for _, action := range meeting.Actions {
		if action.ID == id {
			return action
		}
	}
	return nil
}

func (meeting *Meeting) GetSubjectByID(id string) *Subject {
	for _, subject := range meeting.Subjects {
		if subject.ID == id {
			return subject
		}
	}
	return nil
}

var addSubjectMatcher = regexp.MustCompile(`(?mi)(s#([a-z]+) )?(\d+m)\s+(.*)`)

func (meeting *Meeting) AddSubject(user *User, text string) *Subject {
	// Analyze beginning of text, to see if there's a tag, and take the time
	match := addSubjectMatcher.FindStringSubmatch(text)
	if match == nil {
		return nil
	}

	id := ""
	if len(match[2]) > 0 {
		id = match[2]
	} else {
		id = meeting.NextSubjectID()
	}

	duration, _ := time.ParseDuration(match[3])

	subject := &Subject{
		ID:        id,
		AddedBy:   user,
		Text:      match[4],
		TimeLimit: duration,
	}

	meeting.Subjects = append(meeting.Subjects, subject)

	return subject

}

func (meeting *Meeting) CurrentIsLast() bool {
	return meeting.CurrentSubject == meeting.Subjects[len(meeting.Subjects)-1]
}

// NextSubject is called when switching subject.  Do not call to start.
func (meeting *Meeting) NextSubject() *Subject {
	prevSubject := meeting.CurrentSubject
	if prevSubject != nil {
		// Wrap up counters
		prevSubject.Stop()
	} else {
		meeting.CurrentSubject = meeting.Subjects[0]
		return meeting.CurrentSubject
	}

	getNext := false
	for _, subject := range meeting.Subjects {
		if getNext {
			meeting.CurrentSubject = subject
			return subject
		}
		if prevSubject == subject {
			getNext = true
		}
	}

	// That shouldn't happen, provided there *is* a next subject.
	// You should call CurrentIsLast() prior to calling this.
	return nil
}

func (meeting *Meeting) AddReference(user *User, text string) *Reference {
	ref := &Reference{
		AddedBy: user,
	}
	if strings.HasPrefix(text, "http") {
		chunks := strings.SplitN(text, " ", 2)
		ref.URL = chunks[0]
		ref.Text = chunks[1]
	} else {
		ref.Text = text
	}

	meeting.Refs = append(meeting.Refs, ref)

	return ref
}

func (meeting *Meeting) NextSubjectID() string {
	for i := 1; i < 1000; i++ {
		strID := fmt.Sprintf("%d", i)
		taken := false
		for _, subject := range meeting.Subjects {
			if subject.ID == strID {
				taken = true
				break
			}
		}
		if !taken {
			return strID
		}
	}
	return "fail"
}

func (meeting *Meeting) NextActionID() string {
	for i := 1; i < 1000; i++ {
		strID := fmt.Sprintf("%d", i)
		taken := false
		for _, action := range meeting.Actions {
			if action.ID == strID {
				taken = true
				break
			}
		}
		if !taken {
			return strID
		}
	}
	return "fail"
}

func (meeting *Meeting) Conclude() {
	meeting.EndTime = time.Now()
	if meeting.CurrentSubject != nil {
		meeting.CurrentSubject.Stop()
	}
}

type Message struct {
	From      *User
	Timestamp time.Time
	Text      string
}

type Subject struct {
	ID        string
	AddedBy   *User
	Text      string
	TimeLimit time.Duration

	// Global timing
	Duration  time.Duration
	BeginTime time.Time
	FinalTime time.Time
	// Local timing, when we talk about the subject
	StartTime time.Time
	EndTime   time.Time

	Actions     []*Action
	Plusplus    []*Plusplus
	Refs        []*Reference
	WasExtended bool
}

func (subject *Subject) Timebox() string {
	if subject.TimeLimit == 0 {
		return "[no limit]"
	} else {
		return subject.TimeLimit.String()
	}
}

func (subject *Subject) RecordPlusplus(user *User) {
	pp := NewPlusplus(user)
	subject.Plusplus = append(subject.Plusplus, pp)
}

func (subject *Subject) Stop() {
	subject.EndTime = time.Now()
	subject.Duration += subject.EndTime.Sub(subject.StartTime)
	subject.FinalTime = time.Now()
}
func (subject *Subject) Start() {
	if subject.BeginTime.IsZero() {
		subject.BeginTime = time.Now()
	}
	subject.StartTime = time.Now()
}

type Plusplus struct {
	From *User
}

func NewPlusplus(from *User) *Plusplus {
	pp := &Plusplus{
		From: from,
	}
	return pp
}

type Reference struct {
	AddedBy *User
	URL     string
	Text    string
}

type Action struct {
	ID       string
	AddedBy  *User
	Text     string
	Plusplus []*Plusplus
}

func (action *Action) RecordPlusplus(user *User) {
	pp := NewPlusplus(user)
	action.Plusplus = append(action.Plusplus, pp)
}

type User struct {
	Email    string
	Fullname string
	PhotoURL string
}
