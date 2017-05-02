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

// gobotRenderer assembles the correct templates together and executes them
//  this is mostly to deal with code duplication 
func gobotRenderer(w http.ResponseWriter, tplParams templateParameters, tplFiles ...string) error {
	templates, err := template.ParseGlob(PathToStaticFiles + "/templates/*.tpl")
	if (err != nil) { return err }
	var s = template.New("")
	for _, tplFile := range tplFiles {
		fmt.Println("Looking up ", tplFile, "...")
		s = templates.Lookup(tplFile)
		if (s == nil) { 
			fmt.Println("Glup...", tplFile, "gives error when looking up!")
			continue
		}
	    err = s.ExecuteTemplate(w, tplFile, tplParams)
	    if (err != nil) { return err }
	}
    err = s.Execute(w, tplParams)
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
	err := gobotRenderer(w, tplParams, "header", "main", "navigation", "top-menu", "sidebar-left-menu", "footer")
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
	err := gobotRenderer(w, tplParams, "header", "agents", "navigation", "top-menu", "sidebar-left-menu", "footer")
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
	err := gobotRenderer(w, tplParams, "header", "objects", "navigation", "top-menu", "sidebar-left-menu", "footer")
	checkErr(err)
	return
}