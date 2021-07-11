// Functions to deal with the agGrid (UI).
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v3/zero"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
)

// Auxiliary functions for HTTP handling

// checkErrHTTP returns an error via HTTP and also logs the error.
func checkErrHTTP(w http.ResponseWriter, httpStatus int, errorMessage string, err error) {
	if err != nil {
		http.Error(w, fmt.Sprintf(errorMessage, err), httpStatus)
		pc, file, line, ok := runtime.Caller(1)
		Log.Error("(", http.StatusText(httpStatus), ") ", filepath.Base(file), ":", line, ":", pc, ok, " - error:", errorMessage, err)
	}
}

// checkErrPanicHTTP returns an error via HTTP and logs the error with a panic.
func checkErrPanicHTTP(w http.ResponseWriter, httpStatus int, errorMessage string, err error) {
	if err != nil {
		http.Error(w, fmt.Sprintf(errorMessage, err), httpStatus)
		pc, file, line, ok := runtime.Caller(1)
		Log.Panic("(", http.StatusText(httpStatus), ") ", filepath.Base(file), ":", line, ":", pc, ok, " - panic:", errorMessage, err)
	}
}

// logErrHTTP assumes that the error message was already composed and writes it to HTTP and logs it.
//  this is mostly to avoid code duplication and make sure that all entries are written similarly
func logErrHTTP(w http.ResponseWriter, httpStatus int, errorMessage string) {
	http.Error(w, errorMessage, httpStatus)
	Log.Error("(" + http.StatusText(httpStatus) + ") " + errorMessage)
}

// funcName is @Sonia's solution to get the name of the function that Go is currently running.
//  This will be extensively used to deal with figuring out where in the code the errors are!
//  Source: https://stackoverflow.com/a/10743805/1035977 (20170708)
func funcName() string {
    pc, _, _, _ := runtime.Caller(1)
    return runtime.FuncForPC(pc).Name()
}

// Main functions to respond to agGrid
//
// Each function class has a struct type to deal with database requests

// objectType is a struct to hold data retrieved from the database, used by several functions (including JSON).
type ObjectType struct {
	UUID zero.String
	Name zero.String
	BotKey zero.String
	BotName zero.String
	Type zero.String // `json:"string"`
	Position zero.String
	Rotation zero.String
	Velocity zero.String
	LastUpdate zero.String
	Origin zero.String
	Phantom zero.String // `json:"string"`
	Prims zero.String // `json:"string"`
	BBHi zero.String
	BBLo zero.String
	Coords_region string // These two are not on the database but calculated on demand (20170722)
	Coords_xyz []string //  can be a string since it will never be deJSONified
}

// uiObjects creates a JSON representation of the Obstacles table and spews it out.
func uiObjects(w http.ResponseWriter, r *http.Request) {
	var (
		rowArr []interface{}
		Object ObjectType
	)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanic(err)

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
		// Log.Debug("Row extracted:", Object)
		rowArr = append(rowArr, Object)
	}
	checkErr(err)
	defer rows.Close()

	// produces neatly indented output; see https://blog.golang.org/json-and-go but especially http://stackoverflow.com/a/37084385/1035977
	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		// Log.Debugf("json.MarshalIndent:\n%s\n\n", data)
		_, err := fmt.Fprintf(w, "%s", data)
		//if (err == nil) { Log.Debugf("Wrote %d bytes to interface\n", n) } else { checkErr(err) }
		checkErr(err)
	}
	// return
}

// uiObjectsUpdate receives a JSON representation of one row (from the agGrid) in order to update our database.
func uiObjectsUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body) // from https://stackoverflow.com/questions/15672556/handling-json-post-request-in-go (20170524)
    checkErrPanic(err)
	// Log.Debug("\nBody is >>", string(body), "<<")
	var obj ObjectType
    err = json.Unmarshal(body, &obj)
    checkErrPanic(err)
    // Log.Debug("\nJSON decoded body is >>", obj, "<<")

    // update database
    // open database connection and see if we can update the inventory for this object
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	stmt, err := db.Prepare("REPLACE INTO Obstacles (`UUID`, `Name`, `BotKey`, `BotName`, `Type`, `Position`, `Rotation`, `Velocity`, `LastUpdate`, `Origin`, `Phantom`, `Prims`, `BBHi`, `BBLo`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace prepare failed:", err)

	defer stmt.Close()

	_, err = stmt.Exec(obj.UUID, obj.Name, obj.BotKey, obj.BotName, obj.Type, obj.Position,
		obj.Rotation, obj.Velocity, obj.LastUpdate, obj.Origin, obj.Phantom, obj.Prims, obj.BBHi, obj.BBLo)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace exec failed:", err)
	// w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-type", "text/plain; charset=utf-8")
	// fmt.Fprintln(w, obj, "successfully updated.")
	// Log.Debug(obj, "successfully updated.")
	// return
}

