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
	"math"
	"math/rand"
	"net/http"
	"sort"
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

// Constants for genetic algorithm. Names are retained from the PHP version.
// TODO(gwyneth): Have these constants as variables which are read from the configuration file.

const OS_NPC_SIT_NOW = "0"
// for genetic algorithm
const RADIUS = 10.0 // this is the size of the grid that is thrown around the avatar
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

// When transposing from the PHP version, we now cannot avoid having a few structs and types, since Go
//  is a strongly-typed language (20170726)
//  This was moved out of the GA code body because some external functions need those types (20170727)

// chromosomeType is just a point in a path, really.
type chromosomeType struct {
	x, y, z, distance, obstacle, angle, smoothness float64
}

// popType represents each population as a list of points (= chromosomes) indicating a possible path; it also includes the fitness for this particular path.
type popType struct {
	fitness float64
	chromosomes []chromosomeType
}

// movementJob is used in the worker goroutine which processes the points to move the avatars to, which needs to wait until the avatars have moved.
// Go is so quick in recalculating generations that the avatars never get a chance to reach their destination until we wait for them!
// So the commands to move the avatars need to go into a separate goroutine, to wait on avatars, while the main engine continues (20170730).
type movementJob struct {
	agentUUID string				// Agent UUID to move
	masterControllerPermURL string	// masterController to use (note that the engine may pick one of several active ones)
	destPoint chromosomeType		// destination to go to; it's a chromosome so that we get distance information as well to calculate
									//  for how long we need to sleep until the avatar reaches destination
}
// movementJobChannel is the blocking channel to which we write points for the next bot movement
var movementJobChannel = make(chan movementJob, CHROMOSOMES) // for now, we'll try to 

// Go is tricky. While we send and receive WebSocket messages as it would be expected on a 'normal' 
//  programming language, we actually have an insane amount of goroutines all in parallel. So what we do is to 
//  send messages to a 'channel' (Go's version of a semaphore) and receive them from a different one; two sets
//  of goroutines will have their fun reading and sending messages to the client and updating the channels,
//  so other goroutines only send and receive to the channels and have no concept of 'WebSocket messages'
//  This is sort of neat because it solves parallelism (goroutines block on sockets) but it also allows
//  us to build in other transfer mechanisms and make them abstract using Go channels (20170703)
var wsSendMessage = make(chan WsMessageType)
var wsReceiveMessage = make(chan WsMessageType)
var webSocketActive atomic.Value // this is an attempt to check if we have an active WebSocket, to avoid too many timeouts (20170728)

