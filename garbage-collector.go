// Garbage Collector.
package main

import (
//	_ "github.com/go-sql-driver/mysql"
//	"database/sql"
	"fmt"
)

// garbageCollector goes through the database every few hours or so, pings the objects and sees if they're alive, checks their timestamps,
//  and if they are too old
func garbageCollector() {
	fmt.Println("Garbage Collector called.")
}