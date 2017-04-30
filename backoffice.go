package main

import (
//	"database/sql"
//	"fmt"
//	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"html/template"
	"os"
//	"strings"
//	"crypto/md5"
//	"encoding/hex"
//	"log"
)

func backoffice(w http.ResponseWriter, r *http.Request) {
	// let's load the main template for now, just to make sure this works
	
	mainT, err := template.ParseFiles(PathToStaticFiles + "/templates/main.tpl")
	checkErr(err)
	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{
			"My photos",
			"My blog",
		},
	}
	err = mainT.Execute(os.Stdout, data)
	checkErr(err)
	return
}
	