// serveWs - apparently this is what is 'called' from the outside, and I need to talk to a socket here.
func serveWs(ws *websocket.Conn) {
	// see also how it is implemented here: http://eli.thegreenplace.net/2016/go-websocket-server-sample/ (20170703)
	var err error // to avoid constant redeclarations in tight loop below
	
	if ws == nil {
		log.Panic("Received nil WebSocket — I have no idea why or how this happened!")
	}
	
	/*
	log.Printf("Client connected from %s", ws.RemoteAddr())
	log.Println("entering serveWs with connection config:", ws.Config())
	*/

	webSocketActive.Store(true)
	defer webSocketActive.Store(false)

	go func() {
	//	log.Println("entering send loop")

		for {
			sendMessage := <-wsSendMessage
			
			if err = websocket.JSON.Send(ws, sendMessage); err != nil {
				log.Println("Can't send; error:", err)
				break
			}
		}
	}()

	//log.Println("entering receive loop")
	var receiveMessage WsMessageType

	for {
		if err = websocket.JSON.Receive(ws, &receiveMessage); err != nil {
			log.Println("Can't receive; error:", err)
			break
		}
		// log.Println("Received message", receiveMessage)
		
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

// calcDistance calculates the distance between two points, which are actually arrays of x,y,z string coordinates.
//  TODO(gwyneth): Now that we have a strongly-typed language, we should create real objects for this.
func calcDistance(vec1, vec2 []float64) float64 {
	deltaX := vec2[0] - vec1[0] // using extra variables because multiplication is probably
	deltaY := vec2[1] - vec1[1] //  simpler than calling the math.Pow() function (20170725)
	deltaZ := vec2[2] - vec1[2]
	
	return math.Sqrt(deltaX * deltaX + deltaY * deltaY + deltaZ * deltaZ)
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
	cubes = "\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"" + NullUUID + "\">Clean selection (let engine figure out next cube)</option>\n"

	// As on backofficeCommands, but a little more complicated
	for rows.Next() {
		err = rows.Scan(&uuid, &name, &objType, &objClass, &location, &position)
		checkErr(err)
		// parse name of the region and coordinates
		regionName, xyz = convertLocPos(location, position)
		
		cubes += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s (%s/%s) [%s (%s,%s,%s)]</option>\n", uuid, name, objType, objClass, regionName, xyz[0], xyz[1], xyz[2])
	}

	rows, err = db.Query("SELECT Name, UUID, Location, Position FROM Agents ORDER BY Name")
	checkErr(err)
	
	// defer rows.Close() // already deferred above
 	
	var uuidAgent, agentNames = "", ""
	agentNames = "\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"" + NullUUID + "\">Clean selection (let engine figure out next agent)</option>\n"
	
	// To-Do: Agent options should also have location etc.

	// find all Names and OwnerKeys and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&name, &uuidAgent, &location, &position)
		checkErr(err)
		regionName, xyz = convertLocPos(location, position)
		agentNames += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s  (%s) [%s (%s,%s,%s)]</option>\n", uuidAgent, name, uuidAgent, regionName, xyz)
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
		userDestCube atomic.Value
		curAgent atomic.Value
	)
	
	engineRunning.Store(true) // we start by running the engine; note that this may very well happen before we even have WebSockets up (20170704)
	userDestCube.Store(NullUUID) // we start to nullify these atomic values, either they will be changed by the user,
	curAgent.Store(NullUUID)	//  or the engine will simply go through all agents (20170725)
	webSocketActive.Store(false)	// as soon as we know that we have a connection to the client, we set this to true (20170728) 
	
	sendMessageToBrowser("status", "info", "this is the engine <b>starting</b><br />", "") // browser might not even know we're sending messages to it, so this will just gracefully timeout and be ignored and just appear on the log
	
	// Launch the movement worker goroutine. This is needed because Go is so fast calculating populations that it keeps giving the agents
	// contradictory movement commands. This uses a blocking channel and calculates how long the avatar needs to reach its destination
	// and sleeps for that time. There was something similar done in PHP as well, but PHP took long enough recalculating everything, so
	// it was deemed not to be necessary. (20170730)
	go movementWorker()
	
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
							webSocketActive.Store(true)
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
							webSocketActive.Store(false)
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
					userDestCube.Store(returnValues[0])
					curAgent.Store(returnValues[1])
					
					// Commented out because we know this works and we'll print it out later on anyway (20170730)
					// log.Println("Destination: ", userDestCube.Load().(string), "Agent:", curAgent.Load().(string))
					// sendMessageToBrowser("status", "info", "Received '" + userDestCube.Load().(string) + "|" + curAgent.Load().(string) + "'<br />", "")
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
					sendMessageToBrowser("status", "", "Engine " + messageSubType + "<br />", "")
							
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
		Agents map[string]AgentType // we OUGHT to have a type without those strange zero.String, but it's tough to keep two structs in perfect sync (20170722); this is mapped by Agent UUID (20170725)
		Position PositionType
		Cubes map[string]PositionType // name to be compatible with PHP version; mapped by UUID (20170725).
		Object ObjectType
		Obstacles []ObjectType
		masterController PositionType // set to the most recent Bot Master Controller to send commands (name is the same as in former PHP code).
	)
	// NOTE(gwyneth): The reason why we use maps and not slices (slices may be faster) is just because that way we can
    //  directly address the element by UUID, instead of doing array searches (20170725)
	
	
	// Open database
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErr(err)
	
	defer db.Close() // needed?
	
	// load in Agents! We need them to call the movement algorithm for each one
	// BUG(gwyneth): what if the number of agents _change_ while we're running the engine? We need a way to reset the engine somehow. We have a hack at the moment: send a SIGCONT, it will try to restart the engine in a new goroutine
	rows, err := db.Query("SELECT * FROM Agents ORDER BY Name") // can't hurt much to let the DB engine sort it, that way we humans have an idea on how far we've progressed when adding agents; also, we index it by UUID
	checkErr(err)

	defer rows.Close() // needed?

	Agents = make(map[string]AgentType) // initialise our Agents map (20170725)
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
		//log.Println("Agent UUID is", *Agent.UUID.Ptr())
		Agents[*Agent.UUID.Ptr()] = Agent // mwahahahaha (20170725)
	}
	// release DB resources before we start our job
	rows.Close()
	db.Close()
			
	// if we have zero agents, we cannot go on!
	// TODO(gwyneth): be more graceful handling this, because the engine will stop forever this way
	// TODO(gwyneth): Better to randomly pick an agent from the database, and if none is available, skip a cycle (20170730).
	if len(Agents) == 0 {
		sendMessageToBrowser("status", "error", "Error: no Agents found. Engine cannot run. Aborted. Add an Agent and try sending a <code>SIGCONT</code> to restart engine again<br />"," ")
		return // ouchie...
	}
	
	// prepare data to be saved as a CSV/XML file for later import into Excel and do nice graphics
	var export_rows []string // we place it here because of potential scope issues later on...
	
	// Theoretically endless loop follows (20170730)	
	for {
		// Now, the problem with the approach of going through the list of Agents is that new Agents might appear, old
		//  might be deleted, and then we're stuck! (Remember, the updating of the Agents table is done in parallel to this)
		//  The idea of running goroutines for each Agent will also suffer from the same problem: what if the Agent dies and we don't know
		//  about it? Of course we can check with a ping first. What about *new* Agents? How do we launch new goroutines for them if we
		//  don't know about them beforehand? (20170801)
		
		// Second approach (20170801): initialise lastAgentRunning with NullUUID; pick one agent from the database; if it's the
		//  same as before, pick a new one; if the user has provided us with an agent, use that one instead. This will at least provide
		//  all agents with a chance of running, while allowing new Agents to appear and old ones to die (20170801). The cost of this
		//  solution is that *some* Agents may not have a chance to run (since they're picked randomly), so we might be a little more
		//  evil and use some magical pseudo-random generators from Go which allow a sequence of non-repeated random numbers to be
		//  generated, and try to follow that order if possible, which means reloading the Agent table every cycle, but it might still be
		//  worth it (20170801).
		for i, possibleAgent := range Agents {
			// check if we should be running or not
			if engineRunning.Load().(bool) {
				// let's mess this all up, shall we? If the user submits an agent, we'll simply use it instead
				userSetAgent := curAgent.Load().(string)
				// log.Println("userSetAgent is", userSetAgent)
				if userSetAgent != NullUUID {
					Agent = Agents[userSetAgent] // this is EVIL. EVIL!!! I love it (20170725)
					// log.Println("Agent got from user: ", *Agent.Name.Ptr())
				} else {
					Agent = possibleAgent	// we may skip one agent or two, but who cares?? Eventually we'll get back to
											//  that agent again, and we might do all this in parallel anyway (20170725)
				}
				
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
				
				Cubes = make(map[string]PositionType) // clear array, let the Go garbage collector deal with the memory (20170723)
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
						Cubes[*Position.UUID.Ptr()] = Position // if not a controller, it must be a cube! add it to array!
					}
				}
				
				// load in everything we found out so far on our region(s) but ignore phantom objects
				// end-users ought to set their cubes to phantom as well, or else the agents will think of them as obstacles!
				Obstacles = nil
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
					
					Obstacles = append(Obstacles, Object)
				}
				
				rows.Close()
								
				// Do not trust the database with the exact Agent position: ask the master controller directly
				
				//log.Println(masterController, Position, Cubes, Object, Obstacles)
				if !masterController.PermURL.Valid || !Agent.OwnerKey.Valid {
					log.Println(funcName() + ": Major error with database, we need at least one valid masterController and Agent to proceed. Sleeping for 10 seconds for user to correct this...")
					time.Sleep(10 * time.Second)
					break // go to next iteration, this one has borked data (20170801)
				}
				log.Println("master controller URL:", *masterController.PermURL.Ptr(), "Agent:", *Agent.Name.Ptr(), "Agent's OwnerKey:", *Agent.OwnerKey.Ptr())
				// WHY Agent.Ownerkey?!?! Why not Agent.UUID?!?!?
				curPos_raw, err := callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcGetPos")
				
				// NOTE(gwyneth): Apparently the web server will reply to ALL possible requests, even if the Agent doesn't exist any more;
				//  I still don't know what to do in that situation, so we skip this Agent and try the next one (20170730).
				if curPos_raw == "" || err != nil {
					break;
				}
				
				sendMessageToBrowser("status", "info", "Grid reports that agent " + *Agent.Name.Ptr() + " is at position: " + curPos_raw + "...</p>\n", "") 
				
				// update database with new position
				_, err = db.Exec("UPDATE Agents SET Position = '" + strings.Trim(curPos_raw, " ()<>") +
					"' WHERE OwnerKey = '" + *Agent.OwnerKey.Ptr() + "'")
				checkErr(err)
				
				db.Close()
				
				// sanitize
				Agent.Coords_xyz = strings.Split(strings.Trim(curPos_raw, " <>()\t\n\r"), ",")
				curPos := make([]float64, 3) // to be more similar to the PHP version
				_, err = fmt.Sscanf(curPos_raw, "%f,%f,%f", &curPos[0], &curPos[1], &curPos[2]) // best way to convert strings to floats! (20170728)
				checkErr(err)
				
				log.Println("Avatar", *Agent.Name.Ptr(), "(", *Agent.Name.Ptr(), ") is at recalculated vectorised position:", curPos)
				// calculate distances to nearest obstacles
					
				// TODO(gwyneth): these might become globals, outside the loop, so we don't need to declare them
				var smallestDistanceToObstacle = 1024.0 // will be used later on
				var nearestObstacle ObjectType
				var smallestDistanceToCube = 1024.0 // will be used later on
				var nearestCube PositionType
				obstaclePosition := make([]float64, 3)
				cubePosition := make([]float64, 3)
				var distance float64
				
				for k, point := range Obstacles {
					_, err = fmt.Sscanf(*point.Position.Ptr(), "%f, %f, %f", &obstaclePosition[0], &obstaclePosition[1], &obstaclePosition[2])
					checkErr(err)
					
					distance = calcDistance(curPos, obstaclePosition)
					
					fmt.Println("Obstacle", k, " - ", *point.Name.Ptr(), " - ", *point.Position.Ptr(), "- Distance:", distance)
					
					if distance < smallestDistanceToObstacle {
						smallestDistanceToObstacle = distance
						nearestObstacle = point
					}
				}
				sendMessageToBrowser("status", "info", fmt.Sprintf("Nearest obstacle to agent %s: '%s' (at %f)<br />", *Agent.Name.Ptr(), *nearestObstacle.Name.Ptr(), smallestDistanceToObstacle), "")				
								
				for k, point := range Cubes {
					_, err = fmt.Sscanf(*point.Position.Ptr(), "%f, %f, %f", &cubePosition[0], &cubePosition[1], &cubePosition[2])
					checkErr(err)
					
					distance = calcDistance(curPos, cubePosition)
					
					fmt.Println("Cube", k, " - ", *point.Name.Ptr(), " - ", *point.Position.Ptr(), "- Distance:", distance)
					
					if distance < smallestDistanceToCube {
						smallestDistanceToCube = distance
						nearestCube = point
					}
				}			
				sendMessageToBrowser("status", "info", fmt.Sprintf("Nearest cube to agent %s: '%s' (at %f)<br />", *Agent.Name.Ptr(), *nearestCube.Name.Ptr(), smallestDistanceToCube), "")				
				
				/* Idea for the GA
				
				1. Start with a 20x20 matrix (based loosely on Cosío and Castañeda) around the bot, which contain sensor data (we just sense up to 10 m around the bot). This might need adjustment (i.e. smaller size). 
				This represents the space of possible solutions
				Active cube will determine attraction point (see later)
				Chromosomes: randomly generated points (inside the 20x20 matrix) that the robot has to travel. Start perhaps with 50 with a length of 28 (Castañeda use 7 for 10x10 matrix). Points are bounded within the 20x20 matrix
				Now evaluate each chromosome with fitness function:
				- for each point: see if it's "too near" to an obstacle (potential collision)
					- ray casts are more precise, so give it a highest weight (not implemented yet)
					- normal sensor data give lower weigth
					- we can add modifiers: see number of prims of each obstacle (more prims, more weight, because object might be bigger than predicted); see if the obstacle is an agent (initially: agents might act as deflectors; later: interaction matrix will see if the bot goes nearer to the agent or avoids it)
				- for each point: see if it's closest to the cube. Lowest distance reduces weight. In theory, we wish to find the path with the least distance (less energy wasted)
				- sort chromosomes according to fitness
				- do 20 generations and find next expected point. Move bot to it. Reduce energy calculation on bot. See if it dies!
				- repeat for next bot position
				
				20130520 — Results don't converge. It's hard to get the 'bot in less than a 10m radius.
				Attempt #2 - use a 10x10 matrix, just 7 points, like Castañeda
				Gotshall & Rylander (2002) suggest a population size of about 100-200 for 7 chromosomes
				Attempt #3 - Algorithm from Ismail & Sheta was badly implemented!! 
				Attempt #4 - (to-do) implement Shi & Cui (2010) algorithm for fitness function
				Attempt #5 - Shi & Cui (2010) use a strange way to calculate path smoothness. Attempting Qu, Xing & Alexander (2013) which use angles. Modified mutation function, instead of the classical approach (switching two elements in the path), add random ±x, ±y to point
				André Neubauer (circular schema theorem, cited by Qu et al.) suggest two-point crossover
				Qu et al. suggest sorting path points, after crossover/mutation
				
				*/
				
				/*		goal/target/attractor: where the 'bot is going to go next
						at some point, this ought to be included in the chromosome as well
						for now, we'll hard-code it (walk to the nearest cube)
						on stage two, we'll do a simple check:
							- see what attributes are lowest
							- go to the nearest cube that replenishes the attribute
							- since this will be iterated every time the 'bot moves, we hope it won't die from starvation,
								as moving elsewhere becomes prioritary
				*/
						
				// nearestCube is where we go (20140526 changing it to selected cube by user, named destCube)
				var destCube PositionType
				
				// BUG(gwyneth): Somehow, the code below will just be valid once! (20170728)
				fmt.Println("User-set destination cube for", *Agent.Name.Ptr(), ":", userDestCube.Load().(string), "(NullUUID means no destination manually set)")
				if userDestCube.Load().(string) != NullUUID {
					destCube = Cubes[userDestCube.Load().(string)] 
					log.Println("User has supplied us with a destination cube for", *Agent.Name.Ptr(), "named:", *destCube.Name.Ptr())
				} else {
					destCube = nearestCube
					log.Println("Automatically selecting nearest cube for", *Agent.Name.Ptr(), "to go:", *destCube.Name.Ptr())
				}

				// This is just a test without the GA (20170725)
				// Commented out in 20170730 — forgot completely about this!!
				/*
				sendMessageToBrowser("status", "info", "GA will attempt to move agent '" + *Agent.Name.Ptr() + "' to cube '" + *destCube.Name.Ptr() + "' at position " + *destCube.Position.Ptr(), "")
				_, err = callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcMoveToTarget&vector=<" + *destCube.Position.Ptr() + ">&integer=1")
				checkErr(err)
				*/
				time_start := time.Now()
				
				// Genetic algorithm for movement
				// generate 50 strings (= individuals in the population) with 28 random points (= 1 chromosome) at curpos ± 10m
				
				population := make([]popType, POPULATION_SIZE) // create a population; unlike PHP, Go has to have a few clues about what is being created (20170726)
				// fmt.Println("population len", len(population))
				// initialise the slices of chromosomes; Go needs this to know how much memory to allocate (unlike PHP)
				for k := range population {
					population[k].chromosomes = make([]chromosomeType, CHROMOSOMES)
					//fmt.Println("Chromosome", k, "population len", len(population[k].chromosomes))
				}				
				
				// We calculate now the distance from each point to the destination
				//  Because this is computationally intensive, we will not repeat it every time during each generation
				//  Works well unless the destination moves! Then our calculations might be wrong
				//  But we will catch up on the _next_ iteration (hopefully, unless it moves too fast)
				//  We also use the best and second best path from a previous run of the GA		
			
				// get from the database the last two 'best paths' (if it makes sense)
				
				start_pop := 0 // if we have no best paths, we will generate everything from scratch
				
				// Maybe it makes sense to keep around the last best paths if we're still moving towards the same
				//  cube; so check for this first, and discard the last best paths if the destination changed
				
				// NOTE(gwyneth): Unlike the PHP version, the Go version deals simultaneously with an automated choice of path as well as manual
				//  setting of destination, through user input; so the code here is slightly different. We *already* have the
				//  destination cube in destCube (20170726).
				// We already have the cubePosition with the correct data (array of 3 float64 values for x,y,z).
				
				// NOTE(gwyneth): I have a doubt here, when the algorithm runs again, should the agent keep the CurrentTarget in mind? (20170726)
				//  The PHP code seems to assume that, but it wasn't ready yet for automated runs...
				
				// calculate the center point between current position and target
				// needs to be global for path sorting function (Ruhe's algorithm)
				// NOTE(gwyneth): in PHP we had a global $centerPoint; Go uses capital letters to designate globality (20170726).
				// NOTE(gwyneth): Ruhe's algorithm is not used any more, so we can safely forget this declaration (20170805).
				/*
				CenterPoint := struct {
					x, y, z float64
				}{
					x: 0.5 * (cubePosition[0] + curPos[0]),
					y: 0.5 * (cubePosition[1] + curPos[1]),
					z: 0.5 * (cubePosition[2] + curPos[2]),
				}
				
				sendMessageToBrowser("status", "", fmt.Sprintf("Center point for this iteration is <%f, %f, %f><br/>", CenterPoint.x, CenterPoint.y, CenterPoint.z), "")
				*/
				
				// Now generate from scratch the remaining population
							 
				for i := start_pop; i < POPULATION_SIZE; i++ {
					population[i].fitness = 0.0
		
					for y := 0; y < CHROMOSOMES; y++ {
						// Ismail & Sheta recommend to use the distance between points as part of the fitness
						// edge cases: first point, which is the distance to the current position of the agent
						// and last point, which is the distance between the last point and the target
						// that's why the first and last point have been inserted differently in the population
						// fmt.Println("i", i, "y", y)
						if y == 0 { // first point is (approx.) current position
							population[i].chromosomes[y].x = math.Trunc(curPos[0])
							population[i].chromosomes[y].y = math.Trunc(curPos[1])
							population[i].chromosomes[y].z = math.Trunc(curPos[2])
						} else if y == (CHROMOSOMES - 1) { // last point is (approx.) position of target
							population[i].chromosomes[y].x = math.Trunc(cubePosition[0])
							population[i].chromosomes[y].y = math.Trunc(cubePosition[1])
							population[i].chromosomes[y].z = math.Trunc(cubePosition[2])				
						} else { // others are scattered around the current position
							population[i].chromosomes[y].x = math.Trunc(curPos[0] + (rand.Float64() * 2)*RADIUS - RADIUS)			
							if population[i].chromosomes[y].x < 0.0 { 
								population[i].chromosomes[y].x = 0.0
							} else if population[i].chromosomes[y].x > 255.0 {
								population[i].chromosomes[y].x = 255.0
							}
							
							population[i].chromosomes[y].y = math.Trunc(curPos[1] + (rand.Float64() * 2)*RADIUS - RADIUS)
							if population[i].chromosomes[y].y < 0.0 {
								population[i].chromosomes[y].y = 0.0
							} else if population[i].chromosomes[y].y > 255.0 {
								population[i].chromosomes[y].y = 255.0
							}
							
							population[i].chromosomes[y].z = math.Trunc((cubePosition[2] + curPos[2])/2) // will work for flat terrain but not more
						}
						
						// To implement Shi & Cui (2010) or Qu et al. (2013) we add these distances to obstacles together
						//  If there are no obstacles in our radius, then we keep it clear
	
						population[i].chromosomes[y].obstacle = RADIUS // anything beyond that we don't care
						var point ObjectType
						for _, point = range Obstacles {
							_, err = fmt.Sscanf(*point.Position.Ptr(), "%f, %f, %f", &obstaclePosition[0], &obstaclePosition[1], &obstaclePosition[2])
							checkErr(err)
							distance = calcDistance([]float64 {population[i].chromosomes[y].x,
													population[i].chromosomes[y].y,
													population[i].chromosomes[y].z },
													obstaclePosition)

								
							// Shi & Cui and Qu et al. apparently just uses the distance to the nearest obstacle
							if distance < population[i].chromosomes[y].obstacle {
								 population[i].chromosomes[y].obstacle = 1/distance
								// we use the inverse here, because if we have many distant obstacles it's
								//  better than a single one that is close by
							}
							// TODO(gwyneth): obstacles flagged as ray-casting are far more precise, so they ought to be
							//  more weighted.
							
							// TODO(gwyneth): obstacles could also have bounding box calculations: bigger objects should
							//  be more weighted. However, HUGE objects might have holes in it. We ought to
							//  include the bounding box only for ray-casting, or else navigation would be impossible!
							//  Note that probably OpenSim raycasts only via bounding boxes (need confirmation)
							//  so maybe this is never a good approach. Lots of tests to be done here!
							// NOTE(gwyneth): The latest version of llRayCast, v3 on BulletSim, does NOT use bounding boxes. Confirmed 20170730.
						}
						if RADIUS - population[i].chromosomes[y].obstacle < 0.00001 { // we use a delta to deal with rounding errors with floats
							population[i].chromosomes[y].obstacle = 0.0
						}
						
						// calculate, for this point, its distance to the destination, currently $destCube
						// (exploded to array $cubePosition)
						// might not need this 
						
						population[i].chromosomes[y].distance = calcDistance([]float64 { population[i].chromosomes[y].x,
													population[i].chromosomes[y].y,
													population[i].chromosomes[y].z},
													cubePosition)						
							
						// abandoned: initialize smoothness for Shi & Cui
						// adopted (20140523): smoothness using angles, like Qu et al.
						population[i].chromosomes[y].smoothness = 0.0
						population[i].chromosomes[y].angle = 0.0 // maybe initialize it here
					} // endfor y
					
					// now sort this path. According to Qu et al. (2013) this gets us a smoother path
					// hope it's true, because it's computationally intensive
					// (20140523) we're using Ruhe's algorithm for sorting according to angle, hope it works
					// (20140524) Abandoned Ruhe, using a simple comparison like Qu et al.
					//echo "Before sorting, point i is: "; var_dump(population[i]); echo "<br />\n";
					/*
					$popsort = substr(population[i], 1, -1);
					$first_individual = population[i][0];
					$last_individual  = population[i][CHROMOSOMES - 1];
					usort($popsort, 'point_cmp');
					population[i] = 	array_merge((array)$first_individual, $popsort, (array)$last_individual);
					*/
					
					pop := population[i].chromosomes
					
					sort.Slice(pop, func(a, b int) bool {
						// NOTE(gwyneth): these is still the old PHP comments, kept here for historical reasons
						// global $centerPoint;
						
					
						/* Attempt #1: Ruhe's algorithm
					
						$theta_a = atan2($a.y - $centerPoint.y, $a.x - $centerPoint.x);
					    $angle_a = fmod(M_PI - M_PI_4 + $theta_a, 2 * M_PI);
					
						$theta_b = atan2($b.y - $centerPoint.y, $b.x - $centerPoint.x);
					    $angle_b = fmod(M_PI - M_PI_4 + $theta_a, 2 * M_PI);
					
						if ($angle_a == $angle_b)
							return 0;
					
						return ($angle_a < $angle_b) ? -1 : 1;
						*/
					
						/*
						// Attempt #2: just angles
					
						if ($a.angle == $b.angle)
							return 0;
					
						return (abs($a.angle) < abs($b.angle)) ? -1 : 1;
						*/
						// Attempt #3: Just compare x,y! This is a terrible solution but better than nothing
						// using algorithm from Anonymous on http://www.php.net/manual/en/function.usort.php
						/*
						if ($a.x == $b.x)
						{
							if ($a.y == $b.y)
							{
								return 0;
							}
							elseif ($a.y > $b.y)
							{
								return 1;
							}
							elseif ($a.y < $b.y)
							{
								return -1;
							}
						}
						elseif ($a.x > $b.x)
						{
							return 1;
						}
						elseif ($a.x < $b.x)
						{
							return -1;
						}
						*/
						
						// Attempt #4: order by shortest distance to the target?
						
						return pop[a].distance < pop[b].distance
					})
				} // endfor i
				
				// testing printing the current population (with json we get strange results!)
				showPopulation(population)
				
				//marshalled, err := json.MarshalIndent(population, "", "  ") // debug line just to show population's structure
				//checkErr(err)
				//log.Println("Population", string(marshalled))
				
				// We're not finished yet! We need to calculate angles between all points (duh!) to establish smoothness
				// Let's do it from scratch:		
				
				for i := 0; i < POPULATION_SIZE; i++ {
					population[i].chromosomes[0].angle = 0.0 // curPos has (obviously) angle 0
					
					for j := 1; j < CHROMOSOMES; j++ {
						population[i].chromosomes[j].angle = math.Abs(
							math.Atan2(population[i].chromosomes[j].y - population[i].chromosomes[j-1].y, 
							population[i].chromosomes[j].x - population[i].chromosomes[j-1].x))
					}
				}

			// Initial population done; now loop over generations
			
			for generation := 0; generation < GENERATIONS; generation++	{
				// Calculate fitness
				log.Println("Generating fitness for generation ", generation, " (out of ", GENERATIONS, ") for agent", *Agent.Name.Ptr(), "...")
				
				// When calculating a new population, each element will have its chromosomes reordered
				//  So we have no choice but to calculate fitness for all population elements _again_
				for i := 0; i < POPULATION_SIZE; i++	{
					fitnessW1 := 0.0
					fitnessW2 := 0.0
					fitnessW3 := 0.0
					
					// note that first point is current location; we start from the second point onwards
					for y := 1; y < CHROMOSOMES; y++ {
						// Sub-function of Path Length (using Shi & Cui)
						
						distLastPoint := calcDistance([]float64 {
									population[i].chromosomes[y].x,
									population[i].chromosomes[y].y,
									population[i].chromosomes[y].z,
								},
								[]float64 {
									population[i].chromosomes[y-1].x,
									population[i].chromosomes[y-1].y,
									population[i].chromosomes[y-1].z,
								})
								
						// Eduardo: suggests using square distance, means path will have more
						//  distributed points. (20140704 - 2004)
												
						fitnessW1 += distLastPoint * distLastPoint
						
						// Sub-function of Path Security (using Shi & Cui) — obstacle proximity
						
						fitnessW2 += population[i].chromosomes[y].obstacle
						
						// Sub-function of Smoothness (using Shi & Cui)
						// This measures how zig-zaggy the path is, namely, if points are pointing back etc.
						// We want a smooth path towards the goal
						// Possibly here is where the weight will be added
						// Attempt #5: Qu et al. suggest the angle between line segments
						// used http://stackoverflow.com/questions/20395547/sorting-an-array-of-x-and-y-vertice-points-ios-objective-c
						
						/* Shi & Cui; abandoned
						population[i][$y]["smoothness"] =
							(
								(population[i][$y-1].y - population[i][$y].y) /
								(population[i][$y-1].x - population[i][$y].x)
							)
							-
							(
								(population[i][$y].y - population[i][$y-1].y) /
								(population[i][$y].x - population[i][$y-1].x)
							);
						*/
						
						fitnessW3 += math.Abs(population[i].chromosomes[y].angle) // clever, huh? check if abs makes sense
						
						// and we'll also use the overall distance to the attractor
						//population[i]["fitness"] += population[i][$y]["distance"];
					} // end for y
					population[i].fitness = W1 * fitnessW1 + W2 * fitnessW2 + W3 * fitnessW3					
				} // end for i
				
				// note that the most critical point is the first: it's the one the 'bot will try to walk to. But we need
				//  to calculate the rest of the path, too, which is "the best path so far which the bot plans to travel"
				//  even if at every iteration, it will get calculated over and over again
				
				/*
				$time_end = microtime(true);
				$time = $time_end - $time_start;
		
				echo "<p>CPU time used after fitness calculations for generation " . $generation . ": " . $time
					. " seconds</p>\n";
				
				
				echo "Generation " . $generation . " - Before ordering: <br />\n";
				showPopulation(population);
				echo "<br />\n";
				*/

				// Now we do genetics!
		
				// To pick the 'best' population elements, we need to sort this by fitness, so that the best
				//  elements are at the top
				
				// order by fitness
				sort.Slice(population, func(a, b int) bool {
						return population[a].fitness < population[b].fitness
					})
				
				// TODO(gwyneth): to comment out later (20170727)
				//log.Println("Generation ", generation, " - After ordering by fitness:")
				//showPopulation(population);
				
				time_end := time.Now()
				diffTime := time_end.Sub(time_start)
				log.Println("CPU time used after sorting this generation for agent", *Agent.Name.Ptr(), ":", diffTime)		

				// Selection step. We're using fitness rank
				
				newPopulation := make([]popType, POPULATION_SIZE) // create a new population; see comments above
				for k := range newPopulation {
					newPopulation[k].chromosomes = make([]chromosomeType, CHROMOSOMES)
				}					
				
				// To introduce elitism, we will move the first 2 elements to the new population:
				newPopulation[0] = population[0]
				newPopulation[1] = population[1]
				// we could also delete the remaining two
						
				// for the remaining population:
				for i := 2; i < POPULATION_SIZE; i += 2 {
					// establish if we do crossover
					if rand.Float64() * 100 < CROSSOVER_RATE {		
						// find a crossover point; according to André Neubauer, we might need two crossover points
						crossover_point := int(math.Trunc(rand.Float64() * CHROMOSOMES))
						child0 := make([]chromosomeType, CHROMOSOMES)
						child1 := make([]chromosomeType, CHROMOSOMES)
					
						// echo "Generation " . $generation . " - Crossover for " . i . " and " . (i + 1) . " happening at crossover point: " . $crossover_point . "</br>\n";
						
						// now copy the chromosomes from the first parent, up to the crossover point, to child0
						//  and the remaining chromosomes go to the second child
						// simultaneously, do the reverse for the second parent
						
						// there are probably better/faster string manipulation techniques but this is easy to debug
						for chromosome := 0; chromosome < CHROMOSOMES; chromosome++ {
							if chromosome <= crossover_point {
								child0[chromosome] = population[i].chromosomes[chromosome]
								child1[chromosome] = population[i+1].chromosomes[chromosome]
							} else {
								child0[chromosome] = population[i+1].chromosomes[chromosome]
								child1[chromosome] = population[i].chromosomes[chromosome]
							}
							/* 
							echo "Pop " . i . ", chromosome: " . $chromosome . " Original chromosome: ";
							print_r(population[i][$chromosome]);
							echo " Child 0 chromosome: ";
							print_r(child0[$chromosome]);
							echo "<br />\n";
							*/
						} // endif chromosomes
						
						// test for mutation; note that this is permille and not percent
						if rand.Float64() * 1000 < MUTATION_RATE {
							/*
								Abandoned mutation implementation, which was a classical formulation
								for value-based GA (as opposed to bit-based)
								
							// pick two chromosomes for first child, two for second child
							//  see http://obitko.com/tutorials/genetic-algorithms/crossover-mutation.php
							
							first_chromosome = mt_rand(0, CHROMOSOMES-1);
							second_chromosome = mt_rand(0, CHROMOSOMES-1);
							// exchange them, by using a temporary holder (this is mostly because
							//  the exchange might be for the same chromosome!
							$temp_first_chromosome = child0[first_chromosome];
							$temp_second_chromosome = child0[second_chromosome];
							
							child0[first_chromosome] = $temp_second_chromosome;
							child0[second_chromosome] = $temp_first_chromosome;
	
							// echo "Generation " . $generation . " - Mutation happening for child " . i . " — exchanging chromosomes " . first_chromosome . " and " . second_chromosome . "</br>\n";
							
							// same for second child
	
							first_chromosome = mt_rand(0, CHROMOSOMES-1);
							second_chromosome = mt_rand(0, CHROMOSOMES-1);
							// exchange them, by using a temporary holder (this is mostly because
							//  the exchange might be for the same chromosome!
							$temp_first_chromosome = child1[first_chromosome];
							$temp_second_chromosome = child1[second_chromosome];
							
							child1[first_chromosome] = $temp_second_chromosome;
							child1[second_chromosome] = $temp_first_chromosome;
	
							// echo "Generation " . $generation . " - Mutation happening for child " . (i + 1) . " — exchanging chromosomes " . first_chromosome . " and " . second_chromosome . "</br>\n";
							
							*/
							
							/*
								(20140523) Instead, as Qu et al. do, just pick a point and add some 
								random distance to it
							*/
							first_chromosome  := int(math.Trunc(rand.Float64() * (CHROMOSOMES-1)))
							second_chromosome := int(math.Trunc(rand.Float64() * (CHROMOSOMES-1)))
														
							child0[first_chromosome].x += (rand.Float64() * 2)*RADIUS/2 - RADIUS/2
							if child0[first_chromosome].x < 0 {
								child0[first_chromosome].x = 0
							} else if child0[first_chromosome].x > 255 {
								child0[first_chromosome].x = 255
							}
							child0[first_chromosome].y += (rand.Float64() * 2)*RADIUS/2 - RADIUS/2
							if child0[first_chromosome].y < 0 {
								child0[first_chromosome].y = 0
							} else if child0[first_chromosome].y > 255 {
								child0[first_chromosome].y = 255
							}
							child1[second_chromosome].x += (rand.Float64() * 2)*RADIUS/2 - RADIUS/2
							if child1[second_chromosome].x < 0 {
								child1[second_chromosome].x = 0
							} else if child1[second_chromosome].x > 255 {
								child1[second_chromosome].x = 255
							}
							child1[second_chromosome].y += (rand.Float64() * 2)*RADIUS/2 - RADIUS/2
							if child1[second_chromosome].y < 0 {
								child1[second_chromosome].y = 0
							} else if child1[second_chromosome].y > 255 {
								child1[second_chromosome].y = 255
							}
						} // endif mutation	
						
						/*						
						echo "Generation " . $generation . " - New children for population " . i . ": ";
						print_r(child0);
						echo " and " . (i + 1) . ": ";
						print_r(child1);
						echo "</br>\n";
						*/
						
						/* we need to sort the points on the two childs AGAIN. Duh. And recalculate the angles.
							Duh, duh, duh */
						/*
						child0sort = substr(child0, 1, -1);
						child1sort = substr(child1, 1, -1);
						
						$first_individual_child0 = child0[0];
						$first_individual_child1 = child1[0];
						
						$last_individual_child0 = child0[CHROMOSOMES - 1];
						$last_individual_child1 = child1[CHROMOSOMES - 1];
						
						usort(child0sort, 'point_cmp');
						usort(child1sort, 'point_cmp');
						
						child0 = array_merge((array)$first_individual_child0, child0sort, 
							(array)$last_individual_child0);
						child1 = array_merge((array)$first_individual_child1, child1sort, 
							(array)$last_individual_child1);
						*/
						
						sort.Slice(child0, func(a, b int) bool {
							return child0[a].distance < child0[b].distance
							})
						sort.Slice(child1, func(a, b int) bool {
							return child1[a].distance < child1[b].distance
							})
						
						child0[0].angle = 0.0;
						child0[1].angle = 0.0;
			
						for j := 1; j < CHROMOSOMES; j++ {
							child0[j].angle = math.Atan2(child0[j].y - child0[j-1].y, 
									child0[j].x - child0[j-1].x)
							child1[j].angle = math.Atan2(child1[j].y - child1[j-1].y, 
									child1[j].x - child1[j-1].x)
						}
						// add the two children to the new population; fitness will be calculated on next iteration
						newPopulation[i].chromosomes	= child0
						newPopulation[i+1].chromosomes	= child1
						
						// endif crossover
					} else {
						// no crossover, just move them directly
						newPopulation[i].chromosomes	= population[i].chromosomes
						newPopulation[i+1].chromosomes	= population[i+1].chromosomes
						// echo "No crossover for " . i . " and " . (i + 1) . " - moving parents to new population<br />\n"; 
					}
					// echo "<hr />\n<p>Pop " . i . " finished</p><hr />\n";
					
				}
	/*
				echo "<hr />\n<p>Generation " . $generation . " finished</p><hr />\n";
					
				population = newPopulation; // prepare population
					
				$time_end = microtime(true);
				$time = $time_end - $time_start;
		
				echo "<p>CPU time used after crossover and mutation up to generation " . $generation  . ": "  . $time ." seconds</p>\n<hr />\n";
	*/
			}  // for generation
	
			//fmt.Println("Final result (", GENERATIONS, " generation(s)):")
			//showPopulation(population)

				// at the end, the first point (after the current position) for the last population should give us the nearest point to move to
			//  ideally, the remaining points should also have converged
			//  obviously, as the avatar moves and finds about new obstacles etc. the population will change
	 		
	
			// move to target; integer=1 means "walking" (never "flying")
			// 
			
			// Solution by Eduardo 20140704 — if we're close to the destination, within its radius, then we should
			//  move to the last point — which is our current destination!
			
			// Calculate where we are before we move
			distanceToTarget := calcDistance(curPos, cubePosition)
			
			var target int // declared here for scope issues (PHP has the ternary operator for dealing with that, Go hasn't)
			
			if distanceToTarget < RADIUS {
				target = CHROMOSOMES
			} else {
				target = CHROMOSOMES -1
			}
			
			sendMessageToBrowser("status", "info", fmt.Sprintf("Solution: move agent %s first to (%v, %v, %v) [and follow with %v points]. Distance is %v m", *Agent.Name.Ptr(), population[0].chromosomes[1].x, population[0].chromosomes[1].y, population[0].chromosomes[1].z, target - 1, distanceToTarget), "")
			
			// BUG(gwyneth): Major bug here! Possibly corrected with new approach. (20170805)
			// Basically, we generate the path for the next CHROMOSOME points. But we just need to move the avatar to the NEXT point
			// (i.e. chromosomes[1], because we will recalculate the whole path from then on. On the other hand, we need to print out
			// what the best path was so far, etc. and this will enter the calculations for the next batch of generations and so forth.
			// Nevertheless, the command to move the avatar is just for the NEXT point. (20170805) This should eliminate the 'quirks' of
			// the avatar moving in zig-zag and sometimes even backwards... and hopefully it will avoid them to move towards 0,0,Z...
			
			// Possibly Go is much faster at calculating new paths than the avatar is in moving to the next one.

			// because Go is so fast at calculating generations, we need to push the commands to give on a separate goroutine
			//  which acts as a worker to consume points and wait on them until the avatar has finished walking to the point.
			// We begin with a channel with the capacity of allowing CHROMOSOMES points, maybe this needs to be adjusted in the future
			// Possibly we just need to move to the NEXT point, so the channel capacity would be just one! (20170805)
			movementJobChannel <- movementJob {
						agentUUID: *Agent.OwnerKey.Ptr(),
						masterControllerPermURL: *masterController.PermURL.Ptr(),
						destPoint: population[0].chromosomes[1],
			}
			
			for p := 1; p < target; p++ { // (skip first point — current location)
				// This is added for the CSV export, but I don't know if it makes more sense here or in the movementWorker goroutine
				export_rows = append(export_rows, fmt.Sprintf("%f,%f,%f;", population[0].chromosomes[p].x, population[0].chromosomes[p].y, population[0].chromosomes[p].z))
			}
						
	 		// Save two best solutions for next iteration; attempts to avoid to recalculate always from scratch
	 		// We do it after moving because the avatar needs a few seconds to reach destination 		
	 		
			// Reopen database, we need to write out the new Agent data
			db, err := sql.Open(PDO_Prefix, GoBotDSN)
			checkErr(err)
			
	 		// now update our database with the best paths and the target
			stmt, err := db.Prepare("UPDATE Agents SET BestPath=?, SecondBestPath=?, CurrentTarget=? WHERE UUID=?")
			if err != nil {
				sendMessageToBrowser("status", "error", fmt.Sprintf("%v: Updating database with best path, second best path, and current target for agent %s - prepare failed: %s", 
					funcName(), *Agent.Name.Ptr(), err), "")
			}

			marshalled0, err := json.Marshal(population[0])
			checkErr(err)
			marshalled1, err := json.Marshal(population[1])
			checkErr(err)

			_, err = stmt.Exec(marshalled0,
						marshalled1,
						strings.Trim(*destCube.Position.Ptr(), " \t"),
						*Agent.UUID.Ptr())
			if err != nil {
				sendMessageToBrowser("status", "error", fmt.Sprintf("%v: Updating database with best path, second best path, and current target for agent %s failed: %s", 
					funcName(), *Agent.Name.Ptr(), err), "")
			}
			
			stmt.Close()
			db.Close()

			// See if we're close to the target; absolute precision might be impossible
			// first, read position again, just to see what we get
			curposResult, err := callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcGetPos")
			if err == nil {
				sendMessageToBrowser("status", "info", fmt.Sprintf("Grid reports that agent %s is at position: %v", 
					*Agent.Name.Ptr(), curposResult), "")
			} else { 
				sendMessageToBrowser("status", "error", "We cannot get report from the grid for the position of agent " + *Agent.Name.Ptr(), "") 
			}
	/*
			echo "This is how Destination Cube looks like: <br \>\n";
			var_dump($destCube);
			echo "<br />\n";
	*/		
			currentPosition := make([]float64, 3) // see comment above for doing the same to curPos (20170728)
			
			_, err = fmt.Sscanf(strings.Trim(curposResult, " ()<>"), "%f,%f,%f", &currentPosition[0], &currentPosition[1], &currentPosition[2])
			checkErr(err)			
	
	/*
			echo "Distance comparison: Current Position:<br />\n";
			var_dump($currentPosition);
			echo "Target Cube<br />\n";
			var_dump($targetCube);
			echo "<br />\n";
	*/		
			distance = calcDistance(cubePosition, currentPosition);
							
			if distance < 1.1 { // we might never get closer than this due to rounding errors
				sendMessageToBrowser("status", "info", fmt.Sprintf("Within rounding errors of %s , distance is merely %v m; let's sit %s down", *destCube.Name.Ptr(), distance, *Agent.Name.Ptr()), "")

				// if we're close enough, sit on it
				sitResult, err := callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcSit&key=" + *destCube.UUID.Ptr() + "&integer=" + OS_NPC_SIT_NOW)
				if (err == nil) {
					sendMessageToBrowser("status", "info", "Result from " + *Agent.Name.Ptr() + " sitting: " + sitResult, "")
				} else {
					sendMessageToBrowser("status", "error", "Grid error when trying to sit " + *Agent.Name.Ptr(), "")					
				}
			} else if distance < 2.5 {
				sendMessageToBrowser("status", "warning", fmt.Sprintf("%s is very close to %s, distance is now %v m", *Agent.Name.Ptr(), *destCube.Name.Ptr(), distance), "")
			} else {
				sendMessageToBrowser("status", "warning", fmt.Sprintf("%s is still %v m away from %s (%v, %v, %v)",
					*Agent.Name.Ptr(),
					distance,
					*destCube.Name.Ptr(),
					cubePosition[0],
					cubePosition[1],
					cubePosition[2]), "")
			}
			
			// Now place a button to save/export to CSV or XML
			// TODO(gwyneth): this was on the original code but no button was there; need to see where it is (20170728)
				
				// If the user had set agent + cube, clean them up for now
				//  They never get cleaned! That's the whole point! Unless of course the user WANTS them cleaned (20170729)
				//userDestCube.Store(NullUUID)
				//curAgent.Store(NullUUID)
				
				time_end := time.Now()
				diffTime := time_end.Sub(time_start)
				log.Println("CPU time used", diffTime)
				sendMessageToBrowser("status", "info", fmt.Sprintf("CPU time used: %v", diffTime), "")
				
				// output something to console so that we know this is being run in parallel
				/*
			    fmt.Print("\r|")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r/")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r-")
			    time.Sleep(1000 * time.Millisecond)
			    fmt.Print("\r\\")
			    time.Sleep(1000 * time.Millisecond)
			    */
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
	sendMessageToBrowser("status", "success", "this is the engine <i>stopping</i><br />", "")
}

