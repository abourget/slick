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

## Local build and install

* Copy the `plotbot.sample.conf` file to `$HOME/.plotbot` and tweak at will.

* Build with:

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/plotbot
   go build && ./plotbot
   ```
   
* Note: It is also possible to build plotbot using the stable dependencies found
        within the Godeps directory. This can be done as follows: 
        
        * Install godep: 
        
           go get github.com/tools/godep
           
        * Now build using the godep tool as follows:
        
           cd $GOPATH/src/github.com/plotly/plotbot/plotbot
           godep go build && ./plotbot
              
                   
* Inject static stuff (for the web app) in the binary with:

   ```
   cd $GOPATH/src/github.com/plotly/plotbot/web
   rice append --exec=../plotbot/plotbot
   ```

* Enjoy! You can deploy the binary and it has all the assets in itself now.


## Writing your own plugin

Take inspiration by looking at the different plugins, like `Funny`,
`Healthy`, `Storm`, `Deployer`, etc..  Don't forget to update your
bot's plugins list, like `plotbot/main.go`
