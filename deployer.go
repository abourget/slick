package main

import (
	"bufio"
	"fmt"
	"github.com/kr/pty"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Deployer struct {
	config     *PluginConfig
	runningJob *DeployJob
}

type DeployJob struct {
	process *os.Process
	params  *DeployParams
	quit    chan bool
	kill    chan bool
	killing bool
}

func NewDeployer(bot *Hipbot) *Deployer {
	dep := new(Deployer)

	dep.config = &PluginConfig{
		EchoMessages: false,
		OnlyMentions: true,
	}

	return dep
}

func (dep *Deployer) Config() *PluginConfig {
	return dep.config
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

var deployFormat = regexp.MustCompile(`deploy( ([a-zA-Z0-9_\.-]+))? to ([a-z_-]+)((,| with)? tags?:? ?(.+))?`)

func (dep *Deployer) Handle(bot *Hipbot, msg *BotMessage) {
	if match := deployFormat.FindStringSubmatch(msg.Body); match != nil {
		if dep.runningJob != nil {
			params := dep.runningJob.params
			bot.Reply(msg, fmt.Sprintf("Deploy currently running, initated by %s: env=%s branch=%s tags=%s", params.initiatedBy, params.environment, params.branch, params.Tags()))
			return
		} else {
			params := &DeployParams{environment: match[3], branch: match[2], tags: match[6], initiatedBy: msg.FromNick()}
			dep.handleDeploy(bot, msg, params)
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

type DeployParams struct {
	environment string
	branch      string
	tags        string
	initiatedBy string
}

func (p *DeployParams) Tags() string {
	return strings.Replace(p.tags, " ", "", -1)
}

func (dep *Deployer) handleDeploy(bot *Hipbot, msg *BotMessage, params *DeployParams) {
	bot.Reply(msg, fmt.Sprintf("[process] Running deploy env=%s, branch=%s, tags=%s", params.environment, params.branch, params.Tags()))

	cmdArgs := []string{"ansible-playbook", "-i", "hosts_vagrant", "playbook_vagrant.yml"}
	tags := params.Tags()
	if tags != "" {
		cmdArgs = append(cmdArgs, "--tags", tags)
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = "/home/abourget/plotly/deploy"
	cmd.Env = append(os.Environ(), "ANSIBLE_NOCOLOR=1")

	pty, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	dep.runningJob = &DeployJob{process: cmd.Process, params: params, quit: make(chan bool, 2), kill: make(chan bool, 2)}

	go manageDeployIo(bot, msg, pty)
	go dep.manageKillProcess(pty)

	if err := cmd.Wait(); err != nil {
		bot.Reply(msg, fmt.Sprintf("[process] terminated: %s", err))
	} else {
		bot.Reply(msg, "[process] terminated without error")
	}

	dep.runningJob.quit <- true
	dep.runningJob = nil
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

func manageDeployIo(bot *Hipbot, msg *BotMessage, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		bot.Reply(msg, fmt.Sprintf("%s", scanner.Text()))
	}
}
