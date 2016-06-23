# Slick - A golang Slack bot

[![Build Status](https://drone.io/github.com/abourget/slick/status.png)](https://drone.io/github.com/abourget/slick/latest)


## Features

Supported features:

* Plugin interface for chat messages and HTTP handlers.
* The bot keeps an internal state of channels, users and users' states.
* Easily add listeners with criterias like:
  * Messages directed to the bot only
  * Private or public messages
  * Listens for a duration or until a given time.Time
  * Selectively on a channel, or from a user
  * Expire listeners and unregister them dynamically
  * Supports listening for edits or not
  * Regexp match messages, or Contains checks
* Simple API to reply to users
* Supports listening for any Slack events (ChannelCreated, ChannelJoined, EmojiChanged, FileShared, GroupArchived, etc..)
* Listen easily for reactions, and take actions based on them (like buttons).
* Simple API to private message users
* Simple API to update a previously sent message incrementally
* Simple API to delete bot messages after a given time duration.
* Built-in KV store for data persistence (backed by BoltDB and json serialization)
* The bot has a mood (_happy_ and _hyper_) which changes randomly.. you can base some decisions on it, to spice up conversations.
* An PubSub system to ease inter-plugins (or chat-to-web) communications.

## Stock plugins

1. Recognition: a plugin to recognize your peers (!recognize @user1 for doing an awesome job)

2. Faceoff: a game to learn the names and faces of your colleagues. The code for this one is interesting to learn to build interactive features with `slick`.

3. Vote: a simple voting plugin to decide where to lunch

4. Funny: a bunch of jokes and memes in reply to some strings in channels..

5. Healthy: a very simple plugin that pokes URLs and reports on their health

6. Deployer: an example plugin to do deployments wth ansible (you'll probably want to roll out your own though).


## Local build and install

Try it with:

```
go get github.com/abourget/slick
cd $GOPATH/src/github.com/abourget/slick/example-bot
go install -v && $GOPATH/bin/example-bot
```

Copy the `slick.sample.conf` file to `$HOME/.slick` and tweak at will.


## Writing your own plugin


Example code to handle deployments:

```
// listenDeploy was hooked into a plugin elsewhere..
func listenDeploy() {
	keywords := []string{"project1", "project2", "project3"}
	bot.Listen(&slick.Listener{
		Matches:        regexp.MustCompile("(can you|could you|please|plz|c'mon|icanhaz) deploy (" + strings.Join(keywords, "|") + ") (with|using)( revision| commit)? `?([a-z0-9]{4,42})`?"),
		MentionsMeOnly: true,
		MessageHandlerFunc: func(listen *slick.Listener, msg *slick.Message) {

			projectName := msg.Match[2]
			revision := msg.Match[5]

			go func() {
				go msg.AddReaction("work_hard")
				defer msg.RemoveReaction("work_hard")

				// Do the deployment with projectName and revision...

			}()
		},
	})
}
```




Take inspiration by looking at the different plugins, like `Funny`,
`Healthy`, `Storm`, `Deployer`, etc..  Don't forget to update your
bot's plugins list, like `example-bot/main.go`


## Configuration

You might need `mercurial` installed to get some dependencies.