// uiObjectsRemove receives a list of UUIDs to remove from the Obstacles table.
func uiObjectsRemove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Cannot read body of HTTP Request:", err)
	// Log.Debug("\nObjects body is >>", string(body), "<<")

    // open database connection and see if we can remove the object UUIDs we got
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM Obstacles WHERE UUID IN (%s)", string(body)))
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Objects remove failed:", err)

	Log.Debug("Object UUIDs >>", string(body), "<< successfully removed.")
}

// agentType is a struct to hold data retrieved from the database.
type AgentType struct {
	UUID zero.String
	Name zero.String
	OwnerName zero.String
	OwnerKey zero.String
	Location zero.String
	Position zero.String
	Rotation zero.String
	Velocity zero.String
	Energy zero.String
	Money zero.String
	Happiness zero.String
	Class zero.String
	SubType zero.String
	PermURL zero.String
	LastUpdate zero.String
	BestPath zero.String
	SecondBestPath zero.String
	CurrentTarget zero.String
	Coords_region string
	Coords_xyz []string
}

// uiAgents creates a JSON representation of the Agents table and spews it out.
func uiAgents(w http.ResponseWriter, r *http.Request) {
	var (
		rowArr []interface{}
		Agent AgentType
	)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanic(err)

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
	defer rows.Close()

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	// return
}

// uiAgentsUpdate receives a JSON representation of one row (from the agGrid) in order to update our database.
func uiAgentsUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	var ag AgentType
    err = json.Unmarshal(body, &ag)
    checkErrPanic(err)

    db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	stmt, err := db.Prepare("REPLACE INTO Agents (`UUID`, `Name`, `OwnerName`, `OwnerKey`, `Location`, `Position`, `Rotation`, `Velocity`, `Energy`, `Money`, `Happiness`, `Class`, `SubType`, `PermURL`, `LastUpdate`, `BestPath`, `SecondBestPath`, `CurrentTarget`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace prepare failed:", err)

	defer stmt.Close()

	_, err = stmt.Exec(ag.UUID, ag.Name, ag.OwnerName, ag.OwnerKey, ag.Location, ag.Position,
		ag.Rotation, ag.Velocity, ag.Energy, ag.Money, ag.Happiness, ag.Class, ag.SubType, ag.PermURL,
		ag.LastUpdate, ag.BestPath, ag.SecondBestPath, ag.CurrentTarget)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace exec failed:", err)

	// return
}

// uiAgentsRemove receives a list of UUIDs to remove from the Agents table.
func uiAgentsRemove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	// Log.Debug("\nAgents Body is >>", string(body), "<<")

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM Agents WHERE UUID IN (%s)", string(body)))
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Agents remove failed:", err)

	Log.Debug("Agents UUIDs >>", string(body), "<< successfully removed.")
}

// PositionType is a struct to hold data retrieved from the database, used by several functions (including JSON).
type PositionType struct {
	PermURL zero.String
	UUID zero.String
	Name zero.String
	OwnerName zero.String
	Location zero.String
	Position zero.String
	Rotation zero.String
	Velocity zero.String
	LastUpdate zero.String
	OwnerKey zero.String
	ObjectType zero.String
	ObjectClass zero.String
	RateEnergy zero.String
	RateMoney zero.String
	RateHappiness zero.String
	Coords_region string
	Coords_xyz []string
	DistanceToAgent float64	// This does not get saved to the database, since it's different for every agent (20170811).
}

// uiPositions creates a JSON representation of the Positions table and spews it out.
func uiPositions(w http.ResponseWriter, r *http.Request) {
	var (
		rowArr []interface{}
		Position PositionType
	)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanic(err)

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
	defer rows.Close()

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	// return
}

