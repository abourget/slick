package deployer

import (
	"fmt"

	"golang.org/x/net/websocket"

	"github.com/gorilla/mux"
	"github.com/abourget/slick"
)

func (dep *Deployer) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	privRouter.Handle("/plugins/deployer.ws", websocket.Handler(dep.websocketManager))
}

func (dep *Deployer) websocketManager(ws *websocket.Conn) {
	// Hook into the PubSub
	fmt.Println("Entering websocketManager()")
	incoming := dep.pubsub.Sub("ansible-line")
	defer dep.pubsub.Unsub(incoming, "ansible-line")

	go dep.websocketReader(ws)

	for msg := range incoming {
		if err := websocket.Message.Send(ws, msg); err != nil {
			ws.Close()
			return
		}
	}

	fmt.Println("Ok, done with websocketManager")
	// Send anything from PubSub to this ws connection, until closed
	// We'll read from PubSub populating ansible output on "standup"
	// and push to any web user

}

func (dep *Deployer) websocketReader(ws *websocket.Conn) {
	deployParams := &DeployParams{From: "web"}
	for {
		err := websocket.JSON.Receive(ws, &deployParams)
		if err != nil {
			fmt.Println("Deployer: Error on JSON receive", err)
			return
		}
		fmt.Println("Deployer: Received deployparams from web", deployParams)
		if dep.runningJob != nil {
			fmt.Println("Deployer: Job already running")
			if err := websocket.Message.Send(ws, "[Deployer: Job already running]"); err != nil {
				ws.Close()
				return
			}
			continue
		}

		go dep.handleDeploy(deployParams)
	}
}
