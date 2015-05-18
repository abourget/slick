# Slick - A golang Slack bot

[![Build Status](https://drone.io/github.com/abourget/slick/status.png)](https://drone.io/github.com/abourget/slick/latest)


## Configuration

* Install your Go environment, under Ubuntu, use this method:

    http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go

* Set your `GOPATH`:

    On Ubuntu see [here](http://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit/21012349#21012349)


* Install Ubuntu dependencies needed by various steps in this document:

    ```sudo apt-get install mercurial zip```

* Pull the bot and its dependencies:

    ```go get github.com/abourget/slick/example-bot```

* Install rice:

    ```go get github.com/GeertJohan/go.rice/rice```

* Run "npm install":

   ```
   cd $GOPATH/src/github.com/abourget/slick/web
   npm install
   ```

* Run "npm run build":

   ```
   cd $GOPATH/src/github.com/abourget/slick/web
   npm run build
   ```

## Local build and install

* Copy the `slick.sample.conf` file to `$HOME/.slick` and tweak at will.

* Build with:

   ```
   cd $GOPATH/src/github.com/abourget/slick/example-bot
   go build && ./example-bot
   ```

* Note: It is also possible to build your bot using the stable dependencies found
        within the Godeps directory. This can be done as follows:

        Install godep:

           go get github.com/tools/godep

        Now build using the godep tool as follows:

           cd $GOPATH/src/github.com/abourget/slick/example-bot
           godep go build && ./example-bot


* Inject static stuff (for the web app) in the binary with:

   ```
   cd $GOPATH/src/github.com/abourget/slick/web
   rice append --exec=../example-bot/example-bot
   ```

* Enjoy! You can deploy the binary and it has all the assets in itself now.


## Writing your own plugin

Take inspiration by looking at the different plugins, like `Funny`,
`Healthy`, `Storm`, `Deployer`, etc..  Don't forget to update your
bot's plugins list, like `example-bot/main.go`
