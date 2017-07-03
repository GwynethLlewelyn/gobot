// Here is the main engine app.
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
	"html/template"
	"golang.org/x/net/websocket"
	"strings"
	"encoding/json"
	"gopkg.in/guregu/null.v3/zero"
)

// Define a communications procotol with the client, so that we can selectively
//  send messages to turn options on and off, etc.
// Messages will be JSON.

type WsMessageType struct {
	Type zero.String `json:"type"`
	SubType zero.String `json:"subtype"`
	Text zero.String `json:"text"`
	Id zero.String `json:"id"`
}

// New creates a new WsMessage out of 4 strings
func (wsM *WsMessageType) New(msgType string, msgSubType string, msgText string, msgId string) *WsMessageType {
	wsM.Type = zero.StringFrom(msgType)
	wsM.SubType = zero.StringFrom(msgSubType)
	wsM.Text = zero.StringFrom(msgText)
	wsM.Id = zero.StringFrom(msgId)

	return wsM
}

// Go is tricky. While we send and receive WebSocket messages as it would be expected on a 'normal' 
//  programming language, we actually have an insane amount of goroutines all in parallel. So what we do is to 
//  send messages to a 'channel' (Go's version of a semaphore) and receive them from a different one; two sets
//  of goroutines will have their fun reading and sending messages to the client and updating the channels,
//  so other goroutines only send and receive to the channels and have no concept of 'WebSocket messages'
//  This is sort of neat because it solves parallelism (goroutines block on sockets) but it also allows
//  us to build in other transfer mechanisms and make them abstract using Go channels (20170703)
var wsSendMessage = make(chan WsMessageType)
var wsReceiveMessage = make(chan WsMessageType)

// serveWs - apparently this is what is 'called' from the outside, and I need to talk to a socket here.
func serveWs(ws *websocket.Conn) {
	// see also how it is implemented here: http://eli.thegreenplace.net/2016/go-websocket-server-sample/ (20170703)
	var (
		err error
//		data []byte
	)
	
	if ws == nil {
		log.Panic("Received nil WebSocket — I have no idea why this happened!")
	}
	
	log.Printf("Client connected from %s", ws.RemoteAddr())
	log.Println("entering serveWs with connection config:", ws.Config())

	go func() {
		log.Println("entering send loop")

		for {
			sendMessage := <-wsSendMessage
			
			if err = websocket.JSON.Send(ws, sendMessage); err != nil {
				log.Println("Can't send; error:", err)
				break
			}
		}
	}()

	log.Println("entering receive loop")
	var receiveMessage WsMessageType

	for {
		if err = websocket.JSON.Receive(ws, &receiveMessage); err != nil {
			log.Println("Can't receive; error:", err)
			break
		}
		log.Println("Received message", receiveMessage)
		
		// log.Printf("Received back from client: type '%s' subtype '%s' text '%s' id '%s'\n", *receiveMessage.Type.Ptr(), *receiveMessage.SubType.Ptr(), *receiveMessage.Text.Ptr(), *receiveMessage.Id.Ptr())
		// To-Do Next: client will tell us when it's ready, and send us an agent and a destination cube
		
		wsReceiveMessage <- receiveMessage
	}
}

// convertLocPos converts a SL/OpenSim Location and Position into a single region name and (x,y,z) position coordinates
func convertLocPos(location string, position string) (regionName string, xyz []string) {
	regionName = location[:strings.Index(location, "(")-1]
	coords := strings.Trim(position, "() \t\n\r")
	xyz = strings.Split(coords, ",")
	return regionName, xyz
}

// engineHandler is still being implemented, it uses the old Go websockets interface to try to keep the page updated.
func backofficeEngine(w http.ResponseWriter, r *http.Request) {
	// start gathering the cubes and agents for the Engine form
	checkSession(w, r)
	// Collect a list of existing bots and their PermURLs for the form
	
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	// query for in-world objects that are cubes (i.e. not Bot Controllers)
	rows, err := db.Query("SELECT UUID, Name, ObjectType, ObjectClass, Location, Position FROM Positions WHERE ObjectType <> 'Bot Controller' ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var (
		cubes, regionName = "", ""
		uuid, name, objType, objClass, location, position = "", "", "", "", "", ""
		xyz []string
	)

	// As on backofficeCommands, but a little more complicated
	for rows.Next() {
		err = rows.Scan(&uuid, &name, &objType, &objClass, &location, &position)
		checkErr(err)
		// parse name of the region and coordinates
		regionName, xyz = convertLocPos(location, position)
		
		cubes += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s (%s/%s) [%s (%s,%s,%s)]</option>\n", uuid, name, objType, objClass, regionName, xyz[0], xyz[1], xyz[2])
	}

	rows, err = db.Query("SELECT Name, OwnerKey, Location, Position FROM Agents ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var ownerKey, agentNames = "", ""
	
	// To-Do: Agent options should also have location etc.

	// find all Names and OwnerKeys and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&name, &ownerKey, &location, &position)
		checkErr(err)
		regionName, xyz = convertLocPos(location, position)
		agentNames += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s  (%s) [%s (%s,%s,%s)]</option>\n", ownerKey, name, ownerKey, regionName, xyz)
	}
	
	db.Close()

	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - engine",
			"URLPathPrefix": template.HTML(URLPathPrefix),
			"Host": template.HTML(Host),
			"DestinationOptions": template.HTML(cubes),
			"AgentOptions": template.HTML(agentNames),
			"ServerPort": template.HTML(ServerPort),
			"Content": template.HTML("<hr />"),
	}
	err = GobotTemplates.gobotRenderer(w, r, "engine", tplParams)
	checkErr(err)
}

