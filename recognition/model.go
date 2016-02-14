package recognition

import "time"

type Recognition struct {
	MsgTimestamp string // Slack Msg TS
	CreatedAt    time.Time
	FulfilledAt  time.Time
	Sender       string         // Slack User ID
	Recipients   []string       // Slack User IDs
	Reactions    map[string]int // [slackUID] = count of reactions
	Categories   []string       // ["1.4", "4.5"]
}
