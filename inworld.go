// This deals with calls coming from Second Life or OpenSimulator
// it's essentially a RESTful thingy
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
    "strings"
    "crypto/md5"
    "encoding/hex"
    //"log"
)

// GetMD5Hash takes a string which is to be encoded using MD5 and returns a string with the hex-encoded MD5 sum.
// Got this from https://gist.github.com/sergiotapia/8263278
func GetMD5Hash(text string) string {
    hasher := md5.New()
    hasher.Write([]byte(text))
    return hex.EncodeToString(hasher.Sum(nil))
}

// updateInventory updates the inventory of the object (object key will come in the headers)
func updateInventory(w http.ResponseWriter, r *http.Request) {
	// get all parameters in array
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Extracting parameters failed: %s\n", err), http.StatusServiceUnavailable)
		return
	}
	
	if r.Form.Get("signature") != "" && r.Header.Get("X-Secondlife-Object-Key") != "" {
		signature := GetMD5Hash(r.Header.Get("X-Secondlife-Object-Key") + r.Form.Get("timestamp") + ":9876")
						
		if signature != r.Form.Get("signature") {
			http.Error(w, "Signature does not match - hack attempt?", http.StatusServiceUnavailable) 
			return
		}
		
		// open database connection and see if we can update the inventory for this object
		db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
		
		if err != nil {
			http.Error(w, fmt.Sprintf("Connect failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		defer db.Close()
		
		stmt, err := db.Prepare("REPLACE INTO Inventory (`UUID`, `Name`, `Type`, `Permissions`, `LastUpdate`) VALUES (?,?,?,?,?)");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}		
		
		_, err = stmt.Exec(
			r.Header.Get("X-Secondlife-Object-Key"),
			r.Form.Get("name"),
			r.Form.Get("itemType"),
			r.Form.Get("permissions"),
			r.Form.Get("timestamp"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}

		//_, err := res.RowsAffected()
		//checkErr(err)
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s successfully updated!", r.Header.Get("X-Secondlife-Object-Key"))
		return
	} else {
		http.Error(w, "Signature not found", http.StatusServiceUnavailable) 
		return
	}
	
	/*
	fmt.Fprintf(w, "Root URL is: %s\n", updateInventory()) // send data to client side
	
    r.ParseForm()  // parse arguments, you have to call this by yourself
    fmt.Println(r.Form)  // print form information in server side
    fmt.Println("header connection: ", r.Header.Get("Connection"))
    fmt.Println("all headers:")
    for k, v := range r.Header {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url-long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
	if r.Form["signature"] != nil {
	    fmt.Fprintf(w, "Signature is %s\n", r.Form.Get("signature"))
    }
    */
}

// updateSensor updates the Obstacles database with an additional object found by the sensors
func updateSensor(w http.ResponseWriter, r *http.Request) {
	
	if r.Header.Get("X-Secondlife-Object-Key") != "" {
		// get all parameters in array
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("Extracting parameters failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		// open database connection and see if we can update the inventory for this object
		db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
		
		if err != nil {
			http.Error(w, fmt.Sprintf("Connect failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		defer db.Close()
		
		stmt, err := db.Prepare("REPLACE INTO Obstacles (`UUID`, `Name`, `BotKey`, `BotName`, `Type`, `Origin`, `Position`, `Rotation`, `Velocity`, `Phantom`, `Prims`, `BBHi`, `BBLo`, `LastUpdate`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}		
		
		_, err = stmt.Exec(
			r.Form.Get("key"),
			r.Form.Get("name"),
			r.Header.Get("X-Secondlife-Object-Key"), 
			r.Header.Get("X-Secondlife-Owner-Name"), 
			r.Form.Get("type"),
			r.Form.Get("origin"),
			strings.Trim(r.Form.Get("pos"), "<>()"),
			strings.Trim(r.Form.Get("rot"), "<>()"),
			strings.Trim(r.Form.Get("vel"), "<>()"),
			r.Form.Get("phantom"),
			r.Form.Get("prims"),
			strings.Trim(r.Form.Get("bbhi"), "<>()"),
			strings.Trim(r.Form.Get("bblo"), "<>()"),
			r.Form.Get("timestamp"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}

		//_, err := res.RowsAffected()
		//checkErr(err)
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		
		reply := r.Header.Get("X-Secondlife-Owner-Name") + " sent us:\n" +
		"Key: " + r.Form.Get("key") + " Name: " + r.Form.Get("name")+ "\n" +
		"Position: " + r.Form.Get("pos") + " Rotation: " + r.Form.Get("rot") + "\n" +
		"Type: " + r.Form.Get("type") + "\n" + "Origin: " + r.Form.Get("origin") + "\n" +
		"Velocity: " + r.Form.Get("vel") + 
		" Phantom: " + r.Form.Get("phantom") + 
		" Prims: " + r.Form.Get("prims") + "\n" +
		"BB high: " + r.Form.Get("bbhi") + 
		" BB low: " + r.Form.Get("bblo") + "\n" +
		"Timestamp: " + r.Form.Get("timestamp")
		
		fmt.Fprintf(w, "%s", reply)
		return
	} else {
		http.Error(w, "Not called from within the virtual world.", http.StatusMethodNotAllowed) 
		return
	}
}

// registerPosition saves a HTTP URL for a single object, making it persistent
// POST parameters:
//  permURL: a permanent URL from llHTTPServer 
//  signature: to make spoofing harder
//  timestamp: in-world timestamp retrieved with llGetTimestamp()
func registerPosition(w http.ResponseWriter, r *http.Request) {
	// get all parameters in array
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Extracting parameters failed: %s\n", err), http.StatusServiceUnavailable)
		return
	}
	
	if r.Header.Get("X-Secondlife-Object-Key") == "" {
		// fmt.Printf("Got '%s'\n", r.Header["X-Secondlife-Object-Key"])
		http.Error(w, "Not called from within the virtual world.", http.StatusMethodNotAllowed) 
		return		
	}
	
	if r.Form.Get("signature") != "" {
		// if we don't have the permURL to store, registering this object is pointless
		if r.Form["permURL"] == nil {
			http.Error(w, "No PermURL specified", http.StatusMethodNotAllowed) 
			return
		}
		
		signature := GetMD5Hash(r.Header.Get("X-Secondlife-Object-Key") + r.Form.Get("timestamp") + ":9876")
						
		if signature != r.Form.Get("signature") {
			http.Error(w, "Signature does not match - hack attempt?", http.StatusServiceUnavailable) 
			return
		}
		
		// open database connection and see if we can update the inventory for this object
		db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
		
		if err != nil {
			http.Error(w, fmt.Sprintf("Connect failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		defer db.Close()
		
		stmt, err := db.Prepare("REPLACE INTO Positions (`UUID`, `Name`, `PermURL`, `Location`, `Position`, `Rotation`, `Velocity`, `OwnerKey`, `OwnerName`, `ObjectType`, `ObjectClass`, `RateEnergy`, `RateMoney`, `RateHappiness`, `LastUpdate`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}		
		
		_, err = stmt.Exec(
			r.Header.Get("X-Secondlife-Object-Key"),
			r.Header.Get("X-Secondlife-Object-Name"),
			r.Form.Get("permURL"),
			r.Header.Get("X-Secondlife-Region"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Position"), "<>()"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Rotation"), "<>()"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Velocity"), "<>()"),
			r.Header.Get("X-Secondlife-Owner-Key"),
			r.Header.Get("X-Secondlife-Owner-Name"),
			r.Form.Get("objecttype"),
			r.Form.Get("objectclass"),
			r.Form.Get("rateenergy"),
			r.Form.Get("ratemoney"),
			r.Form.Get("ratehappiness"),
			r.Form.Get("timestamp"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}

		//_, err := res.RowsAffected()
		//checkErr(err)
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "'%s' successfully updated!", r.Header.Get("X-Secondlife-Object-Name"))
		fmt.Printf("These are the headers I got: %v\nAnd these are the parameters %v\n", r.Header, r.Form)
		return
	} else {
		http.Error(w, "Signature not found", http.StatusServiceUnavailable) 
		return
	}
}

// registerAgent saves a HTTP URL for a single agent, making it persistent
// POST parameters:
//  permURL: a permanent URL from llHTTPServer 
//  signature: to make spoofing harder
//  timestamp: in-world timestamp retrieved with llGetTimestamp()
//  request: currently only delete (to remove entry from database when the bot dies)
func registerAgent(w http.ResponseWriter, r *http.Request) {
	// get all parameters in array
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Extracting parameters failed: %s\n", err), http.StatusServiceUnavailable)
		return
	}
	
	if r.Header.Get("X-Secondlife-Object-Key") == "" {
		// fmt.Printf("Got '%s'\n", r.Header["X-Secondlife-Object-Key"])
		http.Error(w, "Only in-world requests allowed.", http.StatusMethodNotAllowed) 
		return		
	}
	
	if r.Form.Get("signature") == "" {
		http.Error(w, "Signature not found", http.StatusServiceUnavailable) 
		return
	}
	
	signature := GetMD5Hash(r.Header.Get("X-Secondlife-Object-Key") + r.Form.Get("timestamp") + ":9876")
						
	if signature != r.Form.Get("signature") {
		http.Error(w, "Signature does not match - hack attempt?", http.StatusServiceUnavailable) 
		return
	}
	
	// open database connection and see if we can update the inventory for this object
	db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Connect failed: %s\n", err), http.StatusServiceUnavailable)
		return
	}
	
	defer db.Close()

	if r.Form.Get("permURL") != "" { // bot registration
		stmt, err := db.Prepare("REPLACE INTO Agents (`UUID`, `Name`, `OwnerKey`, `OwnerName`, `PermURL`, `Location`, `Position`, `Rotation`, `Velocity`, `Energy`, `Money`, `Happiness`, `Class`, `SubType`, `LastUpdate`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}	
		
		_, err = stmt.Exec(
			r.Header.Get("X-Secondlife-Object-Key"),
			r.Header.Get("X-Secondlife-Object-Name"),
			r.Header.Get("X-Secondlife-Owner-Key"),
			r.Header.Get("X-Secondlife-Owner-Name"),
			r.Form.Get("permURL"),
			r.Header.Get("X-Secondlife-Region"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Position"), "<>()"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Rotation"), "<>()"),
			strings.Trim(r.Header.Get("X-Secondlife-Local-Velocity"), "<>()"),
			r.Form.Get("energy"),
			r.Form.Get("money"),
			r.Form.Get("happiness"),
			r.Form.Get("class"),
			r.Form.Get("subtype"),
			r.Form.Get("timestamp"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		replyText := "'" + r.Header.Get("X-Secondlife-Object-Name") +
				"' successfully updated object for NPC '" +
				r.Header.Get("X-Secondlife-Owner-Name") + "' (" +
				r.Header.Get("X-Secondlife-Owner-Key") + "), energy=" +
				r.Form.Get("energy") + ", money=" +
				r.Form.Get("money") + ", happiness=" +
				r.Form.Get("happiness") + ", class=" +
				r.Form.Get("class") + ", subtype=" +
				r.Form.Get("subtype") + "."
		
		fmt.Fprintf(w, replyText)
		fmt.Printf(replyText) // debug
	} else if r.Form.Get("request") == "delete" { // other requests, currently only deletion
		
		stmt, err := db.Prepare("DELETE FROM Agents WHERE UUID=?")
		if err != nil {
			http.Error(w, fmt.Sprintf("Delete agent prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		_, err = stmt.Exec(r.Form.Get("npc"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Delete agent exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "'%s' successfully deleted.", r.Form.Get("npc"))
		return
	}
}

// configureCube Support scripts for remote startup configuration for the cubes
//	This basically gives the lists of options (e.g. energy, happiness; classes of NPCs, etc.) so
//	that we don't need to hardcode them
func configureCube(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented) 
	return
}