// sendMessageToBrowser sends a string to the internal, global channel which is picked up by the websocket handling goroutine.
// In the case of special status messages (info, success, warning, error) we also send the same message to the log.
// If no WebSocket is active (and we check that in two different ways!) the message simply goes to the log instead.
func sendMessageToBrowser(msgType string, msgSubType string, msgText string, msgId string) {
	if webSocketActive.Load().(bool) == true { // no point in sending if nobody is there to receive
		var msgToSend WsMessageType
		
		msgToSend.New(msgType, msgSubType, msgText, msgId)
		
		// Go idiomatic programming: 'select' parallels the output of the two cases and picks the one which finishes; in this case, either
		// we are able to send a message via the channel, or there is a timeout, and Go picks what happens first
		select {
		    case wsSendMessage <- msgToSend:
				// we use this so often as info/warning/error message that we may better send it also to the log
				if msgType == "status" && msgSubType != "" {
					log.Println("(connected via WebSocket)", msgType, "-", msgSubType, "-", msgText, "-", msgId)
				}
				// 'common' messages have the nil string subtype, so we ignore these and don't log them
				//  we might have a Debug facility in the future which allows for more verbosity!
 		    case <-time.After(time.Second * 10):
 		    	// this case exists only if we failed to figure out if the WebSocket is active or not; in most cases, we will
 		    	//  be able to know that in advance, but here we catch the edge cases.
		        log.Println("WebSocket timeout after 10 seconds; coudn't send message:", msgType, "-", msgSubType, "-", msgText, "-", msgId)
		}
	} else {
		// No active WebSocket? Just dump it to the log. Note that this will be the most usual case, since we hardly expect users to be 24/7 in
		//  front of their browsers...
		log.Print("(no WebSocket connection)", msgType, "-", msgSubType, "-", msgText, "-", msgId)
	}
}

