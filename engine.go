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
		// To-Do Next: client will tell us when it's ready, and send us an agent and a destination cube
	}
}

// engineHandler is still being implemented, it uses the old Go websockets interface to try to keep the page updated.
func backofficeEngine(w http.ResponseWriter, r *http.Request) {
	// start gathering the cubes and agents for the Engine form
		checkSession(w, r)
	// Collect a list of existing bots and their PermURLs for the form
	
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	// query for in-world objects that are cubes (i.e. not Bot Controllers)
	rows, err := db.Query("SELECT * FROM Positions WHERE ObjectType <> 'Bot Controller' ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var ( 
		Position PositionType
		Agent AgentType
		Cubes, regionName, coords = "", "", ""
		xyz []string
	)

	// As on backofficeCommands, but a little more complicated
	for rows.Next() {
		err = rows.Scan(
			&Position.PermURL,
			&Position.UUID,
			&Position.Name,
			&Position.OwnerName,
			&Position.Location,
			&Position.Position,
			&Position.Rotation,
			&Position.Velocity,
			&Position.LastUpdate,
			&Position.OwnerKey,
			&Position.ObjectType,
			&Position.ObjectClass,
			&Position.RateEnergy,
			&Position.RateMoney,
			&Position.RateHappiness,
		)
		checkErr(err)
		// parse name of the region and coordinates
		regionName = Position.Location[:strings.Index(Position.Location, "(")-1]
		coords = strings.Trim(Position.Position, "() \t\n\r")
		xyz = strings.Split(coords, ",")
		
		Cubes += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s (%s/%s) [%s (%s,%s,%s)]</option>\n", Position.UUID, Position.Name, Position.ObjectType, Position.ObjectClass, regionName, xyz[0], xyz[1], xyz[2])
	}

	rows, err = db.Query("SELECT Name, OwnerKey FROM Agents ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var AgentNames = ""
	
	// To-Do: Agent options should also have location etc.

	// find all Names and OwnerKeys and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&Agent.Name, &Agent.OwnerKey)
		checkErr(err)
		AgentNames += "\t\t\t\t\t\t\t\t\t\t\t<option value=\"" + Agent.OwnerKey + "\">" + Agent.Name + " (" + Agent.OwnerKey + ")</option>\n"
	}
	
	db.Close()

	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - engine",
			"URLPathPrefix": template.HTML(URLPathPrefix),
			"Host": AgentNames,
			"DestinationOptions": Cubes,
			"AgentOptions": Cubes,
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