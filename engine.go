package main

import (
	"net/http"
	"time"
)

// engine not implemented yet
func engine(w http.ResponseWriter, r *http.Request) {
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - engine",
			"URLPathPrefix": URLPathPrefix,
			"Content": "Not implemented yet",
	}
	err := GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	time.Sleep(time.Second * 10)
	http.Redirect(w, r, URLPathPrefix + "/", 302)
}