package deployer

import (
	"fmt"
	"strings"

	"github.com/abourget/ahipbot"
)

type DeployParams struct {
	Environment     string
	Branch          string
	Tags            string
	InitiatedBy     string
	From            string
	initiatedByChat *ahipbot.BotMessage
}

func (p *DeployParams) ParsedTags() string {
	tags := strings.Replace(p.Tags, " ", "", -1)
	if tags == "" {
		tags = "updt_streambed"
	}
	return tags
}

func (p *DeployParams) String() string {
	branch := p.Branch
	if branch == "" {
		branch = "[default]"
	}
	return fmt.Sprintf("env=%s branch=%s tags=%s by=%s from=%s", p.Environment, p.Branch, p.ParsedTags(), p.InitiatedBy, p.From)
}
