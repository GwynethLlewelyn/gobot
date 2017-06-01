// Functions to deal with the backoffice
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"html/template"
//	"os"
//	"strings"
	"strconv"
//	"crypto/md5"
//	"encoding/hex"
//	"log"
)

// GobotTemplatesType expands on template.Template
//  need to expand it so I can add a few more methods here 
type GobotTemplatesType struct{
	template.Template
}

// GobotTemplates stores all parsed templates for the backoffice
var GobotTemplates GobotTemplatesType

// init parses all templates and puts it inside a (global) var
//  This is supposed to be called just once! (in func main())
func (gt *GobotTemplatesType)init(globbedPath string) error {
	temp, err := template.ParseGlob(globbedPath)
	gt.Template = *temp;
	return err
}

// gobotRenderer assembles the correct templates together and executes them
//  this is mostly to deal with code duplication 
func (gt *GobotTemplatesType)gobotRenderer(w http.ResponseWriter, tplName string, tplParams templateParameters) error {
	var err error

    err = gt.ExecuteTemplate(w, tplName, tplParams)
	return err
}

// backofficeMain is the main page, has some minor statistics, may do this fancier later on
func backofficeMain(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	
	// Open database just to gather some statistics
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename) // presumes sqlite3 for now
	checkErr(err)

	defer db.Close()

	var (
		cnt int
		strAgents, strInventory, strPositions, strObstacles string
	)
		
	err = db.QueryRow("select count(*) from Agents").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strAgents = "Agents: " + strconv.Itoa(cnt)
	} else {
		strAgents = "No Agents."
	}
	
	err = db.QueryRow("select count(*) from Inventory").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strInventory = "Inventory items: " + strconv.Itoa(cnt)
	} else {
		strInventory = "No Inventory items."
	}

	err = db.QueryRow("select count(*) from Positions").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strPositions = "Positions: " + strconv.Itoa(cnt)
	} else {
		strPositions = "No Positions."
	}

	err = db.QueryRow("select count(*) from Obstacles").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strObstacles = "Obstacles: " + strconv.Itoa(cnt)
	} else {
		strObstacles = "No Obstacles."
	}
		
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - main",
			"Content": "This is trash",
			"Agents": strAgents,
			"Inventory": strInventory,
			"Positions": strPositions,
			"Obstacles": strObstacles,
			"URLPathPrefix": URLPathPrefix,
	}
	err = GobotTemplates.gobotRenderer(w, "main", tplParams)
	checkErr(err)
	return
}

// backofficeAgents lists active agents
func backofficeAgents(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - agents",
			"Content": "Hi there, this is the agents template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "agents.js",
	}
	err := GobotTemplates.gobotRenderer(w, "agents", tplParams)
	checkErr(err)
	return
}

// backofficeObjects lists objects seen as obstacles
func backofficeObjects(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - objects",
			"Content": "Hi there, this is the objects template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "objects.js",
	}	
	err := GobotTemplates.gobotRenderer(w, "objects", tplParams)
	checkErr(err)
	return
}

// backofficePositions lists Positions
func backofficePositions(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - positions",
			"Content": "Hi there, this is the positions template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "positions.js",
	}
	err := GobotTemplates.gobotRenderer(w, "positions", tplParams)
	checkErr(err)
	return
}

// backofficeInventory lists the content or inventory currently stored on objects
func backofficeInventory(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - inventory",
			"Content": "Hi there, this is the inventory template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "inventory.js",
	}
	err := GobotTemplates.gobotRenderer(w, "inventory", tplParams)
	checkErr(err)
	return
}

// backofficeLogin deals with authentication (not implemented yet)
func backofficeLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered backoffice login for URL:", r.URL, "using method:", r.Method)
	if r.Method == "GET" {
		tplParams := templateParameters{ "Title": "Gobot Administrator Panel - login",
				"URLPathPrefix": URLPathPrefix,
		}
		err := GobotTemplates.gobotRenderer(w, "login", tplParams)
		checkErr(err)
	} else {
		r.ParseForm()
        // logic part of log in
        fmt.Println("email:", r.Form["email"])
        fmt.Println("password:", r.Form["password"])
        fmt.Println("remember:", r.Form["remember"])
        // we need to set a cookie here etc.
        // redirect to home
        http.Redirect(w, r, URLPathPrefix + "/admin", 302)
	}
	return
}

// backofficeInventory lists the content or inventory currently stored on objects
func backofficeCommands(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - commands",
			"Content": "Blah",
			"URLPathPrefix": URLPathPrefix,
	}
	err := GobotTemplates.gobotRenderer(w, "commands", tplParams)
	checkErr(err)
	return
}