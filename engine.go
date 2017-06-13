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
	_ = GobotTemplates.gobotRenderer(w, r, "main", tplParams) // ignore error, we'll redirect anyway
	time.Sleep(time.Second * 10)
	http.Redirect(w, r, URLPathPrefix + "/", 302)
}