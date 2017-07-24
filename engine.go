// Here is the main engine app.
package main

import (
	_ "github.com/go-sql-driver/mysql"
	"bytes"	
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"gopkg.in/guregu/null.v3/zero"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync/atomic" // used for sync'ing values across goroutines at a low level
	"time"
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

// The following struct is used to hold status information across goroutines
//  It uses the sync/atomic package to do this at a low level, we could have used mutexes (20170704)

//type atomic.Value struct {
//	running bool
//}

// Constants for movement algorithm. Names are retained from the PHP version.
// TODO(gwyneth): Have these constants as variables which are read from the configuration file.

const OS_NPC_SIT_NOW = 0
// for genetic algorithm
const RADIUS = 10 // this is the size of the grid that is thrown around the avatar
const POPULATION_SIZE = 50 // was 50
const GENERATIONS = 20 // was 20 for 20x20 grid
const CHROMOSOMES = 7 // was 28 for 20x20 grid
const CROSSOVER_RATE = 90.0 // = 90%, we use a random number generator for 0-100
const MUTATION_RATE = 5.0   // = 0.005%, we use a random number generator for 0-1000 - TODO(gwyneth): try later with 0.01
const WALKING_SPEED = 3.19 // avatar walking speed in meters per second)
// Weights for Shi & Cui
const W1 = 1.0 // Sub-function of Path Length
const W2 = 10.0 // Sub-function of Path Security
const W3 = 5.0 // Sub-function of Smoothness

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
	
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
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
	
	// defer rows.Close() // already deferred above
 	
	var ownerKey, agentNames = "", ""
	
	// To-Do: Agent options should also have location etc.

	// find all Names and OwnerKeys and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&name, &ownerKey, &location, &position)
		checkErr(err)
		regionName, xyz = convertLocPos(location, position)
		agentNames += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s  (%s) [%s (%s,%s,%s)]</option>\n", ownerKey, name, ownerKey, regionName, xyz)
	}
	
	rows.Close() // closing after deferring to close is probably not good, but I'll try it anyway (20170723)
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
	// we use sync/atomic for making sure we can read a value that is set by a different goroutine
	//  see https://texlution.com/post/golang-lock-free-values-with-atomic-value/ among others (20170704)
	var (
		receiveMessage WsMessageType
		engineRunning atomic.Value // using sync/atomic to make values consistent among goroutines (20170704)
	)
	
	engineRunning.Store(true) // we start by running the engine; note that this may very well happen before we even have WebSockets up (20170704)
	
	fmt.Println("this is the engine starting")
	sendMessageToBrowser("status", "", "this is the engine <b>starting</b><br />", "") // browser might not even know we're sending messages to it, so this will just gracefully timeout and be ignored
	
	// Now, this is a message handler to receive messages while inside the engine, we
	//  block on a message and run a goroutine in the background, so we can safely continue
	//  to run the engine without blocking or errors
	//  I have no idea yet if this is a good idea or not (20170703)
	//  At least it works (20170704)
	go func() {
		var messageType, messageSubType string
				
		for {		
			receiveMessage = <-wsReceiveMessage
			
			if (receiveMessage.Type.Ptr() != nil) {
				messageType = *receiveMessage.Type.Ptr()
			} else {
				messageType = "empty"
			}			
			if (receiveMessage.SubType.Ptr() != nil) {
				messageSubType = *receiveMessage.SubType.Ptr()
			} else {
				messageSubType = "empty"
			}
					
			switch messageType {
				case "status":
					switch messageSubType {
						case "ready": // this is what we get when WebSockets are established on the client
							// check for engine running or not and set the controls
							switch engineRunning.Load().(bool) {
								case true:
									sendMessageToBrowser("htmlControl", "disable", "", "startEngine")
									sendMessageToBrowser("htmlControl", "enable", "", "stopEngine")
								case false:
									sendMessageToBrowser("htmlControl", "enable", "", "startEngine")
									sendMessageToBrowser("htmlControl", "false", "", "stopEngine")
								default: // should never happen, but turn both buttons off just in case
									sendMessageToBrowser("htmlControl", "disable", "", "startEngine")
									sendMessageToBrowser("htmlControl", "disable", "", "stopEngine")
							}
						case "gone": // The client has gone, we have no more websocket for this one (20170704)
							fmt.Println("Client just told us that it went away, we continue on our own")
						default: // no other special functions for now, just echo what the client has sent...
							//unknownMessage := *receiveMessage.Text.Ptr() // better not...
							//fmt.Println("Received from client unknown status message with subtype",
							//	messageSubType, "and text: >>", unknownMessage, "<< — ignoring...")
							fmt.Println("Received from client unknown status message with subtype",
								messageSubType, " — ignoring...")
					}
				case "formSubmit":
					var messageText string
					if receiveMessage.Text.Ptr() != nil {
						messageText = *receiveMessage.Text.Ptr()
					} else {
						messageText = NullUUID + "|" + NullUUID // a bit stupid, we could skip this and do direct assigns, but this way we do a bit more effort wasting CPU cycles for the sake of code clarity (20170704)
					}
					returnValues := strings.Split(messageText, "|")
					Destination := returnValues[0]
					Agent := returnValues[1]
					
					log.Println("Destination: ", Destination, "Agent:", Agent)
					sendMessageToBrowser("status", "info", "Received '" + Destination + "|" + Agent + "'<br />", "")
				case "engineControl":
					switch messageSubType {
						case "start":
							sendMessageToBrowser("htmlControl", "disable", "", "startEngine")
							sendMessageToBrowser("htmlControl", "enable", "", "stopEngine")
							engineRunning.Store(true)
						case "stop":
							sendMessageToBrowser("htmlControl", "enable", "", "startEngine")
							sendMessageToBrowser("htmlControl", "disable", "", "stopEngine")
							engineRunning.Store(false)
						default: // anything will stop the engine!
							sendMessageToBrowser("htmlControl", "enable", "", "startEngine")
							sendMessageToBrowser("htmlControl", "disable", "", "stopEngine")
							engineRunning.Store(false)
					}
					sendMessageToBrowser("status", "info", "Engine " + messageSubType + "<br />", "")
							
				default:
					log.Println("Unknown message type", messageType)
			}
		}
	}()
	
	// We continue with engine. Things may happen in the background, and theoretically we
	//  will be able to catch them. (20170703)
	
	// load whole database in memory. Really. It's so much faster that way! (20170722)
	var (
		Agent AgentType // temporary way to store what comes from database
		Agents []AgentType // we OUGHT to have a type without those strange zero.String, but it's tough to keep two structs in perfect sync (20170722); and this might even become a map, indexed by Agent UUID?
		Position PositionType
		Cubes []PositionType // name to be compatible with PHP version
		Object ObjectType
		Objects []ObjectType
		masterController PositionType // we will need the most recent Bot Master Controller to send commands! (name is the same as in former PHP code).
	)
	
	// Open database
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErr(err)
	
	defer db.Close() // needed?
	
	// load in Agents! We need them to call the movement algorithm for each one
	// BUG(gwyneth): what if the number of agents _change_ while we're running the engine? We need a way to reset the engine somehow. We have a hack at the moment: send a SIGCONT, it will try to restart the engine in a new goroutine
	rows, err := db.Query("SELECT * FROM Agents ORDER BY Name") // can't hurt much to let the DB engine sort it, that way we humans have an idea on how far we've progressed when adding agents
	checkErr(err)

	defer rows.Close() // needed?

	for rows.Next() {
		err = rows.Scan(
			&Agent.UUID,
			&Agent.Name,
			&Agent.OwnerName,
			&Agent.OwnerKey,
			&Agent.Location,
			&Agent.Position,
			&Agent.Rotation,
			&Agent.Velocity,
			&Agent.Energy,
			&Agent.Money,
			&Agent.Happiness,
			&Agent.Class,
			&Agent.SubType,
			&Agent.PermURL,
			&Agent.LastUpdate,
			&Agent.BestPath,
			&Agent.SecondBestPath,
			&Agent.CurrentTarget,
		)		
		// do the magic to extract the actual coords		
		Agent.Coords_xyz = strings.Split(strings.Trim(*Agent.Position.Ptr(), "() \t\n\r"), ",")
		// we should extract the region name from Agent.Location, but I'm lazy!
		Agents = append(Agents, Agent)
	}
	// release DB resources before we start our job
	rows.Close()
	db.Close()
			
	// if we have zero agents, we cannot go on!
	// TODO(gwyneth): be more graceful handling this, because the engine will stop forever this way
	if len(Agents) == 0 {
		log.Println("Error: no Agents found. Engine cannot run. Aborted. Add an Agent and try sending a SIGCONT to restart engine again")
		sendMessageToBrowser("status", "restart", "Error: no Agents found. Engine cannot run. Aborted. Add an Agent and try sending a <code>SIGCONT</code> to restart engine again<br />"," ")
		return
	}
			
	for {
		for i, Agent := range Agents {
			// check if we should be running or not
			if engineRunning.Load().(bool) {
				log.Println("Starting to manipulate Agent", i, "-", *Agent.Name.Ptr())
				// We need to refresh all the data about cubes and positions again!

				// do stuff while it runs, e.g. open databases, search for agents and so forth
				
				log.Println("Reloading database for Cubes (Positions) and Obstacles...")
				
				// Open database
				db, err = sql.Open(PDO_Prefix, GoBotDSN)
				checkErr(err)
				
				defer db.Close() // needed?
				
				// Load in the 'special' objects (cubes). Because the Master Controllers can be somewhere in here, to save code.
				//  and a database query, we simply skip all the Master Controllers until we get the most recent one, which gets saved
				//  The rest of the objects are cubes, so we will need them in the ObjectType array (20170722).
				// BUG(gwyneth): Does not work across regions! We will probably need a map of bot controllers for that and check which one to call depending on the region of the current agent; simple, but I'm lazy (20170722).
				
				Cubes = nil // clear array, let the Go garbage collector deal with the memory (20170723)
				rows, err = db.Query("SELECT * FROM Positions ORDER BY LastUpdate ASC")
				checkErr(err)
						
				defer rows.Close() // needed?
						
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
					Position.Coords_xyz = strings.Split(strings.Trim(*Position.Position.Ptr(), "() \t\n\r"), ",")
					
					// check if we got a Master Bot Controller!
					if (*Position.ObjectType.Ptr() == "Bot Controller") {
						masterController = Position // this will get overwritten until we get the last, most recent one
					} else {
						Cubes = append(Cubes, Position) // if not a controller, it must be a cube! add it to array!
					}
				}
				
				// load in everything we found out so far on our region(s) but ignore phantom objects
				// end-users ought to set their cubes to phantom as well, or else the agents will think of them as obstacles!
				Objects = nil
				rows, err = db.Query("SELECT * FROM Obstacles WHERE Phantom = 0")
				checkErr(err)
						
				for rows.Next() {
					err = rows.Scan(
						&Object.UUID,
						&Object.Name,
						&Object.BotKey,
						&Object.BotName,
						&Object.Type,
						&Object.Position,
						&Object.Rotation,
						&Object.Velocity,
						&Object.LastUpdate,
						&Object.Origin,
						&Object.Phantom,
						&Object.Prims,
						&Object.BBHi,
						&Object.BBLo,
					)
					Object.Coords_xyz = strings.Split(strings.Trim(*Object.Position.Ptr(), "() \t\n\r"), ",")
					
					Objects = append(Objects, Object)
				}
				
				// release DB resources before we start our job
				rows.Close()
				db.Close()
				
				// Do not trust the database with the exact Agent position: ask the master controller directly
				// log.Println(*masterController.PermURL.Ptr())
				
				//log.Println(masterController, Position, Cubes, Object, Objects)
				curposResult, _ := callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcGetPos");
				sendMessageToBrowser("status", "info", "<p class='box'>Grid reports that agent " + *Agent.Name.Ptr() + " is at position: " + curposResult + "</p>\n", "") 
				log.Println("Grid reports that agent", *Agent.Name.Ptr(), "is at position:", curposResult)
				
				// output something to console so that we know this is being run in parallel
			    fmt.Print("\r|")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r/")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r-")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r\\")
			    time.Sleep(1000 * time.Millisecond)
			} else {
				// stop everything!!!
				// in theory this is used to deal with reconfigurations etc.
			    fmt.Print("\r𝔷")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\rz")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\rZ")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\rℤ")
			    time.Sleep(1000 * time.Millisecond)			
			}
		}
    }
    
    // Why should we ever stop? :)
	sendMessageToBrowser("status", "", "this is the engine <i>stopping</i><br />", "")
	fmt.Println("this is the engine stopping")
}

