// gobot is an attempt to do a single, monolithic Go application which deals with autonomous agents in OpenSimulator
package main

import (
	"database/sql"
	"fmt"
 /*   "time" */
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper" // to read config files
	"net/http"
	"log"
)

var (
	// Default configurations, hopefully exported to other files and packages
	RootURL, SQLiteDBFilename, URLPathPrefix, PDO_Prefix, PathToStaticFiles, ServerPort string
)

func main() {
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
	RootURL = viper.GetString("gobot.RootURL")
	URLPathPrefix = viper.GetString("gobot.URLPathPrefix")
	SQLiteDBFilename = viper.GetString("gobot.SQLiteDBFilename")
	PDO_Prefix = viper.GetString("gobot.PDO_Prefix")
	viper.SetDefault("go.PathToStaticFiles", "../src/gobot")
	PathToStaticFiles = viper.GetString("go.PathToStaticFiles")
	viper.SetDefault("gobot.ServerPort", ":3000")
	ServerPort = viper.GetString("gobot.ServerPort")
	
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
	
	// this was just to make tests, now start the web server
	
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
	http.HandleFunc(URLPathPrefix + "/lib/",				http.FileServer(http.Dir(PathToStaticFiles + "/lib"))
	http.HandleFunc(URLPathPrefix + "/templates/",			http.FileServer(http.Dir(PathToStaticFiles + "/templates"))
	
	// Deal with templated output for the admin back office
	//  If this works I'll buy someone lunch! (GwynethLlewelyn 20170430)
	http.HandleFunc(URLPathPrefix + "/admin/",				backoffice) // defined on backoffice.go
	
    err = http.ListenAndServe(ServerPort, nil) // set listen port
    checkErr(err)	
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("gobot: ", err)
		panic(err)
	}
}