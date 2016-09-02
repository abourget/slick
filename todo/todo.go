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

	switch act {
	case "add":
		if len(parts) < 2 {
			msg.ReplyMention(fmt.Sprintf("Add a task with `!todo add [some text]`", act))
			return
		}
		p.createTask(msg, strings.Join(parts[2:], " "))

	case "close", "fix", "scratch", "done", "strike", "ship", ":boom:", "remove", "delete":
		if len(parts) < 3 || !idFormat.MatchString(parts[2]) {
			msg.ReplyMention(fmt.Sprintf("Please %s a task with `!todo %s ID`", act, act))
			return
		}
		if act == "remove" || act == "delete" {
			p.deleteTask(msg, parts[2])
		} else {
			p.closeTask(msg, parts[2])
		}

	case "note", "append", "ref":
		if len(parts) < 4 || !idFormat.MatchString(parts[2]) {
			msg.ReplyMention(fmt.Sprintf("Please %s a task with `!todo %s ID [more notes]`", act, act))
			return
		}

		p.appendToTask(msg, parts[2], strings.Join(parts[3:], " "))

	case "all":
		p.listTasks(msg, true)

	case "list":
		p.listTasks(msg, false)

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
	return fmt.Sprintf("%s\n> Created %s by <@%s>", task.String(), task.CreatedAt.Format("2006-01-02 15:04:05"), task.CreatedBy)
}

func (p *Plugin) createTask(msg *slick.Message, content string) {
	todo := p.store.Get(msg.Channel)

	if len(todo) > 600 {
		msg.ReplyMention("Gosh you have over 600 tasks!!! Clean some up first.")
		return
	}

	id := p.generateRandomID(todo)
	task := &Task{
		ID:        id,
		CreatedAt: time.Now(),
		CreatedBy: msg.FromUser.ID,
		Text:      []string{content},
	}
	todo = append(todo, task)
	p.store.Put(msg.Channel, todo)
	msg.ReplyMention("added: " + task.String())
}

func (p *Plugin) appendToTask(msg *slick.Message, id, text string) {
	todo := p.store.Get(msg.Channel)
	index, err := getTaskIndex(id, todo)
	if err != nil {
		msg.ReplyMention("Task not found...")
		return
	}

	task := todo[index]
	task.Text = append(task.Text, strings.Split(text, " // ")...)
	p.store.Put(msg.Channel, todo)

	msg.ReplyMention("updated " + task.String())
}

func (p *Plugin) listTasks(msg *slick.Message, includeClosed bool) {
	todo := p.store.Get(msg.Channel)
	var answer []string
	for _, task := range todo {
		if task.Closed && !includeClosed {
			continue
		}
		answer = append(answer, task.String())
	}
	if len(answer) == 0 {
		msg.ReplyMention("Nothing to do... Coffee time?")
	} else {
		msg.Reply(strings.Join(answer, "\n"))
	}
}

func (p *Plugin) closeTask(msg *slick.Message, ids string) {
	todo := p.store.Get(msg.Channel)

	parts := strings.Split(msg.Match[0], " ")
	var closingNodes string
	if len(parts) > 3 {
		closingNodes = strings.Join(parts[3:], " ")
	}

	var out []string
	for _, id := range strings.Split(ids, ",") {
		index, err := getTaskIndex(id, todo)
		if err != nil {
			out = append(out, "Task `"+id+"` not found")
			continue
		}

		task := todo[index]
		task.Closed = true
		task.ClosedAt = time.Now()
		task.ClosedBy = msg.FromUser.ID
		task.ClosingNote = closingNodes

		out = append(out, task.String())
	}

	p.store.Put(msg.Channel, todo)

	msg.Reply(strings.Join(out, "\n"))
}

func (p *Plugin) deleteTask(msg *slick.Message, ids string) {
	todo := p.store.Get(msg.Channel)

	var out []string
	for _, id := range strings.Split(ids, ",") {
		index, err := getTaskIndex(id, todo)
		if err != nil {
			out = append(out, "Task `"+id+"` not found")
			continue
		}

		todo = append(todo[:index], todo[index+1:]...)
		out = append(out, fmt.Sprintf("Deleted task `%s`", id))
	}

	p.store.Put(msg.Channel, todo)

	msg.Reply(strings.Join(out, "\n"))
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
	answer := extra + `Here's how you can get things orgnz'ed™:` + "```" + `
!todo add [some text]             - add task
!todo                             - list tasks
!todo all                         - list tasks including closed
!todo strike [id] [opt. details]  - mark as done (opt. closing note)
!todo remove [id]                 - delete task(s) (with [id,id,id])
!todo append [id] [more stuff]    - append text to a task
!todo [id]                        - show details
!todo help                        - show this help
` + "```"
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
