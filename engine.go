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
	//"sort"
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
					userDestCube.Store(returnValues[0])
					curAgent.Store(returnValues[1])
					
					log.Println("Destination: ", userDestCube.Load().(string), "Agent:", curAgent.Load().(string))
					sendMessageToBrowser("status", "info", "Received '" + userDestCube.Load().(string) + "|" + curAgent.Load().(string) + "'<br />", "")
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
		Agents map[string]AgentType // we OUGHT to have a type without those strange zero.String, but it's tough to keep two structs in perfect sync (20170722); this is mapped by Agent UUID (20170725)
		Position PositionType
		Cubes map[string]PositionType // name to be compatible with PHP version; mapped by UUID (20170725)
		Object ObjectType
		Obstacles []ObjectType
		masterController PositionType // we will need the most recent Bot Master Controller to send commands! (name is the same as in former PHP code).
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
	if len(Agents) == 0 {
		log.Println("Error: no Agents found. Engine cannot run. Aborted. Add an Agent and try sending a SIGCONT to restart engine again")
		sendMessageToBrowser("status", "restart", "Error: no Agents found. Engine cannot run. Aborted. Add an Agent and try sending a <code>SIGCONT</code> to restart engine again<br />"," ")
		return
	}
			
	for {
		for i, possibleAgent := range Agents {
			// check if we should be running or not
			if engineRunning.Load().(bool) {
				// let's mess this all up, shall we? If the user submits an agent, we'll simply use it instead
				userSetAgent := curAgent.Load().(string)
				// log.Println("userSetAgent is", userSetAgent)
				if userSetAgent != NullUUID {
					Agent = Agents[userSetAgent] // this is EVIL. EVIL!!! I love it (20170725)
					log.Println("Agent got from user: ", Agent)
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
				// log.Println(*masterController.PermURL.Ptr())
				
				//log.Println(masterController, Position, Cubes, Object, Obstacles)
				curPos_raw, _ := callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcGetPos")
				sendMessageToBrowser("status", "info", "<p class='box'>Grid reports that agent " + *Agent.Name.Ptr() + " is at position: " + curPos_raw + "...</p>\n", "") 
				log.Println("Grid reports that agent", *Agent.Name.Ptr(), "is at position:", curPos_raw, "...")
				
				// update database with new position
				_, err = db.Exec("UPDATE Agents SET Position = '" + strings.Trim(curPos_raw, " ()<>") +
					"' WHERE OwnerKey = '" + *Agent.OwnerKey.Ptr() + "'")
				checkErr(err)
				
				db.Close()
				
				// sanitize
				Agent.Coords_xyz = strings.Split(strings.Trim(curPos_raw, " <>()\t\n\r"), ",")
				curPos := make([]float64, 3) // to be more similar to the PHP version
				_, err = fmt.Sscanf(curPos_raw, "<%f, %f, %f>", &curPos[0], &curPos[1], &curPos[2])
				checkErr(err)
				
				log.Println("Avatar", *Agent.Name.Ptr(), "is at recalculated vectorised position:", curPos)
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
				statusMessage := fmt.Sprintf("Nearest obstacle: '%s' (at %f)", *nearestObstacle.Name.Ptr(), smallestDistanceToObstacle)
				fmt.Println(statusMessage)
				sendMessageToBrowser("status", "info", statusMessage + "<br />", "")				
								
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
				statusMessage = fmt.Sprintf("Nearest cube: '%s' (at %f)", *nearestCube.Name.Ptr(), smallestDistanceToCube)
				fmt.Println(statusMessage)
				sendMessageToBrowser("status", "info", statusMessage + "<br />", "")	
				
				
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
				
				if userDestCube.Load().(string) != NullUUID {
					destCube = Cubes[userDestCube.Load().(string)]
					log.Println("User has supplied us with a destination cube named:", *destCube.Name.Ptr())
				} else {
					destCube = nearestCube
					log.Println("Automatically selecting nearest cube to go:", *destCube.Name.Ptr())
				}

				// This is just a test without the GA (20170725)
				sendMessageToBrowser("status", "info", "GA will attempt to move agent '" + *Agent.Name.Ptr() + "' to cube '" + *destCube.Name.Ptr() + "' at position " + *destCube.Position.Ptr(), "")
				_, _ = callURL(*masterController.PermURL.Ptr(), "npc=" + *Agent.OwnerKey.Ptr() + "&command=osNpcMoveToTarget&vector=<" + *destCube.Position.Ptr() + ">&integer=1")
				
				time_start := time.Now()
				
				// Genetic algorithm for movement
				// generate 50 strings (= individuals in the population) with 28 random points (= 1 chromosome) at curpos ± 10m
				
				// When transposing from the PHP version, we now cannot avoid having a few structs and types, since Go
				//  is a strongly-typed language (20170726)
				
				// chromosomeType is just a point in a path, really.
				type chromosomeType struct {
					x, y, z, distance, obstacle, angle, smoothness float64
				}
				
				// popType represents each population as a list of points (= chromosomes) indicating a possible path; it also includes the fitness for this particular path.
				type popType struct {
					fitness float64
					chromosomes [CHROMOSOMES]chromosomeType
				}
				
				population := make([]popType, POPULATION_SIZE) // create a population; unlike PHP, Go has to have a few clues about what is being created (20170726)
				
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
				CenterPoint := struct {
					x, y, z float64
				}{
					x: 0.5 * (cubePosition[0] + curPos[0]),
					y: 0.5 * (cubePosition[1] + curPos[1]),
					z: 0.5 * (cubePosition[2] + curPos[2]),
				}
				
				// Now generate from scratch the remaining population
							 
				for i := start_pop; i < POPULATION_SIZE; i++ {
					population[i].fitness = 0.0
		
					for y := 0; y < CHROMOSOMES; y++ {
						// Ismail & Sheta recommend to use the distance between points as part of the fitness
						// edge cases: first point, which is the distance to the current position of the agent
						// and last point, which is the distance between the last point and the target
						// that's why the first and last point have been inserted differently in the population
					
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
							
							population[i].chromosomes[y].z = math.Trunc(CenterPoint.z) // will work for flat terrain but not more
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
						}
						if (population[i].chromosomes[y].obstacle == RADIUS) {// we might have to use a delta here, because of rounding errors
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
					//echo "Before sorting, point $i is: "; var_dump($population[$i]); echo "<br />\n";
					/*
					$popsort = substr($population[$i], 1, -1);
					$first_individual = $population[$i][0];
					$last_individual  = $population[$i][CHROMOSOMES - 1];
					usort($popsort, 'point_cmp');
					$population[$i] = 	array_merge((array)$first_individual, $popsort, (array)$last_individual);
					*/
					
					/********* DEAL WITH SORT *****/
					
					//sort.Slice(population[i], func(a, b int) bool {
						// NOTE(gwyneth): this is still the old PHP code, kept here for historical reasons
						// global $centerPoint;
						
					
						/* Attempt #1: Ruhe's algorithm
					
						$theta_a = atan2($a["y"] - $centerPoint["y"], $a["x"] - $centerPoint["x"]);
					    $angle_a = fmod(M_PI - M_PI_4 + $theta_a, 2 * M_PI);
					
						$theta_b = atan2($b["y"] - $centerPoint["y"], $b["x"] - $centerPoint["x"]);
					    $angle_b = fmod(M_PI - M_PI_4 + $theta_a, 2 * M_PI);
					
						if ($angle_a == $angle_b)
							return 0;
					
						return ($angle_a < $angle_b) ? -1 : 1;
					*/
					
						/*
						// Attempt #2: just angles
					
						if ($a["angle"] == $b["angle"])
							return 0;
					
						return (abs($a["angle"]) < abs($b["angle"])) ? -1 : 1;
					*/
						// Attempt #3: Just compare x,y! This is a terrible solution but better than nothing
						// using algorithm from Anonymous on http://www.php.net/manual/en/function.usort.php
						/*
						if ($a["x"] == $b["x"])
						{
							if ($a["y"] == $b["y"])
							{
								return 0;
							}
							elseif ($a["y"] > $b["y"])
							{
								return 1;
							}
							elseif ($a["y"] < $b["y"])
							{
								return -1;
							}
						}
						elseif ($a["x"] > $b["x"])
						{
							return 1;
						}
						elseif ($a["x"] < $b["x"])
						{
							return -1;
						}
						*/
						
						// Attempt #4: order by shortest distance to the target?
						
						//return population[i].chromosomes[a].distance < population[i].chromosomes[b].distance
						//})
				} // endfor i
				
				marshalled, err := json.MarshalIndent(population, "", "  ") // debug line just to show population's structure
				checkErr(err)
				log.Println("Population", marshalled)
				
				// If the user had set agent + cube, clean them up for now
				userDestCube.Store(NullUUID)
				curAgent.Store(NullUUID)
				
				time_end := time.Now()
				diffTime := time_end.Sub(time_start)
				sendMessageToBrowser("status", "info", fmt.Sprintf("<p>CPU time used: %v</p>", diffTime), "")
				
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