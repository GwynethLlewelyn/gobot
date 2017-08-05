// Functions to deal with the backoffice.
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"html/template"
	"github.com/gorilla/securecookie"
	"bytes"
	"strconv"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"strings"
	"log"
	// "gopkg.in/guregu/null.v3/zero"
	"github.com/heatxsink/go-gravatar"
)

// GobotTemplatesType expands on template.Template.
//  need to expand it so I can add a few more methods here 
type GobotTemplatesType struct{
	template.Template
}

// GobotTemplates stores all parsed templates for the backoffice.
var GobotTemplates GobotTemplatesType

// init parses all templates and puts it inside a (global) var.
//  This is supposed to be called just once! (in func main())
func (gt *GobotTemplatesType)init(globbedPath string) error {
	temp, err := template.ParseGlob(globbedPath)
	checkErr(err) // move out later, we just need it here to check what's wrong with the templates (20170706)
	gt.Template = *temp;
	return err
}

// gobotRenderer assembles the correct templates together and executes them.
//  this is mostly to deal with code duplication 
func (gt *GobotTemplatesType)gobotRenderer(w http.ResponseWriter, r *http.Request, tplName string, tplParams templateParameters) error {
	thisUserName :=  getUserName(r)
	
	// add cookie to all templates
	tplParams["SetCookie"] = thisUserName

	// add Gravatar to templates (note that all logins are supposed to be emails)
	
	// calculate hash for the Gravatar hovercard
	hasher := md5.Sum([]byte(thisUserName))
	hash := hex.EncodeToString(hasher[:])
	tplParams["GravatarHash"] = hash // we ought to cache this somewhere
	
	// deal with sizes, we want to have a specific size for the top menu
	var gravatarSize, gravatarSizeMenu = 32, 32

	// if someone set the sizes, then use them; if not, use defaults
	// note that this required type assertion since tplParams is interface{}
	// see https://stackoverflow.com/questions/14289256/cannot-convert-data-type-interface-to-type-string-need-type-assertion
	if tplParams["GravatarSize"] == nil {
		tplParams["GravatarSize"] = gravatarSize
	} else {
		gravatarSize = tplParams["GravatarSize"].(int)
	}
	if tplParams["GravatarSizeMenu"] == nil {
		tplParams["GravatarSizeMenu"] = gravatarSizeMenu
	} else {
		gravatarSizeMenu = tplParams["GravatarSizeMenu"].(int)
	}
	// for Retina displays; we could add a multiplication function for Go templates, but I'm lazy (20170706)
	tplParams["GravatarTwiceSize"] = 2 * gravatarSize
	tplParams["GravatarTwiceSizeMenu"] = 2 * gravatarSizeMenu
	
	// Now call the nice library function to get us the URL to the image, for the two sizes
	g := gravatar.New("identicon", gravatarSize, "g", true)
	tplParams["Gravatar"] = g.GetImageUrl(thisUserName) // we also ought to cache this somewhere
	
	g = gravatar.New("identicon", gravatarSizeMenu, "g", true)
	tplParams["GravatarMenu"] = g.GetImageUrl(thisUserName) // we also ought to cache this somewhere
	
	return gt.ExecuteTemplate(w, tplName, tplParams)
}

// Auxiliary functions for session handling
//  see https://mschoebel.info/2014/03/09/snippet-golang-webapp-login-logout/ (20170603)

var cookieHandler = securecookie.New(		// from gorilla/securecookie
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

// setSession returns a new session cookie with an encoded username.
func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:	"session",
			Value: encoded,
			Path:	"/",
		}
		// fmt.Println("Encoded cookie:", cookie)
		http.SetCookie(response, cookie)
	} else {
		fmt.Println("Error encoding cookie:", err)
	}
 }
 
 // getUserName sees if we have a session cookie with an encoded user name, returning nil if not found.
func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

// clearSession will remove a cookie by setting its MaxAge to -1 and clearing its value.
func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:	"session",
		Value:	 "",
		Path:	 "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// checkSession will see if we have a valid cookie; if not, redirects to login.
