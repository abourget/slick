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

	"github.com/plotly/plotbot"
	"github.com/plotly/plotbot/internal"
)

type Deployer struct {
	runningJob *DeployJob
	bot        *plotbot.Bot
	env        string
	config     *DeployerConfig
	pubsub     *pubsub.PubSub
	internal   *internal.InternalAPI
}

type DeployerConfig struct {
	RepositoryPath string `json:"repository_path"`
}

func init() {
	plotbot.RegisterPlugin(&Deployer{})
}

func (dep *Deployer) InitChatPlugin(bot *plotbot.Bot) {
	var conf struct {
		Deployer DeployerConfig
	}
	bot.LoadConfig(&conf)

	dep.bot = bot
	dep.pubsub = pubsub.New(100)
	dep.config = &conf.Deployer
	dep.env = os.Getenv("PLOTLY_ENV")

	if dep.env == "" {
		dep.env = "debug"
	}

	dep.loadInternalAPI()

	go dep.pubsubForwardReply()

	bot.ListenFor(&plotbot.Conversation{
		HandlerFunc: dep.ChatHandler,
	})
}

func (dep *Deployer) loadInternalAPI() {
	dep.internal = internal.New(dep.bot.LoadConfig)
}

func (dep *Deployer) ChatConfig() *plotbot.ChatPluginConfig {
	return &plotbot.ChatPluginConfig{
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
	process *os.Process
	params  *DeployParams
	quit    chan bool
	kill    chan bool
	killing bool
}

var deployFormat = regexp.MustCompile(`deploy( ([a-zA-Z0-9_\.-]+))? to ([a-z_-]+)( using ([a-z_-]+))?((,| with)? tags?:? ?(.+))?`)

func (dep *Deployer) ChatHandler(conv *plotbot.Conversation, msg *plotbot.Message) {
	bot := conv.Bot

	// Discard non "mention_name, " prefixed messages
	if !strings.HasPrefix(msg.Body, fmt.Sprintf("%s, ", bot.Config.Mention)) {
		return
	}

	if match := deployFormat.FindStringSubmatch(msg.Body); match != nil {
		if dep.runningJob != nil {
			params := dep.runningJob.params
			conv.Reply(msg, fmt.Sprintf("@%s Deploy currently running: %s", msg.FromUser.MentionName, params))
			return
		} else {
			params := &DeployParams{
				Environment: match[3],
				Branch: match[2],
				Tags: match[8],
				DeploymentBranch: match[5],
				InitiatedBy: msg.FromNick(),
				From: "chat",
				initiatedByChat: msg,
			}
			go dep.handleDeploy(params)
		}
		return

	} else if msg.Contains("cancel deploy") {

		if dep.runningJob == nil {
			conv.Reply(msg, "No deploy running, sorry man..")
		} else {
			if dep.runningJob.killing == true {
				conv.Reply(msg, "deploy: Interrupt signal already sent, waiting to die")
				return
			} else {
				conv.Reply(msg, "deploy: Sending Interrupt signal...")
				dep.runningJob.killing = true
				dep.runningJob.kill <- true
			}
		}
		return
	} else if msg.Contains("in the pipe") {
		url := dep.getCompareUrl("prod", "master")
		mention := msg.FromUser.MentionName
		if url != "" {
			conv.Reply(msg, fmt.Sprintf("@%s in master, waiting to reach prod: %s", mention, url))
		} else {
			conv.Reply(msg, fmt.Sprintf("@%s couldn't get current revision on prod", mention))
		}

	} else if msg.Contains("deploy") {
		conv.Reply(msg, "Usage: plot, [please] deploy <branch-name> to <environment>[ using <deployment-branch>][, tags: <ansible-playbook tags>, ..., ...]")
	}
}

func (dep *Deployer) handleDeploy(params *DeployParams) {
	deploymentBranch := params.ParsedDeploymentBranch()
	if err := dep.pullDeployRepo(deploymentBranch); err != nil {
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

	//
	// Launching deploy
	//

	bot := dep.bot
	bot.Notify("Plotly", "purple", "text", fmt.Sprintf("[deployer] Launching: %s", params), true)
	dep.replyPersonnally(params, bot.WithMood("deploying, my friend", "deploying, yyaaahhhOooOOO!"))

	if params.Environment == "prod" {
		url := dep.getCompareUrl(params.Environment, params.Branch)
		if url != "" {
			dep.pubLine(fmt.Sprintf("[deployer] Compare what is being pushed: %s", url))
		}
	}

	dep.pubLine(fmt.Sprintf("[deployer] Running cmd: %s", strings.Join(cmdArgs, " ")))
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = dep.config.RepositoryPath
	cmd.Env = append(os.Environ(), "ANSIBLE_NOCOLOR=1")

	pty, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	dep.runningJob = &DeployJob{
		process: cmd.Process,
		params:  params,
		quit:    make(chan bool, 2),
		kill:    make(chan bool, 2),
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

func (dep *Deployer) pullDeployRepo(deploymentBranch string) error {
	cmd := exec.Command("git", "fetch")
	cmd.Dir = dep.config.RepositoryPath
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error executing git fetch: %s", err)
	}

	cmd = exec.Command("git", "checkout", fmt.Sprintf("origin/%s", deploymentBranch))
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
	dep.bot.ReplyMention(params.initiatedByChat, msg)
}

func (dep *Deployer) getCompareUrl(env, branch string) string {
	if dep.internal == nil {
		return ""
	}

	currentHead := dep.internal.GetCurrentHead(env)
	if currentHead == "" {
		return ""
	}

	url := fmt.Sprintf("https://github.com/plotly/streambed/compare/%s...%s", currentHead, branch)
	return url
}