// callURL encapsulates a call to an URL. It exists as an analogy to the PHP version (20170723).
func callURL(url string, encodedRequest string) (string, error) {
	//  HTTP request as per http://moazzam-khan.com/blog/golang-make-http-requests/
	body := []byte(encodedRequest)
    
    rs, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(body))
    // Code to process response (written in Get request snippet) goes here
	
	if err == nil {
		defer rs.Body.Close()
		
		rsBody, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			errMsg := fmt.Sprintf("Error response from in-world object: '%v'", err)
			log.Println(errMsg)
			return errMsg, err
		} else {
		    // log.Printf("Reply from in-world object %s\n", rsBody)
			return string(rsBody), err
		}
	} else {
		log.Printf("HTTP call to %s failed; error was: '%v'", url, err)
		return fmt.Sprintf("HTTP call to %s failed; error was: '%v'", url, err), err
	}
}

// showPopulation is adapted from the PHP code to pretty-print a whole population
func showPopulation(popul []popType) {
fmt.Println("Population")
	for p, pop := range popul {
		fmt.Print("pop:", p, " - Fitness: ", pop.fitness, " Chromosomes: ")					
		for c, chr := range pop.chromosomes {
			fmt.Printf("[%v] { %v, %v, %v } Distance: %v Obstacle: %v Angle: %v Smoothness %v;\t",
				c, chr.x, chr.y, chr.z, chr.distance, chr.obstacle, chr.angle, chr.smoothness)
		}
		fmt.Println("\n---")
	}	
}

