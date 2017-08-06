// gobot is an attempt to do a single, monolithic Go application which deals with autonomous agents in OpenSimulator.
package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"github.com/fatih/color" // allows ANSI escaping for logging in colour! (20170806)
	"github.com/Pallinder/go-randomdata"
	"github.com/spf13/viper" // to read config files
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
)


var (
	// Default configurations, hopefully exported to other files and packages
	// we probably should have a struct for this
	Host, GoBotDSN, URLPathPrefix, PDO_Prefix, PathToStaticFiles,
	ServerPort, FrontEnd, MapURL, LSLSignaturePIN string
)

const NullUUID = "00000000-0000-0000-0000-000000000000" // always useful when we deal with SL/OpenSimulator...

//type templateParameters map[string]string
type templateParameters map[string]interface{}

// loadConfiguration loads all the configuration from the config.toml file.
// It's a separate function because we want to be able to do a killall -HUP gobot to force the configuration to be read again
func loadConfiguration() {
	log.Print("Reading Gobot configuration...")
	// Open our config file and extract relevant data from there
	viper.SetConfigName("config")
	viper.SetConfigType("toml") // just to make sure; it's the same format as OpenSimulator (or MySQL) config files
	viper.AddConfigPath("$HOME/go/src/gobot/") // that's how I have it
	viper.AddConfigPath("$HOME/go/src/github.com/GwynethLlewelyn/gobot/") // that's how you'll have it
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	checkErr(err) // Handle errors reading the config file
	
	// Without these set, we cannot do anything
	viper.SetDefault("gobot.Host", "localhost") // to prevent bombing out with panics
	Host = viper.GetString("gobot.Host"); fmt.Print(".")
	URLPathPrefix = viper.GetString("gobot.URLPathPrefix"); fmt.Print(".")
	GoBotDSN = viper.GetString("gobot.GoBotDSN"); fmt.Print(".")
	PDO_Prefix = viper.GetString("gobot.PDO_Prefix"); fmt.Print(".")
	viper.SetDefault("gobot.PathToStaticFiles", "~/go/src/gobot")
	path, err := expandPath(viper.GetString("gobot.PathToStaticFiles")); fmt.Print(".")
	checkErr(err)
	PathToStaticFiles = path
	viper.SetDefault("gobot.ServerPort", ":3000")
	ServerPort = viper.GetString("gobot.ServerPort"); fmt.Print(".")
	FrontEnd = viper.GetString("gobot.FrontEnd"); fmt.Print(".")
	MapURL = viper.GetString("opensim.MapURL"); fmt.Print(".")
	viper.SetDefault("gobot.LSLSignaturePIN", "9876") // better than no signature at all
	LSLSignaturePIN = viper.GetString("opensim.LSLSignaturePIN"); fmt.Print(".")
}

