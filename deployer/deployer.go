package deployer

/*
  TODO:
  * hook "git pull" on the "deploy/" repo before doing anything
    * report any error
  * start by having something that runs on prod
  * report on "Plotly" that someone runs deploy
  * report to "Devops" the full log
    * also, to any listening Websocket
*/

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/kr/pty"
	"github.com/tuxychandru/pubsub"

	"github.com/abourget/ahipbot"
)

type Deployer struct {
	runningJob *DeployJob
	bot        *ahipbot.Bot
	env        string
	config     *DeployerConfig
	pubsub     *pubsub.PubSub
}

type DeployerConfig struct {
	RepositoryPath string `json:"repository_path"`
}

func init() {
	ahipbot.RegisterPlugin(func(bot *ahipbot.Bot) ahipbot.Plugin {
		var conf struct {
			Deployer DeployerConfig
		}
		bot.LoadConfig(&conf)

		dep := &Deployer{
			bot:    bot,
			pubsub: pubsub.New(100),
			config: &conf.Deployer,
			env:    os.Getenv("PLOTLY_ENV"),
		}
		if dep.env == "" {
			dep.env = "debug"
		}

		go dep.pubsubForwardReply()

		return dep
	})
}

func (dep *Deployer) Config() *ahipbot.PluginConfig {
	return &ahipbot.PluginConfig{
		OnlyMentions: true,
	}
}

/**
 * Examples:
 *   deploy to stage, branch boo, tags boom, reload-streambed
 *   deploy to stage the branch santa-claus with tags boom, reload-streambed
 *   deploy on prod, branch boo with tags: ahuh, mama, papa
 *   deploy to stage the branch master
 *   deploy prod branch boo  // shortest form
 * or second regexp:
 *   deploy branch boo to stage
 *   deploy santa-claus to stage with tags: kaboom
 */

type DeployJob struct {
	process   *os.Process
	params    *DeployParams
	quit      chan bool
	kill      chan bool
	killing   bool
}

var deployFormat = regexp.MustCompile(`deploy( ([a-zA-Z0-9_\.-]+))? to ([a-z_-]+)((,| with)? tags?:? ?(.+))?`)

func (dep *Deployer) Handle(bot *ahipbot.Bot, msg *ahipbot.BotMessage) {
	// Discard non "mention_name, " prefixed messages
	if !strings.HasPrefix(msg.Body, fmt.Sprintf("%s, ", bot.Config.Mention)) {
		return
	}

	if match := deployFormat.FindStringSubmatch(msg.Body); match != nil {
		if dep.runningJob != nil {
			params := dep.runningJob.params
			bot.Reply(msg, fmt.Sprintf("Deploy currently running: %s", params))
			return
		} else {
			params := &DeployParams{Environment: match[3], Branch: match[2], Tags: match[6], InitiatedBy: msg.FromNick(), From: "chat", initiatedByChat: msg}
			go dep.handleDeploy(params)
		}
		return
	}

	if msg.Contains("cancel deploy") {
		if dep.runningJob == nil {
			bot.Reply(msg, "No deploy running, sorry man..")
		} else {
			if dep.runningJob.killing == true {
				bot.Reply(msg, "deploy: Interrupt signal already sent, waiting to die")
				return
			} else {
				bot.Reply(msg, "deploy: Sending Interrupt signal...")
				dep.runningJob.killing = true
				dep.runningJob.kill <- true
			}
		}
		return
	}
}

func (dep *Deployer) handleDeploy(params *DeployParams) {
	if err := dep.pullDeployRepo(); err != nil {
		dep.pubLine(fmt.Sprintf("[deployer] Unable to pull from deploy/ repo: %s. Aborting.", err))
		return
	} else {
		dep.pubLine(`[deployer] Using latest deploy/ revision`)
	}
	hostsFile := fmt.Sprintf("hosts_%s", params.Environment)
	playbookFile := fmt.Sprintf("playbook_%s.yml", params.Environment)
	tags := params.ParsedTags()
	cmdArgs := []string{"ansible-playbook", "-i", hostsFile, playbookFile, "--tags", tags}

	if params.Environment == "prod" {
		if params.Branch != "" {
			dep.pubLine(fmt.Sprintf(`[deployer] WARN: Branch specified (%s).  Ignoring while pushing to "prod"`, params.Branch))
		}
	} else {
		if params.Branch != "" {
			cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("streambed_pull_revision=origin/%s", params.Branch))
		}
	}

	dep.bot.Notify("Plotly", "purple", "text", fmt.Sprintf("[deployer] Launching: %s", params), true)
	dep.replyPersonnally(params, "deploying my friend")

	dep.pubLine(fmt.Sprintf("[deployer] Running cmd: %s", strings.Join(cmdArgs, " ")))
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = dep.config.RepositoryPath
	cmd.Env = append(os.Environ(), "ANSIBLE_NOCOLOR=1")

	pty, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	dep.runningJob = &DeployJob{
		process:   cmd.Process,
		params:    params,
		quit:      make(chan bool, 2),
		kill:      make(chan bool, 2),
	}

	go dep.manageDeployIo(pty)
	go dep.manageKillProcess(pty)

	if err := cmd.Wait(); err != nil {
		dep.pubLine(fmt.Sprintf("[deployer] terminated with error: %s", err))
		dep.replyPersonnally(params, fmt.Sprintf("your deploy failed: %s", err))
	} else {
		dep.pubLine("[deployer] terminated successfully")
		dep.replyPersonnally(params, "your deploy was successful")
	}

	dep.runningJob.quit <- true
	dep.runningJob = nil
}

func (dep *Deployer) pullDeployRepo() error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = dep.config.RepositoryPath
	return cmd.Run()
}

func (dep *Deployer) pubLine(str string) {
	dep.pubsub.Pub(str, "ansible-line")
}

func (dep *Deployer) manageKillProcess(pty *os.File) {
	runningJob := dep.runningJob
	select {
	case <-runningJob.quit:
		return
	case <-runningJob.kill:
		dep.runningJob.process.Signal(os.Interrupt)
		time.Sleep(3 * time.Second)
		if dep.runningJob != nil {
			dep.runningJob.process.Kill()
		}
	}
}

func (dep *Deployer) pubsubForwardReply() {
	for msg := range dep.pubsub.Sub("ansible-line") {
		line := msg.(string)
		dep.bot.SendToRoom("123823_devops", line)
	}
}

func (dep *Deployer) manageDeployIo(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if dep.runningJob == nil {
			continue
		}
		dep.pubLine(scanner.Text())
	}
}

func (dep *Deployer) replyPersonnally(params *DeployParams, msg string) {
	if params.initiatedByChat == nil {
		return
	}
	fromUser := params.initiatedByChat.FromUser
	dep.bot.Reply(params.initiatedByChat, fmt.Sprintf("@%s %s", fromUser.MentionName, msg))
}
