# The Plotly Bot -- A simple bot written in Go

[![Build Status](https://drone.io/github.com/plotly/plotbot/status.png)](https://drone.io/github.com/plotly/plotbot/latest)


## Configuration

* Install your Go environment, under Ubuntu, use this method:

    http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go

* Pull the bot and its dependencies:

    go get github.com/plotly/plotbot/plotbot
    go install github.com/GeertJohan/go.rice/rice

* Copy the `plotbot.sample.conf` file to `$HOME/.plotbot` and tweak at will.

* Build with:

    cd $GOPATH/src/github.com/plotly/plotbot/plotbot
    go build && ./plotbot

* Inject static stuff in the binary with:

    cd $GOPATH/src/github.com/plotly/plotbot/web
    rice append --exec=../plotbot/plotbot

* Enjoy! You can deploy the binary and it has all the assets in itself now.


## Writing your own plugin

Take inspiration by looking at the different plugins, like `Funny`,
`Healthy`, `Storm`, `Deployer`, etc..  Don't forget to update your
bot's plugins list, like `plotbot/main.go`