func checkSession(w http.ResponseWriter, r *http.Request) {
	// valid cookie and no errors?
	if getUserName(r) == "" {
		http.Redirect(w, r, URLPathPrefix + "/admin/login/", 302)	
	}
}


// Function handlers for HTTP requests (main functions for this file)

// backofficeMain is the main page, has some minor statistics, may do this fancier later on.
func backofficeMain(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r) // make sure we've got a valid cookie, or else send to login page
	// let's load the main template for now, just to make sure this works
	
	// Open database just to gather some statistics
	db, err := sql.Open(PDO_Prefix, GoBotDSN) // presumes sqlite3 for now
	checkErr(err)
	
	defer db.Close()

	var (
		cnt, obstacles, phantom int
		strAgents, strInventory, strPositions, strObstacles string
	)
		
	err = db.QueryRow("select count(*) from Agents").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strAgents = "Agents: " + strconv.Itoa(cnt)
	} else {
		strAgents = "No Agents."
	}
	
	err = db.QueryRow("select count(*) from Inventory").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strInventory = "Inventory items: " + strconv.Itoa(cnt)
	} else {
		strInventory = "No Inventory items."
	}

	err = db.QueryRow("select count(*) from Positions").Scan(&cnt)
	checkErr(err)	
	if (cnt != 0) {
		strPositions = "Positions: " + strconv.Itoa(cnt)
	} else {
		strPositions = "No Positions."
	}

	err = db.QueryRow("select count(*) from Obstacles").Scan(&obstacles)
	checkErr(err)		
	if (obstacles != 0) {
		strObstacles = "Obstacles: " + strconv.Itoa(obstacles)
	} else {
		strObstacles = "No Obstacles."
	}
	err = db.QueryRow("select count(*) from Obstacles where Phantom <> 1 AND Type <> 1").Scan(&phantom)
	checkErr(err)		
	if (phantom != 0) {
		strObstacles += " (" + strconv.Itoa(phantom) + " phantom)"
	}
	// Generate markers for the Leaflet-based map (20170605)
	
	// template: L.marker([127, 127], { title: 'Test' }).bindPopup(
	//								L.popup({ maxWidth: 180 })
	//									.setContent('Blah')
	//							).addTo(map);
								
	// First, get Agents (there are not many)
	var (
		Agent AgentType // this is defined on ui.go, ugh
		markersOutput string = ""
		xyz []string
		//position string
		coords string
	)
	
	rows, err := db.Query("SELECT * FROM Agents")
	checkErr(err)
	
	defer rows.Close()

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
		// do the magic to extract the actual coords
		coords = strings.Trim(*Agent.Position.Ptr(), "() \t\n\r")
		xyz = strings.Split(coords, ",")
		
		markersOutput += fmt.Sprintf("L.marker([%s, %s], { title: 'Agent: %s', riseOnHover: true, icon: agentMarker }).bindPopup(" +
										"L.popup({ maxWidth: 180 })" +
											".setContent('UUID: <a href=\"%s\">%s</a><br />Agent Name: %s<br />Position: %s')" +
										").addTo(map);", 
								xyz[0], xyz[1], *Agent.Name.Ptr(), URLPathPrefix + "/admin/agents/?UUID=" +
								*Agent.UUID.Ptr(), *Agent.UUID.Ptr(), *Agent.Name.Ptr(),
								*Agent.Position.Ptr())
	}
	checkErr(err)
	
	// now do positions
	var Position PositionType
	
	rows, err = db.Query("SELECT * FROM Positions")
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
		coords = strings.Trim(*Position.Position.Ptr(), "() \t\n\r")
		xyz = strings.Split(coords, ",")
		
		markersOutput += fmt.Sprintf("L.marker([%s, %s], { title: 'Position: %s', riseOnHover: true, icon: positionMarker }).bindPopup(" +
										"L.popup({ maxWidth: 180 })" +
											".setContent('UUID: <a href=\"%s\">%s</a><br />Position Name: %s<br />Position: %s')" +
										").addTo(map);", 
								xyz[0], xyz[1], *Position.Name.Ptr(),
								URLPathPrefix + "/admin/positions/?UUID=" + *Position.UUID.Ptr(),
								*Position.UUID.Ptr(), *Position.Name.Ptr(), *Position.Position.Ptr())
	}
	checkErr(err)

	// and at last add all the stupid obstacles...
	var Object ObjectType
	
	rows, err = db.Query("SELECT * FROM Obstacles WHERE Phantom <> 1 AND Type <> 1")
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
		coords = strings.Trim(*Object.Position.Ptr(), "() \t\n\r")
		xyz = strings.Split(coords, ",")
		
		markersOutput += fmt.Sprintf("L.marker([%s, %s], { title: 'Object: %s', riseOnHover: true, icon: objectMarker }).bindPopup(" +
										"L.popup({ maxWidth: 180 })" +
											".setContent('UUID: <a href=\"%s\">%s</a><br />Object Name: %s<br />Position: %s')" +
										").addTo(map);", 
								xyz[0], xyz[1], *Object.Name.Ptr(), URLPathPrefix + "/admin/objects/?UUID=" +
								*Object.UUID.Ptr(), *Object.UUID.Ptr(), *Object.Name.Ptr(), 
								*Object.Position.Ptr())
	}
	checkErr(err)

	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - main",
			"Agents": strAgents,
			"Inventory": strInventory,
			"Positions": strPositions,
			"Obstacles": strObstacles,
			"ObstaclePieChart": true,
			"obstaclesCnt": obstacles,
			"phantomCnt": phantom,
			"URLPathPrefix": URLPathPrefix,
			"MapURL": MapURL,
			"GravatarSize": 64,
			"MapMarkers": template.JS(markersOutput),
	}
	err = GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}