// movementWorker reads one point from the movementJobChannel and sends a command to the avatar to move to it.
// Then it blocks for the necessary amount of estimated time for the avatar to reach that destination until it reads the next
// point. Right now, the channel accepts CHROMOSOME jobs at a time (20170730).
func movementWorker() {
	var nextPoint movementJob
	curPos := make([]float64, 3) // no need to allocate this over and over again
	for { // once started, never stops?
		nextPoint = <-movementJobChannel // consume one point from the channel
	
		// these must be set, nothing like a bit of error checking for eventual bugs elsewhere
		if nextPoint.masterControllerPermURL == "" || nextPoint.agentUUID == "" {
			break
		}
	
		// The code below was commented in the PHP code, but we reuse it here as it was because it makes sense! (20170730)	

		// How much should we wait? Well, we calculate the distance to the next point! And since avatars move pretty much
		//  at the same speed all the time, we can make a rough estimate of the time they will take.
		
		// ask the avatar where he is
		curposResult, err := callURL(nextPoint.masterControllerPermURL, "npc=" + nextPoint.agentUUID + "&command=osNpcGetPos")
		checkErr(err)
		// convert string result to array of floats				
		_, err = fmt.Sscanf(strings.Trim(curposResult, " ()<>"), "%f,%f,%f", &curPos[0], &curPos[1], &curPos[2])
		checkErr(err)
		
		walkingDistance := calcDistance(
			curPos, []float64 {	nextPoint.destPoint.x, nextPoint.destPoint.y, nextPoint.destPoint.z })
		timeToTravel := walkingDistance / WALKING_SPEED // we might adjust this to assume the in-world calls took some time as well
		// do not wait too long, though!
		if timeToTravel > 5.0 {
			timeToTravel = 5.0
		}
		sendMessageToBrowser("status", "", fmt.Sprintf("[%s]: Next point at %v metres; waiting %v secs for avatar %s to go to next point...<br />",
			funcName(), walkingDistance, timeToTravel, nextPoint.agentUUID), "")
	
		moveResult, err := callURL(nextPoint.masterControllerPermURL, 
					fmt.Sprintf("npc=%s&command=osNpcMoveToTarget&vector=<%v,%v,%v>&integer=1",
					nextPoint.agentUUID, nextPoint.destPoint.x, nextPoint.destPoint.y, nextPoint.destPoint.z))
		checkErr(err)
				
		sendMessageToBrowser("status", "", fmt.Sprintf("[%s]: In-world result call from moving %s to (%v, %v, %v): %s<br />",
			funcName(), nextPoint.agentUUID, nextPoint.destPoint.x, nextPoint.destPoint.y, nextPoint.destPoint.z, moveResult), "")
		time.Sleep(time.Second * time.Duration(timeToTravel))
	}
}