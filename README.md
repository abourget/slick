# The Plotly Bot -- A simple bot written in Go

[![Build Status](https://drone.io/github.com/plotly/plotbot/status.png)](https://drone.io/github.com/plotly/plotbot/latest)


## Configuration

(Must be done to run Plotbot locally as well as to deploy it via Ansible.)

* Install your Go environment, under Ubuntu, use this method:

    http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go

* Set your `GOPATH`:

    On Ubuntu see [here](http://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit/21012349#21012349)


* Install Ubuntu dependencies needed by various steps in this document:

    ```sudo apt-get install mercurial zip```

* Pull the bot and its dependencies:

    ```go get github.com/plotly/plotbot/plotbot```

* Install rice:

    ```go get github.com/GeertJohan/go.rice/rice```

* Run "npm install":

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/web
   npm install
   ```

* Run "npm run build":

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/web
   npm run build
   ```

* Patch "https://github.com/tkawachi/hipchat":
  > This is an unfortunate step --- someone will have to make a PR on
  > tkawachi's module or fork it.

   ```
   cd $GOPATH/src/github.com/tkawachi/hipchat/xmpp
   ```
   Open `xmpp.go` and find the codeblock that looks like:
   ```go
   func (c *Conn) UseTLS() {
       c.outgoing = tls.Client(c.outgoing, nil)
       c.incoming = xml.NewDecoder(c.outgoing)
    }
   ```
   and make it look like:
   ```go
   func (c *Conn) UseTLS() {
       c.outgoing = tls.Client(c.outgoing, &tls.Config{ServerName: "chat.hipchat.com"})
       c.incoming = xml.NewDecoder(c.outgoing)
    }
   ```



## Local build and install

* Copy the `plotbot.sample.conf` file to `$HOME/.plotbot` and tweak at will.

* Build with:

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/plotbot
   go build && ./plotbot
   ```
   
   If you are getting an error e.g. `../web/auth.go:11:2:` , try:
   
   ```
   cd $GOPATH/src/github.com/golang/oauth2
   git checkout 49f48241375b4e
   ```
   
   

* Inject static stuff in the binary with:

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/web
   rice append --exec=../plotbot/plotbot
   ```

* Enjoy! You can deploy the binary and it has all the assets in itself now.


## Writing your own plugin

Take inspiration by looking at the different plugins, like `Funny`,
`Healthy`, `Storm`, `Deployer`, etc..  Don't forget to update your
bot's plugins list, like `plotbot/main.go`
