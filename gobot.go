package main

import (
	"database/sql"
	"fmt"
 /*   "time" */
	_ "github.com/mattn/go-sqlite3"
)

const DbFilename string = "../../../../web/database/botmover.db"

func main() {
	db, err := sql.Open("sqlite3", DbFilename)
	checkErr(err)

	// query
	rows, err := db.Query("SELECT UUID, Name, Location, Position FROM Agents")
	checkErr(err)
 	   
	var UUID string
	var Name string
 //	var OwnerName string
 //	var OwnerKey string
	var Location string
	var Position string
/*	var Rotation string
	var Velocity string
	var Energy string
	var Money string
	var Happiness string
	var Class string
	var SubType string
	var PermURL string
	var LastUpdate string
	var BestPath string
	var SecondBestPath string
	var CurrentTarget string
*/

	for rows.Next() {
		err = rows.Scan(&UUID, &Name, &Location, &Position)
		checkErr(err)
		fmt.Println(UUID)
		fmt.Println(Name)
		fmt.Println(Location)
		fmt.Println(Position)
	}
	
	db.Close()

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}