// backofficeAgents lists active agents.
func backofficeAgents(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - agents",
			"Content": "Hi there, this is the agents template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "agents.js",
	}
	err := GobotTemplates.gobotRenderer(w, r, "agents", tplParams)
	checkErr(err)
	return
}

// backofficeObjects lists objects seen as obstacles.
func backofficeObjects(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - objects",
			"Content": "Hi there, this is the objects template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "objects.js",
	}	
	err := GobotTemplates.gobotRenderer(w, r, "objects", tplParams)
	checkErr(err)
	return
}

// backofficePositions lists Positions.
func backofficePositions(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - positions",
			"Content": "Hi there, this is the positions template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "positions.js",
	}
	err := GobotTemplates.gobotRenderer(w, r, "positions", tplParams)
	checkErr(err)
	return
}

// backofficeInventory lists the content or inventory currently stored on objects.
func backofficeInventory(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - inventory",
			"Content": "Hi there, this is the inventory template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "inventory.js",
	}
	err := GobotTemplates.gobotRenderer(w, r, "inventory", tplParams)
	checkErr(err)
	return
}

// backofficeUserManagement deals with adding/removing application users. Just login(email) and password right now, no profiles, no email confirmations, etc. etc. etc.
//  This is basically a stub for more complex user management, to be reused by other developments...
//  I will not develop this further, except perhaps to link usernames to in-world avatars (may be useful)
func backofficeUserManagement(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - User Management",
			"Content": "Hi there, this is the User Management template",
			"URLPathPrefix": URLPathPrefix,
			"gobotJS": "user-management.js",
	}
	err := GobotTemplates.gobotRenderer(w, r, "user-management", tplParams)
	checkErr(err)
	return
}