// engine does everything but the kitchen sink.
func engine() {
	var receiveMessage WsMessageType
	var engineRunning = true // this MAY have some race conditions... but it is mostly just to start/stop the engine, so it should be ok (20170703)
	
	// The theory is the following: when the browser is ready with a connection, it sends us
	//  a message first. We don't know when this happens, so we block on the message queue until
	//  we get something. This might not be a good idea if the client dies, but we have a problem
	//  figuring out when both client and server are ready to exchange messages with each other! (20170703)
	receiveMessage = <-wsReceiveMessage
	fmt.Println("this is the engine starting")
	sendMessageToBrowser("status", "", "this is the engine <b>starting</b><br />", "")
	
	// Now, this is a message handler to receive messages while inside the engine, we
	//  block on a message and run a goroutine in the background, so we can safely continue
	//  to run the engine without blocking or errors
	//  I have no idea yet if this is a good idea or not (20170703)
	go func() {
		for {
			receiveMessage = <-wsReceiveMessage
			
			var messageType = receiveMessage.Type.Ptr()
			switch *messageType {
				case "formSubmit":
					messageText := *receiveMessage.Text.Ptr()
					returnValues := strings.Split(messageText, "|")
					Destination := returnValues[0]
					Agent := returnValues[1]
					
					log.Println("Destination: ", Destination, "Agent:", Agent)
					sendMessageToBrowser("status", "info", "Received '" + Destination + "|" + Agent + "'<br />", "")
				case "engineControl":
					var messageSubType = receiveMessage.SubType.Ptr()
					switch *messageSubType {
						case "start":
							engineRunning = true
						case "stop":
						default:
							engineRunning = false
					}
					sendMessageToBrowser("status", "info", "Engine " + *messageSubType + "<br />", "")
							
				default:
					log.Println("Unknown message type", &messageType)
			}
		}
	}()
	
	// We continue with engine. Things may happen in the background, and theoretically we
	//  will be able to catch them. (20170703)
	fmt.Println("Pretending to do something in parallel while we wait for connections etc...")
	for true {
		if engineRunning {
		    fmt.Print("\b|")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\b/")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\b-")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\b\\")
		    time.Sleep(1000 * time.Millisecond)
		} else {
		    fmt.Print("\bz")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\bzZ")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\bzZz")
		    time.Sleep(1000 * time.Millisecond)
		    fmt.Print("\bzZzZ")
		    time.Sleep(1000 * time.Millisecond)			
		}
    }
	sendMessageToBrowser("status", "", "this is the engine <i>stopping</i><br />", "")
	fmt.Println("this is the engine stopping")
}

// sendMessageToBrowser sends a string to the internal, global channel which is hopefully picked up by the websocket handling goroutine.
func sendMessageToBrowser(msgType string, msgSubType string, msgText string, msgId string) {
	var msgToSend WsMessageType
	
	msgToSend.New(msgType, msgSubType, msgText, msgId)
	
/*	msgToSend.Type = zero.StringFrom(msgType)
	msgToSend.SubType = zero.StringFrom(msgSubType)
	msgToSend.Text = zero.StringFrom(msgText)
	msgToSend.Id = zero.StringFrom(msgId)
*/
	

	marshalled, err := json.MarshalIndent(msgToSend, "", " ") // debug line just to show msgToSend's structure
	checkErr(err)
	
	select {
	    case wsSendMessage <- msgToSend:
			fmt.Println("Sent: ", string(marshalled))
	    case <-time.After(time.Second * 10):
	        fmt.Println("timeout after 10 seconds; coudn't send:", string(marshalled))
	}
}