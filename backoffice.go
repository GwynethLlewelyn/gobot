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

type GobotTemplatesType struct{
	template.Template
}

var GobotTemplates GobotTemplatesType

//
func (gt *GobotTemplatesType)init(globbedPath string) error {
	temp, err := template.ParseGlob(globbedPath)
	gt.Template = *temp;
	return err
}

// gobotRenderer assembles the correct templates together and executes them
//  this is mostly to deal with code duplication 
func (gt *GobotTemplatesType)gobotRenderer(w http.ResponseWriter, tplName string, tplParams templateParameters) error {
    err := gt.ExecuteTemplate(w, tplName, tplParams)
	if (err != nil) { return err }
	
	return nil
}

// backofficeMain is the main page, probably will have some statistics and such
func backofficeMain(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	fmt.Println("Entered backoffice main func for URL:", r.URL)
	
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