// backofficeLogin deals with authentication.
func backofficeLogin(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Entered backoffice login for URL:", r.URL, "using method:", r.Method)
	if r.Method == "GET" {
		tplParams := templateParameters{ "Title": "Gobot Administrator Panel - login",
				"URLPathPrefix": URLPathPrefix,
		}
		err := GobotTemplates.gobotRenderer(w, r, "login", tplParams)
		checkErr(err)
	} else { // POST is assumed
		r.ParseForm()
        // logic part of logging in
        email		:= r.Form.Get("email")
        password	:= r.Form.Get("password")
        
        //fmt.Println("email:", email)
        //fmt.Println("password:", password)
        
        if email == "" || password == "" { // should never happen, since the form checks this
	        http.Redirect(w, r, URLPathPrefix + "/", 302)        
        }
        
        // Check username on database
        db, err := sql.Open(PDO_Prefix, GoBotDSN)
		checkErr(err)
		
		defer db.Close()
	
		// query
		rows, err := db.Query("SELECT Email, Password FROM Users")
		checkErr(err)

		defer rows.Close()
		
		var (
			Email string
			Password string
		)
	 
		// enhash the received password; I just use MD5 for now because there is no backoffice to create
		//  new users, so it's easy to generate passwords manually using md5sum;
		//  however, MD5 is not strong enough for 'real' applications, it's just what we also use to
		//  communicate with the in-world scripts (20170604)
		pwdmd5 := fmt.Sprintf("%x", md5.Sum([]byte(password))) //this has the hash we need to check
	  
		authorised := false // outside of the for loop because of scope
	
		for rows.Next() {	// we ought just to have one entry, but...
			_ = rows.Scan(&Email, &Password)
			// ignore errors for now, either it checks true or any error means no authentication possible
			if Password == pwdmd5 {
				authorised = true
				break
			}		
		}
		
	    if authorised {
	        // we need to set a cookie here
	        setSession(email, w)
	        // redirect to home
	        http.Redirect(w, r, URLPathPrefix + "/admin", 302)
		} else {
			// possibly we ought to give an error and then redirect, but I don't know how to do that (20170604)
			http.Redirect(w, r, URLPathPrefix + "/", 302) // will ask for login again
		}
		return
	}
}

// backofficeLogout clears session and returns to login prompt.
func backofficeLogout(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, URLPathPrefix + "/", 302)
}

// backofficeCommands is a form-based interface to give commands to individual bots.
func backofficeCommands(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	// Collect a list of existing bots and their PermURLs for the form
	
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErr(err)

	// query
	rows, err := db.Query("SELECT Name, PermURL FROM Agents ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var name, permURL, AvatarPermURLOptions = "", "", ""

	// find all agent (NPC) Names and PermURLs and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&name, &permURL)
		checkErr(err)
		AvatarPermURLOptions += "\t\t\t\t\t\t\t\t\t\t\t<option value=\"" + permURL + "\">" + name + "|" + permURL + "</option>\n"
	}
	
	db.Close()

	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - commands",
			"PanelHeading": "Select your command",
			"URLPathPrefix": URLPathPrefix,
			"AvatarPermURLOptions": template.HTML(AvatarPermURLOptions), // trick to get valid HTML not to be escaped by the Go template engine
	}
	err = GobotTemplates.gobotRenderer(w, r, "commands", tplParams)
	checkErr(err)
	return
}

// backofficeCommandsExec gets the user-selected params from the backofficeCommands form and sends them to the user, giving feedback.
//  This may change in the future, e.g. using Ajax to get inline results on the form.
func backofficeCommandsExec(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	err := r.ParseForm()
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Extracting parameters failed: %s\n", err)

	var content = ""
	
	// test: just gather the values from the form, to make sure it works properly
	for key, values := range r.Form {   // range over map
		for _, value := range values {    // range over []string
			content += "<b>" + key + "</b> -> " + value + "<br />"
  		}
	}
	content += "<p></p><h3>In-world results:</h3>"
	
	// prepare the call to the agent (OpenSimulator NPC)
    body := "command=" + r.Form.Get("command") + "&" + 
    	r.Form.Get("param1") + "=" + r.Form.Get("data1") + "&" +
    	r.Form.Get("param2") + "=" + r.Form.Get("data2")
    	
    rsBody, err := callURL(r.Form.Get("PermURL"), body)
    if (err != nil) {
	    content += "<p class=\"text-danger\">" + rsBody + "</p>"
    } else {
	    content += "<p class=\"text-success\">" + rsBody + "</p>"
    }
    
    log.Printf("Sending to in-world object %s ... %s\n", r.Form.Get("PermURL"), body) // debug
    
    /*
    rs, err := http.Post(r.Form.Get("PermURL"), "application/x-www-form-urlencoded", bytes.NewBuffer(body))
    // Code to process response (written in Get request snippet) goes here

	defer rs.Body.Close()
	
	rsBody, err := ioutil.ReadAll(rs.Body)
	if (err != nil) {
		errMsg := fmt.Sprintf("Error response from in-world object: %s", err)
		log.Println(errMsg)
		content += "<p class=\"text-danger\">" + errMsg + "</p>"
	} else {
	    log.Printf("Reply from in-world object %s\n", rsBody)
		content += "<p class=\"text-success\">" + string(rsBody) + "</p>"
	}
	*/
	
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - Commands Exec Result",
		"Preamble": template.HTML("<p>Results coming from in-world object:</p>"),
		"Content": template.HTML(content),
		"URLPathPrefix": URLPathPrefix,
		"ButtonText": "Another command",
		"ButtonURL": "/admin/commands/",
	}
	err = GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}

