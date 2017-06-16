// Here is the main engine app.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"html/template"
	"golang.org/x/net/websocket"
)

var wsSendMessage = make(chan string)

// serveWs - apparently this is what is 'called' from the outside, and I need to talk to a socket here.
func serveWs(ws *websocket.Conn) {
	var err error
	log.Println("entering serveWs with connection config:", ws.Config())

	go func() {
		log.Println("entering send loop")

		for {
			sendMessage := <-wsSendMessage
			if err = websocket.Message.Send(ws, sendMessage); err != nil {
				fmt.Println("Can't send; error:", err)
				break
			}
		}
	}()

	log.Println("entering receive loop")
	var receiveMessage string

	for {
		if err = websocket.Message.Receive(ws, &receiveMessage); err != nil {
			fmt.Println("Can't receive; error:", err)
			break
		}
		log.Println("Received back from client: '" + receiveMessage + "'")
	}
}

// engineHandler is still being implemented, it uses the old Go websockets interface to try to keep the page updated.
func backofficeEngine(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - engine",
			"URLPathPrefix": template.HTML(URLPathPrefix),
			"Host": Host,
			"ServerPort": template.HTML(ServerPort),
			"Content": template.HTML("<hr />"),
	}
	err := GobotTemplates.gobotRenderer(w, r, "engine", tplParams)
	checkErr(err)

	go engine()
}

// engine does everything but the kitchen sink.
func engine() {
	fmt.Println("this is the engine starting")
	sendMessageToBrowser("this is the engine <b>starting</b><br />")
	for i := 1; i <= 10; i++ {
		time.Sleep(time.Second * 1)
		fmt.Println(i, " second(s) elapsed")
	}
	sendMessageToBrowser("this is the engine <i>stopping</i><br />")
	fmt.Println("this is the engine stopping")
}

// sendMessageToBrowser sends a string to the internal, global channel which is hopefully picked up by the websocket handling goroutine.
func sendMessageToBrowser(msg string) {
	select {
	    case wsSendMessage <- msg:
			fmt.Println("Sent: " + msg)
	    case <-time.After(time.Second * 10):
	        fmt.Println("timeout after 10 seconds; coudn't send:", msg)
	}
}