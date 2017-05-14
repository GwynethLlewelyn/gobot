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
)

var (
	// Default configurations, hopefully exported to other files and packages
	RootURL, SQLiteDBFilename, URLPathPrefix, PDO_Prefix, PathToStaticFiles, ServerPort string
)

type templateParameters map[string]string;

func main() {
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
	
	fmt.Println("\nGobot configuration read, now testing opening database connection at ", SQLiteDBFilename, "\nPath to static files is:", PathToStaticFiles)
	
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename) // presumes sqlite3 for now
	checkErr(err)

	// query
	rows, err := db.Query("SELECT UUID, Name, Location, Position FROM Agents")
	checkErr(err)
 	   
	var UUID string
	var Name string
 //	var OwnerName string
 //	var OwnerKey string
	var Location string
	var Position string
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
	
	http.HandleFunc(URLPathPrefix + "/admin/agents/",		backofficeAgents)
	http.HandleFunc(URLPathPrefix + "/admin/login/",		backofficeLogin) // unimplemented yet
	http.HandleFunc(URLPathPrefix + "/admin/objects/",		backofficeObjects)
	http.HandleFunc(URLPathPrefix + "/admin/positions/",	backofficePositions)
	http.HandleFunc(URLPathPrefix + "/admin/inventory/",	backofficeInventory)
	http.HandleFunc(URLPathPrefix + "/admin/",				backofficeMain)
	
	// deal with agGrid UI elements
	
	http.HandleFunc(URLPathPrefix + "/uiObjects/",			uiObjects)
	http.HandleFunc(URLPathPrefix + "/uiAgents/",			uiAgents)
	http.HandleFunc(URLPathPrefix + "/uiPositions/",		uiPositions)
	http.HandleFunc(URLPathPrefix + "/uiInventory/",		uiInventory)

	go paralelate() // run everything but the kitchen sink in parallel; yay goroutines!
	
    err = http.ListenAndServe(ServerPort, nil) // set listen port
    checkErr(err)
}

// checkErr logs a fatal error and panics
func checkErr(err error) {
	if err != nil {
		log.Fatal("gobot: ", err)
		panic(err)
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
    for x := 0;x < 100; x++ {
	    fmt.Print(x)
	    time.Sleep(1000 * time.Millisecond)
    }
    fmt.Println("Done! (But I'm hopefully still serving requests)")
}