// backofficeControllerCommands is a form-based interface to give commands to the Bot Controller.
func backofficeControllerCommands(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	// Collect a list of existing bots and their PermURLs for the form
	
	db, err := sql.Open(PDO_Prefix, GoBotDSN)
	checkErr(err)

	// query for in-world objects that are Bot Controllers
	rows, err := db.Query("SELECT Name, Location, Position, PermURL FROM Positions WHERE ObjectType ='Bot Controller' ORDER BY Name")
	checkErr(err)
	
	defer rows.Close()
 	
	var name, location, position, permURL, MasterBotControllers, regionName, coords = "", "", "", "", "", "", ""
	var xyz []string

	// As on backofficeCommands, but a little more complicated
	for rows.Next() {
		err = rows.Scan(&name, &location, &position, &permURL)
		checkErr(err)
		// parse name of the region and coordinates
		regionName = location[:strings.Index(location, "(")-1]
		coords = strings.Trim(position, "() \t\n\r")
		xyz = strings.Split(coords, ",")
		
		MasterBotControllers += fmt.Sprintf("\t\t\t\t\t\t\t\t\t\t\t<option value=\"%s\">%s [%s (%s,%s,%s)]</option>\n", permURL, name, regionName, xyz[0], xyz[1], xyz[2])
	}

	rows, err = db.Query("SELECT Name, OwnerKey FROM Agents ORDER BY Name")
	checkErr(err)
	
	// defer rows.Close()
 	
	var ownerKey, AgentNames = "", "" // we're reusing 'name' from above

	// find all Names and OwnerKeys and create select options for each of them
	for rows.Next() {
		err = rows.Scan(&name, &ownerKey)
		checkErr(err)
		AgentNames += "\t\t\t\t\t\t\t\t\t\t\t<option value=\"" + ownerKey + "\">" + name + " (" + ownerKey + ")</option>\n"
	}
	
	db.Close()

	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - Bot Controller Commands",
			"PanelHeading": "Select your Bot Controller and give it a command",
			"URLPathPrefix": URLPathPrefix,
			"MasterBotControllers": template.HTML(MasterBotControllers),
			"AgentNames": template.HTML(AgentNames),
	}
	err = GobotTemplates.gobotRenderer(w, r, "controller-commands", tplParams)
	checkErr(err)
	return
}

