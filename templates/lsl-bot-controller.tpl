{{ define "lsl-bot-controller" }}
<p>This is the master bot controller script, which allows controlling all agents, including creating new ones.</p>
<p>Copy the below code to a script called <code>bot controller.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
// Handles registration with the external database
// Send inventory and deal with external commands

string registrationURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-position/";
string externalURL; // this is what we'll get from SL to get incoming connections
string webServerURLupdateInventory = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-inventory/";
key registrationRequest;    // used to track down the request for registration
key updateRequest;    // used to track down the request for registration
key serverKey; // for inventory updates
key httpRequestKey;
integer LSLSignaturePIN = {{.LSLSignaturePIN}};

init()
{
    llSetText("Registering bot controller...", <1.0,0.0,0.0>, 1.0);
    llOwnerSay("Registering bot controller...");
    llSetTimerEvent(0.0);
    
    // release URLs before requesting a new one
    llReleaseURL(externalURL);
    externalURL = "";            
    llRequestURL();
}

default
{
    state_entry()
    {
        llOwnerSay("On state_entry");
        llSetText("Listening on 10", <0, 255, 0>, 1);
        llSetTimerEvent(0.0);
    }
    
    on_rez(integer what)
    {
        llOwnerSay("On on_rez");
        init();
    }

    touch_start(integer total_number)
    {
         llOwnerSay("On touch_start");
        // just re-register
        
        if (llDetectedKey(0) == llGetOwner() || llDetectedGroup(0))
        {
            if (llDetectedTouchFace(0) != 0)
                init();
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
                + "&objecttype=" + llEscapeURL("Bot Controller")
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
            // incoming request for bot to do things
            llSay(0, "[Request from server:] " + body);
            
            list params = llParseStringKeepNulls(llUnescapeURL(body), ["&", "="], []);
            string response; // what we return
            
            // first parameter will always be be npc=<UUID>
            
            key NPC = llList2String(params, 1);
            if (osIsNpc(NPC))
                llSay(0, "Sanity check: This is an NPC with key " + (string)NPC);
            else
                llSay(0, "Sanity check failed: Key " + (string)NPC + " is NOT an NPC");
            
            // llOwnerSay("List parsed: " + (string) params);
            
            // commands begin on the second parameter
            string commandTag = llList2String(params, 2);
            string command = llList2String(params, 3);
            
            if (commandTag == "command")
            {
                if (command == "osNpcGetRot")
                {
                    response = (string)osNpcGetRot(NPC);
                }
                else if (command == "osNpcSetRot")
                {
                    osNpcSetRot(NPC, llList2Rot(params, 5));
                    response = "Rotation set.";
                }
                else if (command == "osNpcGetPos")
                {
                    response = (string)osNpcGetPos(NPC);
                }
                else if (command == "osNpcGetOwner")
                {
                    response = (string)osNpcGetOwner(NPC);
                }
                else if (command == "osNpcMoveToTarget")
                {
                    osNpcMoveToTarget(NPC, llList2Vector(params, 5), llList2Integer(params, 7));
                    response = "Moving to target " + llList2String(params, 3); 
                }
                else if (command == "osNpcStopMoveToTarget")
                {
                    osNpcStopMoveToTarget(NPC);
                    response = "Stopping.";
                }
                else if (command == "osNpcSit")
                {
                    osNpcSit(NPC, llList2Key(params, 5), llList2Integer(params, 7));
                    response = "Sitting on " +
                        llKey2Name(llList2Key(params, 5))
                        + " (" + llList2Key(params, 5) + ").";
                }
                else if (command == "osNpcStand")
                {
                    osNpcStand(NPC);
                    response = "Standing up.";
                }
                else if (command == "osNpcSay")
                {
                    osNpcSay(NPC, llList2Integer(params, 5), llUnescapeURL(llList2String(params, 7)));
                    response = "Saying \"" +
                        llUnescapeURL(llList2String(params, 7)) + "\".";
                }
                else if (command == "osNpcShout")
                {
                    osNpcShout(NPC, llList2Integer(params, 5), llUnescapeURL(llList2String(params, 7)));
                    response = "Shouting \"" +
                        llUnescapeURL(llList2String(params, 7)) + "\".";
                }
                else if (command == "osNpcWhisper")
                {
                    osNpcWhisper(NPC, llList2Integer(params, 5), llUnescapeURL(llList2String(params, 7)));
                    response = "Whispering \"" +
                        llUnescapeURL(llList2String(params, 7)) + "\".";
                }
                else if (command == "osNpcPlayAnimation")
                {
                    osNpcPlayAnimation(NPC, llList2String(params, 5));
                    response = "Playing animation \"" +
                        llList2String(params, 5) + "\".";
                }               
                else if (command == "osNpcStopAnimation")
                {
                    osNpcStopAnimation(NPC, llList2String(params, 5));
                    response = "Stopping animation \"" +
                        llList2String(params, 5) + "\".";
                }               
                else if (command == "osNpcLoadAppearance")
                {
                    osNpcLoadAppearance(NPC, llList2String(params, 5));
                    response = "Loading appearance \"" +
                        llList2String(params, 5) + "\".";
                }
                else if (command == "osNpcTouch")
                {
                    osNpcTouch(NPC, llList2Key(params, 5), llList2Integer(params, 7));
                    response = "Touching " + llKey2Name(llList2Key(params, 5))
                        + " (" + llList2Key(params, 5) + ").";
                }
                else if (command == "osNpcCreate")
                {
                    string FirstName = "My";
                    string LastName = "Bot";
                    string fullName = llList2String(params, 5);
                    integer index = llSubStringIndex(fullName, " ");
                    if (~index)
                        FirstName = llDeleteSubString(fullName, index, -1);
                    else FirstName = fullName;
                    
                    if (~index)
                        LastName = llDeleteSubString(fullName, 0, index);
                    else LastName = "Bot";
                    
                    // add some randomness
                    float xRand = llFrand(30) - 30;
                    float yRand = llFrand(30) - 30;
                    vector pos = llGetPos();
                    
                    key newNPC = osNpcCreate(FirstName, LastName, pos + &lt;xRand, yRand, 0&gt;, llList2String(params, 7), OS_NPC_SENSE_AS_AGENT);
llSleep(1);
                    rotation npcCurrRot = osNpcGetRot(newNPC);
                    float angle = llAtan2( llVecMag(pos % (pos + &lt;xRand, yRand, 0&gt;)), pos * (pos + &lt;xRand, yRand, 0&gt;) );
                    rotation npcNewRot = npcCurrRot * llEuler2Rot&lt;0.0, 0.0, angle&gt;);
                    osNpcSetRot(newNPC, npcNewRot);

                    response = "New NPC: " + (string)newNPC;
                }
                else if (command == "osNpcRemove")
                {
                    osNpcRemove(NPC);
                    response = "Removing " + llKey2Name(NPC);
                }
                else
                {
                    response = "";
                    llHTTPResponse(id, 405, "Unknown engine command " + command + ".");
                }
            }
            
            if (response) 
            {
                llSay(0, "Sending back response for " + 
                    command + " '" +
                    response + "'...");
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
        if (c & (CHANGED_REGION | CHANGED_OWNER | CHANGED_REGION_START | CHANGED_TELEPORT ) )
        {
            init();
        }
        // Deal with inventory changes
        else if (c & CHANGED_INVENTORY)
            state read_inventory;
    }
    
    timer()
    {
        llSetText("Timed out, trying again to\nregister bot controller...", <1.0,0.0,0.0>, 1.0);
        llOwnerSay("Timed out, trying again to register bot controller...");
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
        
        llSetTimerEvent(3600.0); // timeout if the web server is too slow in responding
        
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
        state default;
    }
    
    state_exit()
    {
        llOwnerSay("state_exit inventory");
        llSetTimerEvent(0.0);
    }
}
</code></pre>
{{ end }}