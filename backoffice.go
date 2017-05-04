// Functions to deal with the backoffice
package main

import (
//	"database/sql"
	"fmt"
//	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"html/template"
//	"os"
//	"strings"
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
//	err := gt.ExecuteTemplate(w, "header", tplParams)
//	if (err != nil) { return err }
//	err = gt.ExecuteTemplate(w, "navigation", tplParams)
//	if (err != nil) { return err }
//	err = gt.ExecuteTemplate(w, "top-menu", tplParams)
//	if (err != nil) { return err }
//	err = gt.ExecuteTemplate(w, "sidebar-left-menu", tplParams)
//	if (err != nil) { return err }
//	err = gt.ExecuteTemplate(w, "footer", tplParams)
//	if (err != nil) { return err }

    err = gt.ExecuteTemplate(w, tplName, tplParams)
	return err
}

// backofficeMain is the main page, probably will have some statistics and such
func backofficeMain(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	fmt.Println("Entered backoffice main func for URL:", r.URL, "URLPathPrefix is:", URLPathPrefix)
	
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - main",
			"Content": "Hi there, this is the main template",
			"URLPathPrefix": URLPathPrefix,
	}
	err := GobotTemplates.gobotRenderer(w, "main", tplParams)
	checkErr(err)
	return
}

// backofficeAgents lists active agents
func backofficeAgents(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered backoffice agents func for URL:", r.URL)
	
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - agents",
			"Content": "Hi there, this is the agents template",
			"URLPathPrefix": URLPathPrefix,
	}
	err := GobotTemplates.gobotRenderer(w, "agents", tplParams)
	checkErr(err)
	return
}

// backofficeObjects lists objects seen as ibstacles
func backofficeObjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered backoffice objects func for URL:", r.URL)
	
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - objects",
			"Content": "Hi there, this is the objects template",
			"URLPathPrefix": URLPathPrefix,
	}
	err := GobotTemplates.gobotRenderer(w, "objects", tplParams)
	checkErr(err)
	return
}