// sendMessageToBrowser sends a string to the internal, global channel which is hopefully picked up by the websocket handling goroutine.
func sendMessageToBrowser(msgType string, msgSubType string, msgText string, msgId string) {
	var msgToSend WsMessageType
	
	msgToSend.New(msgType, msgSubType, msgText, msgId)
	
	marshalled, err := json.MarshalIndent(msgToSend, "", " ") // debug line just to show msgToSend's structure
	checkErr(err)
	
	select {
	    case wsSendMessage <- msgToSend:
			fmt.Println("Sent: ", string(marshalled))
	    case <-time.After(time.Second * 10):
	        fmt.Println("timeout after 10 seconds; coudn't send:", string(marshalled))
	}
}

// callURL encapsulates a call to an URL. It exists as an analogy to the PHP version (20170723).
func callURL(url string, encodedRequest string) (string, error) {
	//  HTTP request as per http://moazzam-khan.com/blog/golang-make-http-requests/
	body := []byte(encodedRequest)
    
    rs, err := http.Post(url, "body/type", bytes.NewBuffer(body))
    // Code to process response (written in Get request snippet) goes here

	defer rs.Body.Close()
	
	rsBody, err := ioutil.ReadAll(rs.Body)
	if (err != nil) {
		errMsg := fmt.Sprintf("Error response from in-world object: %s", err)
		log.Println(errMsg)
		return errMsg, err
	} else {
	    log.Printf("Reply from in-world object %s\n", rsBody)
		return string(rsBody), err
	}	
}