// backofficeControllerCommandsExec gets the user-selected params from the backofficeControllerCommands form and sends them to the user, giving feedback.
//  This may change in the future, e.g. using Ajax to get inline results on the form.
func backofficeControllerCommandsExec(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	err := r.ParseForm()
	checkErrPanicHTTP(w, http.StatusServiceUnavailable, funcName() + ": Extracting parameters failed: %s\n", err)
	
	var content = ""
	
	// test: just gather the values from the form, to make sure it works properly
	for key, values := range r.Form {   // range over map
		for _, value := range values {    // range over []string
			content += "<b>" + key + "</b> -> " + value + "<br />"
  		}
	}
	content += "<p></p><h3>In-world results:</h3>"
	
	// prepare the call to the in-world Bot Controller
	//  HTTP request as per http://moazzam-khan.com/blog/golang-make-http-requests/
    body := []byte("npc=" + r.Form.Get("NPC") + "&" +
    	"command=" + r.Form.Get("command") + "&" +
    	r.Form.Get("param1") + "=" + r.Form.Get("data1") + "&" +
    	r.Form.Get("param2") + "=" + r.Form.Get("data2"))
    
    fmt.Printf("Sending to agent %s via Bot Controller %s ... %s\n", r.Form.Get("NPC"),
    	r.Form.Get("PermURL"), body)
    
    rs, err := http.Post(r.Form.Get("PermURL"), "application/x-www-form-urlencoded", bytes.NewBuffer(body))
    // Code to process response (written in Get request snippet) goes here
    checkErr(err)

	defer rs.Body.Close()
	
	rsBody, err := ioutil.ReadAll(rs.Body)
	if (err != nil) {
		errMsg := fmt.Sprintf("Error response from in-world object: %s", err)
		log.Println(errMsg)
		content += "<p class=\"text-danger\">" + errMsg + "</p>"
	} else {
	    log.Printf("Reply from in-world object %s\n", rsBody)
		content += "<p class=\"text-success\">" + string(rsBody) + "</p>"
	}
	
	tplParams := templateParameters{ "Title": "Gobot Administrator Panel - Controller Commands Exec Result",
		"Preamble": template.HTML("<p>Results coming from in-world object:</p>"),
		"Content": template.HTML(content),
		"URLPathPrefix": URLPathPrefix,
		"ButtonText": "Another controller command",
		"ButtonURL": "/admin/controller-commands/",
	}
	err = GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}

// backofficeLSLRegisterObject creates a LSL script for registering cubes, using the defaults set by the user.
//  This is better than using 'template' LSL scripts which people may fill in wrongly, this way at least
//   we won't get errors about wrong signature PIN or hostnames etc.
func backofficeLSLRegisterObject(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot LSL Generator - register object.lsl",
			"URLPathPrefix": URLPathPrefix,
//			"LSLRegisterObject": true,
			"Host": Host,
			"ServerPort": ServerPort,
			"LSLSignaturePIN": LSLSignaturePIN,
			"LSL": "lsl-register-object", // this will change some formatting on the 'main' template (20170706)
	}
	// check if we have a frontend (it's configured on the config.toml file); if no, use the ServerPort
	//  the 'frontend' will be nginx, Apache, etc. to cache replies from Go and serve static files from port 80 (20170706)
	if FrontEnd == "" {
		tplParams["ServerPort"] = ServerPort
	}
	err := GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}

// backofficeLSLBotController creates a LSL script for the Master Bot Controller.
//  Note that it will also deal with deleting agents
func backofficeLSLBotController(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot LSL Generator - bot controller.lsl",
			"URLPathPrefix": URLPathPrefix,
//			"LSLBotController": true,
			"Host": Host,
			"ServerPort": ServerPort,
			"LSLSignaturePIN": LSLSignaturePIN,
			"LSL": "lsl-bot-controller",
	}
	if FrontEnd == "" {
		tplParams["ServerPort"] = ServerPort
	}
	err := GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}

// backofficeLSLAgentScripts creates 3 scripts to be placed inside a transparent box attached to the agent's avatar.
//  These will be used to register the avatar, allow the agent to receive commands, and deal with sensors and llCastRay() detection.
func backofficeLSLAgentScripts(w http.ResponseWriter, r *http.Request) {
	checkSession(w, r)
	tplParams := templateParameters{ "Title": "Gobot LSL Generator - Agent scripts",
			"URLPathPrefix": URLPathPrefix,
//			"LSLAgentScripts": true,
			"Host": Host,
			"ServerPort": ServerPort,
			"LSLSignaturePIN": LSLSignaturePIN,
			"LSL": "lsl-agent-scripts",
	}
	if FrontEnd == "" {
		tplParams["ServerPort"] = ServerPort
	}
	err := GobotTemplates.gobotRenderer(w, r, "main", tplParams)
	checkErr(err)
	return
}