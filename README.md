# ahipbot -- A simple bot written in Go

[![Build Status](https://drone.io/github.com/abourget/ahipbot/status.png)](https://drone.io/github.com/abourget/ahipbot/latest)


## Configuration

* Install your Go environment, under Ubuntu, use this method:

    http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go

* Pull the bot and its dependencies:

    go get github.com/abourget/ahipbot

* Copy the `dot.hipbot` file to `$HOME/.hipbot` and tweak at will.

* Build with:

    cd src/github.com/abourget/ahipbot
    go build && ./ahipbot

* Inject static stuff in the binary with:

    rice append --exec=ahipbot

* Enjoy! You can deploy the binary and it has all the assets in itself now.


## Writing your own plugin

Take inspiration by looking at `funny.go`.  Write your own, and don't forget to
add your plugin to `Hipbot.registerPlugins()` in `hipbot.go`.
