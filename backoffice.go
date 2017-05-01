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

func backoffice(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	
	fmt.Println("Entered backoffice main func for request", r.URL)
	
	mainT, err := template.ParseFiles(PathToStaticFiles + "/templates/main.tpl")
	checkErr(err)
	tplVars := map[string]string {
		"Title": "Gobot Administrator Panel",
		"Content": "Hi there, " + r.URL.Path,
		"URLPathPrefix": URLPathPrefix,
	}
	err = mainT.Execute(w, tplVars)
	checkErr(err)
	return
}
