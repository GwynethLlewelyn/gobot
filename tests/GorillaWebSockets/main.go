package main

import (
	"fmt"
	//   "io"
	"log"
	"net/http"
	//   "github.com/gorilla/websocket"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
    fmt.Println("WebSocket Endpoint Hit")
    conn, err := Upgrade(w, r)
    if err != nil {
        fmt.Fprintf(w, "%+v\n", err)
    }

	if (conn == nil) {
		log.Println("Client not using websocket protocol")
		fmt.Fprintf(w, "%+v\n", err)
		return
	}

    client := &Client{
        Conn: conn,
        Pool: pool,
    }

    pool.Register <- client
    client.Read()
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
    pool := NewPool()
    go pool.Start()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(pool, w, r)
    })
}

func main() {
    fmt.Println("Distributed Chat App v0.01")
    setupRoutes()
    log.Fatal(http.ListenAndServe(":8080", nil))
}