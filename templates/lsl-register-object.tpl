{{ define "lsl-register-object" }}
<p>Copy the below code to a script called <code>GoBot Registering Button.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
// Handles registration with the external database
// Send current configuration for class, type, and rates for energy/money/happiness
// To-do: Accepts remote calls to trigger animations, send IMs to people etc

string registrationURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-position/";
string externalURL; // this is what we'll get from SL to get incoming connections
string webServerURLupdateInventory = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-inventory/";
key registrationRequest;    // used to track down the request for registration
key updateRequest;    // used to track down the request for registration
key serverKey; // for inventory updates
key httpRequestKey;
integer LSLSignaturePIN = {{.LSLSignaturePIN}};
string type = "money";
string class = "peasant";
float rateEnergy;   // cube configuration parameters; touch on the face
float rateMoney;
float rateHappiness;
vector touchST; // for the sliders

init()
{
    llSetText("Registering position...", <1.0,0.0,0.0>, 1.0);
    llOwnerSay("Registering position...");
    // parse description field
    parseDescription();
    
    // release URLs before requesting a new one
    llReleaseURL(externalURL);
    externalURL = "";
    llRequestURL();
}

// parse description field, which contains the type of box and the class it applies to
parseDescription()
{
    list params = llParseString2List(llGetObjectDesc(), [";"], []);
    
    type =          llList2String(params, 0);
    class =         llList2String(params, 1);
    rateEnergy =    llList2Float(params, 2);
    rateMoney =     llList2Float(params, 3);
    rateHappiness = llList2Float(params, 4);
    updateSetText();
}

// update settext with energy, money, happiness
updateSetText()
{
    llSetText("Energy: " + (string)rateEnergy +
            "\nMoney: " + (string)rateMoney +
            "\nHappiness: " + (string)rateHappiness, <1.0,1.0,1.0>, 1.0);
}