// uiPositionsUpdate receives a JSON representation of one row (from the agGrid) in order to update our database.
func uiPositionsUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	var pos PositionType
    err = json.Unmarshal(body, &pos)
    checkErrPanic(err)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	stmt, err := db.Prepare("REPLACE INTO Positions (`PermURL`, `UUID`, `Name`, `OwnerName`, `Location`, `Position`, `Rotation`, `Velocity`, `LastUpdate`, `OwnerKey`, `ObjectType`, `ObjectClass`, `RateEnergy`, `RateMoney`, `RateHappiness`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace prepare failed:", err)

	defer stmt.Close()

	_, err = stmt.Exec(pos.PermURL, pos.UUID, pos.Name, pos.OwnerName, pos.Location, pos.Position,
		pos.Rotation, pos.Velocity, pos.LastUpdate, pos.OwnerKey, pos.ObjectType, pos.ObjectClass,
		pos.RateEnergy, pos.RateMoney, pos.RateHappiness)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace exec failed:", err)

	// return
}

// uiPositionsRemove receives a list of UUIDs to remove from the Positions table.
func uiPositionsRemove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	// Log.Debug("\nPositions Body is >>", string(body), "<<")

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM Positions WHERE UUID IN (%s)", string(body)))
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Positions remove failed:", err)

	Log.Debug("Positions UUIDs >>", string(body), "<< successfully removed.")
}

// inventoryType is a struct to hold data retrieved from the database, used by several functions (including JSON).
type inventoryType struct {
	UUID zero.String
	Name zero.String
	Type zero.String
	LastUpdate zero.String
	Permissions zero.String
}

// uiInventory creates a JSON representation of the Inventory table and spews it out.
func uiInventory(w http.ResponseWriter, r *http.Request) {
	var (
		rowArr []interface{}
		Inventory inventoryType
	)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanic(err)

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
	defer rows.Close()

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	// return
}

// uiInventoryUpdate receives a JSON representation of one row (from the agGrid) in order to update our database.
func uiInventoryUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	var inv inventoryType
    err = json.Unmarshal(body, &inv)
    checkErrPanic(err)

    db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	stmt, err := db.Prepare("REPLACE INTO Inventory (`UUID`, `Name`, `Type`, `LastUpdate`, `Permissions`) VALUES (?,?,?,?,?)")
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace prepare failed:", err)

	defer stmt.Close()

	_, err = stmt.Exec(inv.UUID, inv.Name, inv.Type, inv.LastUpdate, inv.Permissions)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace exec failed:", err)

	// return
}

// uiInventoryRemove receives a list of UUIDs to remove from the Inventory table.
func uiInventoryRemove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	// Log.Debug("\nInventory Body is >>", string(body), "<<")

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM Inventory WHERE UUID IN (%s)", string(body)))
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Inventory remove failed:", err)

	Log.Debug("Inventory UUIDs >>", string(body), "<< successfully removed.")
}


// userManagementType is a struct to hold data retrieved from the database, used by several functions (including JSON).
type userManagementType struct {
	Email zero.String
	Password zero.String
}

// uiUserManagement creates a JSON representation of the Users table and spews it out.
func uiUserManagement(w http.ResponseWriter, r *http.Request) {
	var (
		rowArr []interface{}
		UserManagement userManagementType
	)

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanic(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT * FROM Users")
	checkErrPanic(err)

	for rows.Next() {
		err = rows.Scan(
			&UserManagement.Email,
			&UserManagement.Password,
		)
		rowArr = append(rowArr, UserManagement)
	}
	checkErr(err)
	defer rows.Close()

	if data, err := json.MarshalIndent(rowArr, "", " "); err != nil {
		checkErr(err)
	} else {
		_, err := fmt.Fprintf(w, "%s", data)
		checkErr(err)
	}
	// return
}

// uiUserManagementUpdate receives a JSON representation of one row (from the agGrid) in order to update our database.
func uiUserManagementUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	var user userManagementType
    err = json.Unmarshal(body, &user)
    checkErrPanic(err)

    db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	stmt, err := db.Prepare("REPLACE INTO Users (`Email`, `Password`) VALUES (?,?)")
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace prepare failed:", err)

	defer stmt.Close()

	_, err = stmt.Exec(user.Email, user.Password)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Replace exec failed:", err)

	// return
}

// uiUserManagementRemove receives a list of UUIDs to remove from the UserManagement table.
func uiUserManagementRemove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
    checkErrPanic(err)
	// Log.Debug("\nInventory Body is >>", string(body), "<<")

	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Connect failed:", err)

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DELETE FROM Users WHERE Email IN (%s)", string(body)))
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Users remove failed:", err)

	Log.Debug("User(s) Email(s) >>", string(body), "<< successfully removed.")
}
