package main

import (
	"log"
	"net/http"
//	"time"
	"html/template"
	"github.com/gorilla/websocket"
)

// Stuff to deal with WebSockets, based on https://github.com/gorilla/websocket/blob/master/examples/filewatch/main.go

var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
}

var message = make(chan []byte) // yay, we have a socket or something here

// reader - I have no idea what this does
func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
//  Commented out until I figure out what this is for
//	ws.SetReadDeadline(time.Now().Add(pongWait))
//	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writer - I hope I'm getting this right
func writer(ws *websocket.Conn) {
	defer ws.Close()

	var msg []byte = <-message

	if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		return
	}
}

// serveWs - apparently this is what is 'called' from the outside, and I need to talk to a socket here
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	defer ws.Close()
	log.Println("Got here")
	
	go writer(ws)
	reader(ws)
}

// blockingWrite sends a message via a goroutine, I hope
func blockingWrite(msg string){
	message <- []byte(msg)
}

// engine is still being implemented, it uses Gorilla's WebSockets to try to keep the page updated
func engine(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - engine",
			"URLPathPrefix": template.HTML(URLPathPrefix),
			"Host": Host,
			"ServerPort": template.HTML(ServerPort),
			"Content": "Under implementation",
	}
	err := GobotTemplates.gobotRenderer(w, r, "engine", tplParams)
	checkErr(err)
	
	go blockingWrite("Did I write anything?")
	
	//message <- []byte("Did I write anything?")
	// go writer(ws) // apparently we have to put this in a goroutine so that it doesn't block us
	// reader(ws) // do we *really* need this shit?
}