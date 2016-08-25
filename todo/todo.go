package todo

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/abourget/slick"
)

func (p *Plugin) listenTodo() {
	p.bot.Listen(&slick.Listener{
		Matches:            regexp.MustCompile(`^!todo.*`),
		MessageHandlerFunc: p.handleTodo,
	})
}

func (p *Plugin) handleTodo(listen *slick.Listener, msg *slick.Message) {

	idFormat := regexp.MustCompile(`^[a-z]{2}$`)
	match := msg.Match
	parts := strings.Split(match[0], " ")
	if len(parts) == 1 {
		p.listTasks(msg, false)
		return
	}
	act := parts[1]
	allCommands := map[string]bool{
		"all":            true,
		"-a":             true,
		"--all":          true,
		":allthethings:": true,
	}

	switch act {
	case "add":
		p.createTask(msg)

	case "close", "fix", "scratch", "done", "strike", "ship", ":boom:", "remove":
		if len(parts) < 3 || !idFormat.MatchString(parts[2]) {
			msg.ReplyMention(fmt.Sprintf("Please %s a task with `!todo %s ID`", act, act))
			return
		}
		if act == "remove" {
			p.deleteTask(msg, parts[2])
		} else {
			p.closeTask(msg, parts[2])
		}

	case "list":
		includeClosed := len(parts) > 2 && allCommands[parts[2]]
		p.listTasks(msg, includeClosed)

	case "help":
		p.replyHelp(msg, "")

	default:
		if idFormat.MatchString(act) {
			p.detailTask(msg, act)
		} else {
			p.replyHelp(msg, "Wooops, not sure what you wanted.\n")
		}
	}
}

func (p *Plugin) detailTask(msg *slick.Message, id string) {
	todo := p.store.Get(msg.Channel)
	index, err := getTaskIndex(id, todo)
	if err != nil {
		msg.ReplyMention("Task not found...")
		return
	}
	task := todo[index]
	msg.Reply(printTaskDetails(task))
}

func printTaskDetails(task *Task) string {
	return "```" + `
ID          ` + task.ID + `
CreatedAt   ` + task.CreatedAt.Format("RFC3339") + `
User        ` + task.User + `
Text        ` + strings.Join(task.Text, " // ") + `
Closed      ` + fmt.Sprintf("%v", task.Closed) + `
ClosingNote ` + task.ClosingNote + `
ClosedAt    ` + task.ClosedAt.Format("RFC3339") + "```"
}

func (p *Plugin) createTask(msg *slick.Message) {
	var text []string
	text = append(text, strings.TrimPrefix(msg.Match[0], "!todo add "))
	todo := p.store.Get(msg.Channel)

	if len(todo) > 600 {
		msg.ReplyMention("Gosh you have over 600 tasks!!! Clean some up first.")
		return
	}

	id := p.generateRandomID(todo)
	task := &Task{
		ID:        id,
		CreatedAt: time.Now(),
		User:      msg.FromUser.ID,
		Text:      text,
		Closed:    false,
	}
	todo = append(todo, task)
	p.store.Put(msg.Channel, todo)
	msg.Reply("`" + task.ID + "` added to the todo")
}

func (p *Plugin) listTasks(msg *slick.Message, includeClosed bool) {
	todo := p.store.Get(msg.Channel)
	var answer []string
	for _, task := range todo {
		if task.Closed && !includeClosed {
			continue
		}
		// TODO format if task is closed
		text := "`" + task.ID + "` " + strings.Join(task.Text, " // ")
		answer = append(answer, text)
	}
	if len(answer) == 0 {
		msg.Reply("Nothing to do... Coffee time?")
	} else {
		msg.Reply(strings.Join(answer, "\n"))
	}
}

func (p *Plugin) closeTask(msg *slick.Message, id string) {
	todo := p.store.Get(msg.Channel)
	index, err := getTaskIndex(id, todo)
	if err != nil {
		msg.ReplyMention("Task not found...")
		return
	}
	parts := strings.Split(msg.Match[0], " ")
	task := todo[index]
	task.Closed = true
	task.ClosedAt = time.Now()
	if len(parts) > 3 {
		task.ClosingNote = strings.Join(parts[3:], " ")
	}
	p.store.Put(msg.Channel, todo)
	msg.Reply("`" + task.ID + "` ~" + strings.Join(task.Text, " // ") + "~ " + task.ClosingNote)
}

func (p *Plugin) deleteTask(msg *slick.Message, id string) {
	todo := p.store.Get(msg.Channel)
	index, err := getTaskIndex(id, todo)
	if err != nil {
		msg.ReplyMention("Task not found...")
		return
	}
	todo = append(todo[:index], todo[index+1:]...)
	p.store.Put(msg.Channel, todo)
	msg.Reply(fmt.Sprintf("Deleted task `%s`", id))
}

func getTaskIndex(id string, todo Todo) (int, error) {
	for i, task := range todo {
		if task.ID == id {
			return i, nil
		}
	}
	return 0, errors.New("Not found")
}

func (p *Plugin) replyHelp(msg *slick.Message, extra string) {
	answer := extra + `
	Here's how you can get things orgnz'ed™:
		!todo add sometin to get done
		!todo list
		!todo strike ID
		!todo remove ID
		!todo help
		!todo ID
	`
	msg.Reply(answer)
	return
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (p *Plugin) generateRandomID(todo Todo) string {
	for {
		id := randSeq(2)
		if idInList(id, todo) {
			continue
		}
		return id
	}
}

func idInList(id string, todo Todo) bool {
	for _, task := range todo {
		if task.ID == id {
			return true
		}
	}
	return false
}