default
{
    state_entry()
    {
        parseDescription();
    }    

    on_rez(integer what)
    {
        init();
    }

    touch_start(integer total_number)
    {
        // this has to be redone: touching will allow to set the cube's class, type, etc.
        
        if (llDetectedKey(0) == llGetOwner() || llDetectedGroup(0))
        {
            updateSetText();
            if (llDetectedTouchFace(0) != 0)
                init();
        }
    }
    
    touch(integer total_number)
    {
        if (llDetectedKey(0) == llGetOwner() || llDetectedGroup(0))
        {
            // Only face 0 is active for this
            if (llDetectedTouchFace(0) == -1)
                llWhisper(PUBLIC_CHANNEL, "Sorry, your viewer doesn't support touched faces.");
            else if (llDetectedTouchFace(0) == 0)
            {
                touchST = llDetectedTouchST(0);

                if (touchST != TOUCH_INVALID_TEXCOORD)
                {
                    // happiness: 0 <= 0.33
                    // money: > 0.33 <= 0.67
                    // energy: > 0.67 <= 1.0
                    
                    if (touchST.y <= 0.33)
                    {
                        rateHappiness = (2 * touchST.x) - 1.0;
                    }
                    else if (touchST.y > 0.33 && touchST.y <= 0.67)
                    {
                        rateMoney = (2 * touchST.x) - 1.0;
                    }
                    else if (touchST.y > 0.67)
                    {
                        rateEnergy = (2 * touchST.x) - 1.0;
                    }
                }
            }
            updateSetText();
        }
    }
    
    touch_end(integer who)
    {
        if ((llDetectedKey(0) == llGetOwner() || llDetectedGroup(0)) && llDetectedTouchFace(0) == 0)
        {
         // save to description
            llSetObjectDesc(type + ";" + class + ";" + (string)rateEnergy + ";" + (string)rateMoney
                + ";" + (string)rateHappiness);
            
            // if permURL is empty, do a full registration
            if (externalURL == "")
            {
                init();
            }
            else 
            {
	            llInstantMessage(llDetectedKey(0), "Updating information...");
				string myTimestamp = llGetTimestamp();
                updateRequest = llHTTPRequest(registrationURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"],
                	"permURL=" + llEscapeURL(externalURL)
                    + "&objecttype=" + llEscapeURL(type)
                    + "&objectclass=" + llEscapeURL(class)
                    + "&rateenergy=" + llEscapeURL((string)rateEnergy)
                    + "&ratemoney=" + llEscapeURL((string)rateMoney)
                    + "&ratehappiness=" + llEscapeURL((string)rateHappiness)
                    + "&amp;timestamp=" + myTimestamp
                    + "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
            }
        }           
    }
    
    http_response(key request_id, integer status, list metadata, string body)
    {
        if (request_id == registrationRequest || request_id == updateRequest)
        {
            if (status == 200)
            {           
                llOwnerSay(body);
                // new registration? switch to inventory reading
                if (request_id == registrationRequest)
                    state read_inventory;
                // if it's just an update, no need to do anything else for now
                updateSetText();
            }
            else
            {
                llSetText("!!! BROKEN !!!", <1.0,0.0,0.0>, 1.0);
                llOwnerSay("Error " +(string)status + ": " + body);
            }
        }
    }
    http_request(key id, string method, string body)
    {
        if (method == URL_REQUEST_GRANTED)
        {
            externalURL = body;
            
            string myTimestamp = llGetTimestamp();
            registrationRequest = llHTTPRequest(registrationURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"], 
                "permURL=" + llEscapeURL(externalURL)
                + "&objecttype=" + llEscapeURL(type)
                + "&objectclass=" + llEscapeURL(class)
                + "&rateenergy=" + llEscapeURL((string)rateEnergy)
                + "&ratemoney=" + llEscapeURL((string)rateMoney)
                + "&ratehappiness=" + llEscapeURL((string)rateHappiness)
                + "&amp;timestamp=" + myTimestamp
                + "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
            
            llSetTimerEvent(3600.0);    // if the registration fails, try later

        }
        else if (method == URL_REQUEST_DENIED)
        {
            llSetText("!!! BROKEN !!!", <1.0,0.0,0.0>, 1.0);
            llOwnerSay("Something went wrong, no url. " + body);
        }
        else if (method == "POST" || method == "GET")
        {
			// NOTE(gwyneth): This will allow the web-based interface to send commands to each and every object.
			//  Right now, it's just used to see if the objects are alive; this is needed by the Garbage Collector,
			//  it sends a 'command=ping' and expects a 'pong'; if not, it assumes that the object is 'dead' and cleans up. (20170729)

			list params = llParseStringKeepNulls(llUnescapeURL(body), ["&", "="], []);
			string response; // what we return

			string commandTag = llList2String(params, 0);
			string command = llList2String(params, 1);
			
			if (commandTag == "command")
			{
				if (command == "ping")
				{
					 response = "pong";
				}
				else
				{
					response = "";
					llHTTPResponse(id, 405, "Unknown engine command " + command + ".");
				}
			}
			
			if (response) 
			{
				//llSay(0, "Sending back response for " + 
				//	  command + " '" +
				//	  response + "'...");
				llHTTPResponse(id, 200, response);
			}
			else
				llSay(0, "ERROR: No response or no command found!");
        }       
        else
        {
            llHTTPResponse(id, 405, "Method unsupported");
        }
    }
        
    changed(integer c)
    {
        // Region changed, get a new PermURL
        if (c & (CHANGED_REGION | CHANGED_REGION_START | CHANGED_TELEPORT ) )
        {
            init();
        }
        // Deal with inventory changes
        else if (c & CHANGED_INVENTORY)
            state read_inventory;
    }
    
    timer()
    {
        llSetText("Timed out, trying again to\nregister position...", <1.0,0.0,0.0>, 1.0);
        llSetTimerEvent(0.0);
        init();
    }
}

state read_inventory
{
    state_entry()
    {
        llSetText("Sending to webserver - 0%", <0.3, 0.7, 0.2>, 1.0);
                // now prepare this line for sending to web server
        
        string httpBody;
        string itemName;
        string myTimeStamp;
        integer i;
        integer length = llGetInventoryNumber(INVENTORY_ALL);
        serverKey = llGetKey();
        
        llSetTimerEvent(360.00); // timeout if the web server is too slow in responding
        
        // Now add the new items.
        // This needs two passes: on the first one, we'll skip textures
        // The second pass will add them later
        llSetText("Checking inventory...", <1.0,1.0,0.0>, 1.0);
        
        for (i = 0; i < length; i++)
        {
            itemName = llGetInventoryName(INVENTORY_ALL, i);
            
            if (llGetInventoryType(itemName) != INVENTORY_SCRIPT && llGetInventoryType(itemName) != INVENTORY_TEXTURE) // skip script, skip textures
            {
                myTimeStamp = llGetTimestamp();
                
                httpBody =  "name=" + llEscapeURL(itemName) + 
                            "&amp;timestamp=" + myTimeStamp +
                            "&permissions=" + (string) llGetInventoryPermMask(itemName, MASK_NEXT) +
                            "&itemType=" + (string) llGetInventoryType(itemName) +
                            "&signature=" + llMD5String((string)serverKey + myTimeStamp, LSLSignaturePIN);
                llSleep(1.0);
     
                httpRequestKey = llHTTPRequest(webServerURLupdateInventory,         
                                [HTTP_METHOD, "POST",
                                 HTTP_MIMETYPE,"application/x-www-form-urlencoded"], 
                                httpBody);
                //llOwnerSay("Object " + (string) i + ": " + httpBody);
                if (httpRequestKey == NULL_KEY) 
                    llOwnerSay("Error contacting webserver on item #" + (string)i);
            
                llSetText("Sending to webserver - " + (string) ((integer)((float)i/(float)length*100)) + "%", <0.3, 0.7, 0.2>, 1.0);
            }
        }
        state default;
    }
    
    http_response(key request_id, integer status, list metadata, string body)
    { 
        llSetText("", <0.0,0.0,0.0>, 1.0);
        
        if (request_id == httpRequestKey)
        {
            if (status != 200)
            {
                llOwnerSay("HTTP Error " + (string)status + ": " + body);
            }
            else 
            {
                llOwnerSay("Web-server reply: " + body); 
                if (body == "closed")
                    state default;
                              
            }
        }
    }    
    
    timer()
    {
        // HTTP server does not work, go to default state for now
        llOwnerSay("Web server did not reply after 3 minutes - not updated - try again later");
    }
}
</code></pre>
{{ end }}