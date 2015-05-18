package deployer

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

	"github.com/abourget/slick"
	"github.com/abourget/slick/internal"
)

type Deployer struct {
	runningJob *DeployJob
	bot        *slick.Bot
	env        string
	config     *DeployerConfig
	pubsub     *pubsub.PubSub
	internal   *internal.InternalAPI
	lockedBy   string
}

type DeployerConfig struct {
	RepositoryPath string `json:"repository_path"`
	AnnounceRoom string `json:"announce_room"`
	ProgressRoom string `json:"progress_room"`
	DefaultDeploymentBranch string `json:"default_deployment_branch"`
	DefaultStreambedBranch string `json:"default_streambed_branch"`
	AllowedProdBranches []string `json:"allowed_prod_branches"`
}

func init() {
	slick.RegisterPlugin(&Deployer{})
}

func (dep *Deployer) InitChatPlugin(bot *slick.Bot) {
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

	bot.ListenFor(&slick.Conversation{
		HandlerFunc: dep.ChatHandler,
	})
}

func (dep *Deployer) loadInternalAPI() {
	dep.internal = internal.New(dep.bot.LoadConfig)
}

func (dep *Deployer) ChatConfig() *slick.ChatPluginConfig {
	return &slick.ChatPluginConfig{
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

var deployFormat = regexp.MustCompile(`deploy( ([a-zA-Z0-9_\.-]+))? to ([a-z_-]+)( using ([a-zA-Z0-9_\.-]+))?((,| with)? tags?:? ?(.+))?`)

func (dep *Deployer) ChatHandler(conv *slick.Conversation, msg *slick.Message) {
	bot := conv.Bot

	// Discard non "mention_name, " prefixed messages
	if !strings.HasPrefix(msg.Body, fmt.Sprintf("%s, ", bot.Config.Mention)) {
		return
	}

	if match := deployFormat.FindStringSubmatch(msg.Body); match != nil {
		if dep.lockedBy != "" {
			conv.Reply(msg, fmt.Sprintf("Deployment was locked by %s.  Unlock with '%s, unlock deployment' if they're OK with it.", dep.lockedBy, dep.bot.Config.Mention))
			return
		}
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
		url := dep.getCompareUrl("prod", dep.config.DefaultStreambedBranch)
		mention := msg.FromUser.MentionName
		if url != "" {
			conv.Reply(msg, fmt.Sprintf("@%s in %s branch, waiting to reach prod: %s", mention, dep.config.DefaultStreambedBranch, url))
		} else {
			conv.Reply(msg, fmt.Sprintf("@%s couldn't get current revision on prod", mention))
		}
	} else if msg.Contains("unlock deploy") {
		dep.lockedBy = ""
		conv.Reply(msg, fmt.Sprintf("Deployment is now unlocked."))
		bot.Notify(dep.config.AnnounceRoom, "purple", "text", fmt.Sprintf("%s has unlocked deployment", msg.FromUser.Name), true)
	} else if msg.Contains("lock deploy") {
		dep.lockedBy = msg.FromUser.Name
		conv.Reply(msg, fmt.Sprintf("Deployment is now locked.  Unlock with '%s, unlock deployment' ASAP!", dep.bot.Config.Mention))
		bot.Notify(dep.config.AnnounceRoom, "purple", "text", fmt.Sprintf("%s has locked deployment", dep.lockedBy), true)
	} else if msg.Contains("deploy") || msg.Contains("push to") {
		mention := dep.bot.Config.Mention
		conv.Reply(msg, fmt.Sprintf(`Usage: %s, [please|insert reverence] deploy [<branch-name>] to <environment> [using <deployment-branch>][, tags: <ansible-playbook tags>, ..., ...]
examples: %s, please deploy to prod
%s, deploy thing-to-test to stage
%s, deploy complicated-thing to stage, tags: updt_streambed, blow_up_the_sun
other commands: %s, what's in the pipe? - show what's waiting to be deployed to prod
%s, lock deployment - prevent deployment until it's unlocked`, mention, mention, mention, mention, mention, mention))
	}
}

func (dep *Deployer) handleDeploy(params *DeployParams) {
	deploymentBranch := params.ParsedDeploymentBranch(dep.config.DefaultDeploymentBranch)
	if err := dep.pullDeployRepo(deploymentBranch); err != nil {
		errorMsg := fmt.Sprintf("Unable to pull from deployment/ repo: %s. Aborting.", err)
		dep.pubLine(fmt.Sprintf("[deployer] %s", errorMsg))
		dep.replyPersonnally(params, errorMsg)
		return
	} else {
		dep.pubLine(fmt.Sprintf("[deployer] Using %s deployment/ branch (latest revision)", deploymentBranch))
	}
	hostsFile := fmt.Sprintf("hosts_%s", params.Environment)
	if params.Environment == "prod" {
		hostsFile = "tools/plotly_ec2.py"
	}
	playbookFile := fmt.Sprintf("playbook_%s.yml", params.Environment)
	tags := params.ParsedTags()
	cmdArgs := []string{"ansible-playbook", "-i", hostsFile, playbookFile, "--tags", tags}

	if params.Branch != "" {
		if params.Environment == "prod" {
			ok := false
			for _, branch := range dep.config.AllowedProdBranches {
				if branch == params.Branch {
					ok = true
					break
				}
			}
			if !ok {
				errorMsg := fmt.Sprintf("%s is not a legal streambed branch for prod.  Aborting.", params.Branch)
				dep.pubLine(fmt.Sprintf("[deployer] %s", errorMsg))
				dep.replyPersonnally(params, errorMsg)
				return
			}
		}
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("streambed_pull_revision=origin/%s", params.Branch))
	} else {
		params.Branch = dep.config.DefaultStreambedBranch
	}

	//
	// Launching deploy
	//

	bot := dep.bot
	bot.Notify(dep.config.AnnounceRoom, "purple", "text", fmt.Sprintf("[deployer] Launching: %s", params), true)
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
		dep.replyPersonnally(params, bot.WithMood("your deploy was successful", "your deploy was GREAT, you're great !"))
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
		dep.bot.SendToRoom(dep.config.ProgressRoom, line)
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