// main() starts here.
func main() {
	// to change the flags on the default logger
	// see https://stackoverflow.com/a/24809859/1035977
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	loadConfiguration() // this gets loaded always, on the first time it runs
	
	// prepares a special channel to look for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
	
	// goroutine which listens to signals and calls the loadConfiguration() function if someone sends us a HUP
	go func() {
		for {
	        sig := <-sigs
	        log.Println("Got signal", sig)
	        switch sig {
		        case syscall.SIGHUP:
		        	log.Println(" ... reloading Gobot configuration again:")
		        	loadConfiguration()
		        case syscall.SIGUSR1:
		        	sendMessageToBrowser("status", "", randomdata.FullName(randomdata.Female) + "<br />", "") // defined on engine.go for now
		        case syscall.SIGUSR2:
		        	sendMessageToBrowser("status", "", randomdata.Country(randomdata.FullCountry) + "<br />", "") // defined on engine.go for now
		        case syscall.SIGCONT:
		        	// HACK(gwyneth): if the engine dies, send a SIGCONT to get it running again (20170723).
		        	log.Println("SIGCONT caught; trying to launch the engine again")
		        	sendMessageToBrowser("status", "restart", "<code>SIGCONT</code> caught; trying to launch the engine again<br />", "")
		        	go engine() // is this a good idea? Maybe we ought to have a flag saying if we're running or not! (20170723)
		        default:
		        	log.Println("Unknown UNIX signal caught!! Ignoring...")
	        }
        }
    }()
	
	// do some database tests. If it fails, it means the database is broken or corrupted and it's worthless
	//  to run this application anyway!
	log.Println("\nTesting opening database connection at ", GoBotDSN, "\nPath to static files is:", PathToStaticFiles)
	
	db, err := sql.Open(PDO_Prefix, GoBotDSN) // presumes sqlite3 for now
	checkErr(err) // abort if it cannot even open the database

	// query
	rows, err := db.Query("SELECT UUID, Name, Location, Position FROM Agents")
	checkErr(err) // if select fails, probably the table doesn't even exist
 	
 	var agent AgentType; // type defined on ui.go to be used on database requests

	for rows.Next() {
		err = rows.Scan(&agent.UUID, &agent.Name, &agent.Location, &agent.Position)
		checkErr(err) // if we get some errors here, we will get in trouble later on
		log.Println(agent.UUID)
		log.Println(agent.Name)
		log.Println(agent.Location)
		log.Println(agent.Position)
	}
	rows.Close()
	db.Close()
	
	log.Println("\n\nDatabase tests ended.\n\nStarting Gobot application at port", ServerPort, "\nfor URL:", URLPathPrefix)
	
	// this was just to make tests; now start the engine as a separate goroutine in the background
	
	go engine() // run everything but the kitchen sink in parallel; yay goroutines!

	go garbageCollector() // this will periodically remove from the database all old items that are 'dead' (20170730)

	// Now prepare the web interface
	
	// Load all templates
	err = GobotTemplates.init(PathToStaticFiles + "/templates/*.tpl")
	checkErr(err) // abort if templates are not found
	
	// Configure routers for our many inworld scripts
	// In my case, paths with /go will be served by gobot, the rest by nginx as before
	// Exception is for static files
	http.HandleFunc(URLPathPrefix + "/update-inventory/",	updateInventory) 
	http.HandleFunc(URLPathPrefix + "/update-sensor/",		updateSensor) 
	http.HandleFunc(URLPathPrefix + "/register-position/",	registerPosition) 
	http.HandleFunc(URLPathPrefix + "/register-agent/",		registerAgent) 
	http.HandleFunc(URLPathPrefix + "/configure-cube/",		configureCube)
	
	// Static files. This should be handled directly by nginx, but we include it here
	//  for a standalone version...	
	fslib := http.FileServer(http.Dir(PathToStaticFiles + "/lib"))
	http.Handle(URLPathPrefix + "/lib/", http.StripPrefix(URLPathPrefix + "/lib/", fslib))

	templatelib := http.FileServer(http.Dir(PathToStaticFiles + "/templates"))
	http.Handle(URLPathPrefix + "/templates/",
		http.StripPrefix(URLPathPrefix + "/templates/", templatelib)) // not sure if this is needed
	
	// Deal with templated output for the admin back office, defined on backoffice.go
	// For now this is crude, each page is really very similar, but there are not many so each will get its own handler function for now
	
	http.HandleFunc(URLPathPrefix + "/admin/agents/",					backofficeAgents)
	http.HandleFunc(URLPathPrefix + "/admin/logout/",					backofficeLogout)
	http.HandleFunc(URLPathPrefix + "/admin/login/",					backofficeLogin) // probably not necessary
	http.HandleFunc(URLPathPrefix + "/admin/objects/",					backofficeObjects)
	http.HandleFunc(URLPathPrefix + "/admin/positions/",				backofficePositions)
	http.HandleFunc(URLPathPrefix + "/admin/inventory/",				backofficeInventory)
	http.HandleFunc(URLPathPrefix + "/admin/user-management/",			backofficeUserManagement)
	http.HandleFunc(URLPathPrefix + "/admin/commands/exec/",			backofficeCommandsExec)
	http.HandleFunc(URLPathPrefix + "/admin/commands/",					backofficeCommands)
	http.HandleFunc(URLPathPrefix + "/admin/controller-commands/exec/",	backofficeControllerCommandsExec)
	http.HandleFunc(URLPathPrefix + "/admin/controller-commands/",		backofficeControllerCommands)
	http.HandleFunc(URLPathPrefix + "/admin/engine/",					backofficeEngine)
	// LSL Template Generator
	http.HandleFunc(URLPathPrefix + "/admin/lsl-register-object/",		backofficeLSLRegisterObject)	
	http.HandleFunc(URLPathPrefix + "/admin/lsl-bot-controller/",		backofficeLSLBotController)	
	http.HandleFunc(URLPathPrefix + "/admin/lsl-agent-scripts/",		backofficeLSLAgentScripts)
	// fallthrough for admin
	http.HandleFunc(URLPathPrefix + "/admin/",							backofficeMain)
	
	// deal with agGrid UI elements
	http.HandleFunc(URLPathPrefix + "/uiObjects/",						uiObjects)
	http.HandleFunc(URLPathPrefix + "/uiObjectsUpdate/",				uiObjectsUpdate) // to change the database manually
	http.HandleFunc(URLPathPrefix + "/uiObjectsRemove/",				uiObjectsRemove) // to remove rows of the database manually
	http.HandleFunc(URLPathPrefix + "/uiAgents/",						uiAgents)
	http.HandleFunc(URLPathPrefix + "/uiAgentsUpdate/",					uiAgentsUpdate)
	http.HandleFunc(URLPathPrefix + "/uiAgentsRemove/",					uiAgentsRemove)
	http.HandleFunc(URLPathPrefix + "/uiPositions/",					uiPositions)
	http.HandleFunc(URLPathPrefix + "/uiPositionsUpdate/",				uiPositionsUpdate)
	http.HandleFunc(URLPathPrefix + "/uiPositionsRemove/",				uiPositionsRemove)
	http.HandleFunc(URLPathPrefix + "/uiInventory/",					uiInventory)
	http.HandleFunc(URLPathPrefix + "/uiInventoryUpdate/",				uiInventoryUpdate)
	http.HandleFunc(URLPathPrefix + "/uiInventoryRemove/",				uiInventoryRemove)
	http.HandleFunc(URLPathPrefix + "/uiUserManagement/",				uiUserManagement)
	http.HandleFunc(URLPathPrefix + "/uiUserManagementUpdate/",			uiUserManagementUpdate)
	http.HandleFunc(URLPathPrefix + "/uiUserManagementRemove/",			uiUserManagementRemove)	
	
	// Handle Websockets on Engine
	http.Handle(URLPathPrefix + "/wsEngine/",							websocket.Handler(serveWs))

	http.HandleFunc(URLPathPrefix + "/",								backofficeLogin) // if not auth, then get auth
	
    err = http.ListenAndServe(ServerPort, nil) // set listen port
    checkErr(err) // if it can't listen to all the above, then it has to abort anyway
}

// checkErrPanic logs a fatal error and panics.
func checkErrPanic(err error) {
	if err != nil {
		color.Set(color.FgRed)
		defer color.Unset()
		log.Panic("gobot panic: ", err)
	}
}

// checkErr checks if there is an error, and if yes, it logs it out and continues.
//  this is for 'normal' situations when we want to get a log if something goes wrong but do not need to panic
func checkErr(err error) {
	if err != nil {
		color.Set(color.FgYellow)
		log.Println("gobot error: ", err)
		color.Unset()
	}
}

// expandPath expands the tilde as the user's home directory.
//  found at http://stackoverflow.com/a/43578461/1035977
func expandPath(path string) (string, error) {
    if len(path) == 0 || path[0] != '~' {
        return path, nil
    }

    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    return filepath.Join(usr.HomeDir, path[1:]), nil
}