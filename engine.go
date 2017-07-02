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

//var wsSendMessage = make(chan string)
var wsSendMessage = make(chan WsMessageType)

// serveWs - apparently this is what is 'called' from the outside, and I need to talk to a socket here.
func serveWs(ws *websocket.Conn) {
	var (
		err error
		data []byte
	)
	
	log.Println("entering serveWs with connection config:", ws.Config())

	go func() {
		log.Println("entering send loop")

		for {
			sendMessage := <-wsSendMessage
			
			data, err = json.MarshalIndent(sendMessage, "", " ")
			checkErr(err)
			
			if err = websocket.Message.Send(ws, data); err != nil {
				log.Println("Can't send; error:", err)
				break
			}
		}
	}()

	log.Println("entering receive loop")
	var receiveMessage WsMessageType

	for {
		if err = websocket.Message.Receive(ws, &data); err != nil {
			log.Println("Can't receive; error:", err)
			break
		}
		err = json.Unmarshal(data, &receiveMessage)
		checkErr(err)
		
		log.Printf("Received back from client: type '%s' subtype '%s' text '%s' id '%s'\n", *receiveMessage.Type.Ptr(), *receiveMessage.SubType.Ptr(), *receiveMessage.Text.Ptr(), *receiveMessage.Id.Ptr())
		// To-Do Next: client will tell us when it's ready, and send us an agent and a destination cube
		
		var messageType = receiveMessage.Type.Ptr()
		switch *messageType {
			case "formSubmit":
				var messageText = *receiveMessage.Text.Ptr()
				returnValues := strings.Split(messageText, "|")
				Destination := returnValues[0]
				Agent := returnValues[1]
				
				log.Printf("Destination: ", Destination, "Agent:", Agent)
			default:
				log.Printf("Unknown message type", &messageType)
		}
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
		
		cubes += fmt.Sprintf("\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s (%s/%s) [%s (%s,%s,%s)]</option>\n", uuid, name, objType, objClass, regionName, xyz[0], xyz[1], xyz[2])
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
		agentNames += fmt.Sprintf("\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s  (%s) [%s (%s,%s,%s)]</option>\n", ownerKey, name, ownerKey, regionName, xyz)
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

	go engine()
}

// engine does everything but the kitchen sink.
func engine() {
	fmt.Println("this is the engine starting")
	sendMessageToBrowser("status", "", "this is the engine <b>starting</b><br />", "")
	for i := 1; i <= 10; i++ {
		time.Sleep(time.Second * 1)
		fmt.Println(i, " second(s) elapsed")
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
	

	marshalled, err := json.MarshalIndent(msgToSend, "", " ")
	checkErr(err)
	
	select {
	    case wsSendMessage <- msgToSend:
			fmt.Println("Sent: ", string(marshalled))
	    case <-time.After(time.Second * 10):
	        fmt.Println("timeout after 10 seconds; coudn't send:", string(marshalled))
	}
}