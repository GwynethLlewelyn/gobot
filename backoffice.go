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

// backofficeMain is the main page, probably will have some statistics and such
func backofficeMain(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	fmt.Println("Entered backoffice main func for URL:", r.URL)
	
	templates, err := template.ParseFiles(PathToStaticFiles + "/templates/header.tpl", PathToStaticFiles + "/templates/footer.tpl", PathToStaticFiles + "/templates/main.tpl") // menu will come here too
	checkErr(err)
	
	tplParam := templateParameters{ "Title": "Gobot Administrator Panel - main",
			"Content": "Hi there, this is the main template",
			"URLPathPrefix": URLPathPrefix,
		}
	
	s1 := templates.Lookup("header.tpl")
    err = s1.ExecuteTemplate(w, "header", tplParam)
    checkErr(err)
    s2 := templates.Lookup("main.tpl")
    err = s2.ExecuteTemplate(w, "main", tplParam)
    checkErr(err)
    s3 := templates.Lookup("footer.tpl")
    err = s3.ExecuteTemplate(w, "footer", tplParam)
    checkErr(err)
    err = s3.Execute(w, tplParam)
	checkErr(err)
	
	return
}

// backofficeAgents lists active agents
func backofficeAgents(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered backoffice agents func for URL:", r.URL)
	
	mainT, err := template.ParseFiles(PathToStaticFiles + "/templates/agents.tpl")
	checkErr(err)
	err = mainT.Execute(w, templateParameters{ "Title": "Gobot Administrator Panel - agents",
			"Content": "Hi there, this is the agents template",
			"URLPathPrefix": URLPathPrefix,
		})
	checkErr(err)
	return
}

// backofficeObjects lists objects
func backofficeObjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered backoffice objects func for URL:", r.URL)
	
	mainT, err := template.ParseFiles(PathToStaticFiles + "/templates/objects.tpl")
	checkErr(err)
	err = mainT.Execute(w, templateParameters{ "Title": "Gobot Administrator Panel - objects",
			"Content": "Hi there, this is the objects template",
			"URLPathPrefix": URLPathPrefix,
		})
	checkErr(err)
	return
}