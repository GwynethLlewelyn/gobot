// This deals with calls coming from Second Life or OpenSimulator
// it"s essentially a RESTful thingy
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
	var signature string
	
	if r.Form["signature"] != nil && r.Header["HTTP_X_SECONDLIFE_OBJECT_KEY"] != nil {
		signature = GetMD5Hash(r.Header.Get("HTTP_X_SECONDLIFE_OBJECT_KEY") + r.Form.Get("timestamp") + ":9876")
						
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
		
		stmt, err := db.Prepare("REPLACE INTO Inventory SET `UUID` = ?, `Name` = ?, `Type` = ?, `Permissions` = ?, `LastUpdate` = ?");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}		
		
		_, err = stmt.Exec(r.Header.Get("HTTP_X_SECONDLIFE_OBJECT_KEY"), r.Form.Get("name"),
							r.Form.Get("itemType"), r.Form.Get("permissions"), r.Form.Get("timestamp"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace exec failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}

		//_, err := res.RowsAffected()
		//checkErr(err)
		
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s successfully updated!", r.Header.Get("HTTP_X_SECONDLIFE_OBJECT_KEY"))
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
    fmt.Println(r.Form["url_long"])
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
	if r.Header["HTTP_X_SECONDLIFE_OBJECT_KEY"] != nil {
		
		// open database connection and see if we can update the inventory for this object
		db, err := sql.Open(PDO_Prefix, SQLiteDBFilename)
		
		if err != nil {
			http.Error(w, fmt.Sprintf("Connect failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}
		
		defer db.Close()
		
		stmt, err := db.Prepare("REPLACE INTO Obstacles SET `UUID` = ?, `Name` = ?, `BotKey` = ?, `BotName` = ?, `Type` = ?, `Origin` = ?, `Position` = ?, `Rotation` = ?, `Velocity` = ?, `Phantom` = ?, `Prims` = ?, `BBHi` = ?, `BBLo` = ?, `LastUpdate` = ?");
		if err != nil {
			http.Error(w, fmt.Sprintf("Replace prepare failed: %s\n", err), http.StatusServiceUnavailable)
			return
		}		
		
		_, err = stmt.Exec(
			r.Form.Get("key"),
			r.Form.Get("name"),
			r.Header.Get("HTTP_X_SECONDLIFE_OBJECT_KEY"), 
			r.Header.Get("HTTP_X_SECONDLIFE_OWNER_NAME"), 
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
		
		reply := r.Header.Get("HTTP_X_SECONDLIFE_OWNER_NAME") + " sent us:\n" +
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

// registerPosition is a mystery even to me
func registerPosition(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented) 
	return
}

// registerAgent is probably useful
func registerAgent(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented) 
	return
}

// configureCube is a miracle how it works
func configureCube(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented) 
	return
}