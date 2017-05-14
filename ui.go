// Functions to deal with the agGrid (UI)
package main

import (
	"database/sql"
	"fmt"
	"encoding/json"
	"net/http"
)

// uiObjects creates a JSON representation of the Obstacles table and spews it out
func uiObjects(w http.ResponseWriter, r *http.Request) {
// struct to hold data retrieved from the database
	type objectType struct {
		UUID string
		Name string
		BotKey string
		BotName string
		Type int
		Position string
		Rotation string
		Velocity string
		LastUpdate string
		Origin string
		Phantom int
		Prims int
		BBHi string
		BBLo string
	}

	var (
		rowArr []interface{}
		Object objectType
	)

	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT * FROM Obstacles")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(
			&Object.UUID,
			&Object.Name,
			&Object.BotKey,
			&Object.BotName,
			&Object.Type,
			&Object.Position,
			&Object.Rotation,
			&Object.Velocity,
			&Object.LastUpdate,
			&Object.Origin,
			&Object.Phantom,
			&Object.Prims,
			&Object.BBHi,
			&Object.BBLo,
		)
		//fmt.Println("Row extracted:", Object)
		rowArr = append(rowArr, Object)
	}
	checkErr(err)

	// produces neatly indented output; see https://blog.golang.org/json-and-go but especially http://stackoverflow.com/a/37084385/1035977
	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		// fmt.Printf("json.MarshalIndent:\n%s\n\n", data)
		_, err := fmt.Fprintf(w, "%s", data)
		//if (err == nil) { fmt.Printf("Wrote %d bytes to interface\n", n) } else { checkErr(err) }
		checkErr(err)
	}
	return
}

// uiAgents creates a JSON representation of the Agents table and spews it out
func uiAgents(w http.ResponseWriter, r *http.Request) {
	type agentType struct {
		UUID string
		Name string
		OwnerName string
		OwnerKey string
		Location string
		Position string
		Rotation string
		Velocity string
		Energy string
		Money string
		Happiness string
		Class string
		SubType string
		PermURL string
		LastUpdate string
		BestPath string
		SecondBestPath string
		CurrentTarget string
	}

	var (
		rowArr []interface{}
		Agent agentType
	)

	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT * FROM Agents")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(
			&Agent.UUID,
			&Agent.Name,
			&Agent.OwnerName,
			&Agent.OwnerKey,
			&Agent.Location,
			&Agent.Position,
			&Agent.Rotation,
			&Agent.Velocity,
			&Agent.Energy,
			&Agent.Money,
			&Agent.Happiness,
			&Agent.Class,
			&Agent.SubType,
			&Agent.PermURL,
			&Agent.LastUpdate,
			&Agent.BestPath,
			&Agent.SecondBestPath,
			&Agent.CurrentTarget,
		)		
		rowArr = append(rowArr, Agent)
	}
	checkErr(err)

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	return
}

// uiPositions creates a JSON representation of the Positions table and spews it out
func uiPositions(w http.ResponseWriter, r *http.Request) {
	type positionType struct {
		PermURL string
		UUID string
		Name string
		OwnerName string
		Location string
		Position string
		Rotation string
		Velocity string
		LastUpdate string
		OwnerKey string
		ObjectType string
		ObjectClass string
		RateEnergy string
		RateMoney string
		RateHappiness string
	}

	var (
		rowArr []interface{}
		Position positionType
	)

	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT * FROM Positions")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(
			&Position.PermURL,
			&Position.UUID,
			&Position.Name,
			&Position.OwnerName,
			&Position.Location,
			&Position.Position,
			&Position.Rotation,
			&Position.Velocity,
			&Position.LastUpdate,
			&Position.OwnerKey,
			&Position.ObjectType,
			&Position.ObjectClass,
			&Position.RateEnergy,
			&Position.RateMoney,
			&Position.RateHappiness,
		)		
		rowArr = append(rowArr, Position)
	}
	checkErr(err)

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	return
}

// uiInventory creates a JSON representation of the Inventory table and spews it out
func uiInventory(w http.ResponseWriter, r *http.Request) {
	type inventoryType struct {
		UUID string
		Name string
		Type string
		LastUpdate string
		Permissions string
	}

	var (
		rowArr []interface{}
		Inventory inventoryType
	)

	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	checkErr(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT * FROM Inventory")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(
			&Inventory.UUID,
			&Inventory.Name,
			&Inventory.Type,
			&Inventory.LastUpdate,
			&Inventory.Permissions,
		)		
		rowArr = append(rowArr, Inventory)
	}
	checkErr(err)

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	return
}