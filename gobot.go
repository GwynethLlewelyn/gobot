// gobot is an attempt to do a single, monolithic Go application which deals with autonomous agents in OpenSimulator
package main

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper" // to read config files
	"net/http"
	"log"
	"os/user"
	"path/filepath"
	"gopkg.in/guregu/null.v3/zero" // can be deleted once we move the database tests elsewhere
)

var (
	// Default configurations, hopefully exported to other files and packages
	RootURL, SQLiteDBFilename, URLPathPrefix, PDO_Prefix, PathToStaticFiles, ServerPort, MapURL string
)

//type templateParameters map[string]string
type templateParameters map[string]interface{}

func main() {
	// to change the flags on the default logger
	// see https://stackoverflow.com/a/24809859/1035977
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	fmt.Print("Reading Gobot configuration...")
	// Open our config file and extract relevant data from there
	viper.SetConfigName("config")
	viper.SetConfigType("toml") // just to make sure; it's the same format as OpenSimulator (or MySQL) config files
	viper.AddConfigPath("$HOME/go/src/gobot/") // that's how I have it
	viper.AddConfigPath("$HOME/go/src/github.com/GwynethLlewelyn/gobot/") // that's how you'll have it
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error in config file: %s \n", err))
	}
	
	// Without these set, we cannot do anything
	RootURL = viper.GetString("gobot.RootURL"); fmt.Print(".")
	URLPathPrefix = viper.GetString("gobot.URLPathPrefix"); fmt.Print(".")
	SQLiteDBFilename = viper.GetString("gobot.SQLiteDBFilename"); fmt.Print(".")
	PDO_Prefix = viper.GetString("gobot.PDO_Prefix"); fmt.Print(".")
	viper.SetDefault("go.PathToStaticFiles", "~/go/src/gobot")
	path, err := expandPath(viper.GetString("gobot.PathToStaticFiles")); fmt.Print(".")
	checkErr(err)
	PathToStaticFiles = path
	viper.SetDefault("gobot.ServerPort", ":3000")
	ServerPort = viper.GetString("gobot.ServerPort"); fmt.Print(".")
	MapURL = viper.GetString("opensim.MapURL"); fmt.Print(".")
	
	fmt.Println("\nGobot configuration read, now testing opening database connection at ", SQLiteDBFilename, "\nPath to static files is:", PathToStaticFiles)
	
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename) // presumes sqlite3 for now
	checkErr(err)

	// query
	rows, err := db.Query("SELECT UUID, Name, Location, Position FROM Agents")
	checkErr(err)
 	   
	var UUID zero.String
	var Name zero.String
 //	var OwnerName string
 //	var OwnerKey string
	var Location zero.String
	var Position zero.String
/*	var Rotation string
	var Velocity string
	var Energy string
	var Money string
	var Happiness string
	var Class string
	var SubType string
	var PermURL string
	var LastUpdate string
	var BestPath string
	var SecondBestPath string
	var CurrentTarget string
*/

	for rows.Next() {
		err = rows.Scan(&UUID, &Name, &Location, &Position)
		checkErr(err)
		fmt.Println(UUID)
		fmt.Println(Name)
		fmt.Println(Location)
		fmt.Println(Position)
	}
	
	db.Close()
	
	fmt.Println("\n\nDatabase tests ended.\n\nStarting Gobot application at port", ServerPort, "\nfor URL:", URLPathPrefix)
	
	// this was just to make tests, now start the web server
	
	// Load all templates
	err = GobotTemplates.init(PathToStaticFiles + "/templates/*.tpl")
	checkErr(err)
	
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
	http.HandleFunc(URLPathPrefix + "/admin/engine/",					backofficeMain)
	http.HandleFunc(URLPathPrefix + "/admin/",							backofficeMain)
	http.HandleFunc(URLPathPrefix + "/",								backofficeLogin) // if not auth, then get auth
	
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

	go paralelate() // run everything but the kitchen sink in parallel; yay goroutines!
	// very likely we will open the database, look at all agents, and run a goroutine for each (20170516)
	
    err = http.ListenAndServe(ServerPort, nil) // set listen port
    checkErr(err)
}

// checkErrPanic logs a fatal error and panics
func checkErrPanic(err error) {
	if err != nil {
		log.Panic("gobot panic: ", err)
	}
}

// checkErr checks if there is an error, and if yes, it logs it out and continues
//  this is for 'normal' situations when we want to get a log if something goes wrong but do not need to panic
func checkErr(err error) {
	if err != nil {
		log.Print(err)
	}
}

// expandPath expands the tilde as the user's home directory
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

// paralelate is a first attempt at a goroutine
func paralelate() {
	fmt.Println("Testing parallelism...")
    for true {
	    fmt.Print("\b|")
	    time.Sleep(1000 * time.Millisecond)
	    fmt.Print("\b/")
	    time.Sleep(1000 * time.Millisecond)
	    fmt.Print("\b-")
	    time.Sleep(1000 * time.Millisecond)
	    fmt.Print("\b\\")
	    time.Sleep(1000 * time.Millisecond)
    }
    fmt.Println("Done! (But I'm hopefully still serving requests)")
}