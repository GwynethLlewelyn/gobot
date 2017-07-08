{{ define "lsl-bot-controller" }}
<p>Copy the below code to a script called <code>bot controller.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
// Handles agent registration with the external database
// Send current configuration for agent class, subtype, and energy/money/happiness
// On the first version it will only work in OpenSim and call osNPC functions

string registrationURL = "{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-agent/";
string externalURL; // this is what we'll get from SL to get incoming connections
key registrationRequest;    // used to track down the request for registration
key httpRequestKey;
integer LSLSignaturePIN = {{.LSLSignaturePIN}};
string class = "peasant";
string subtype = "publican";
string home = ""; // place where NPC will return when energy exhausted
float npcEnergy = 1.0;    // start with max energy and happiness
float npcMoney = 0.0; // but no money
float npcHappiness = 1.0;

init()
{
    llSetObjectName("Bot Controller - " + llKey2Name(llGetOwner())); 
    llSetText("Registering agent...", <1.0,0.0,0.0>, 1.0);
    llSay(0, "Registering agent...");
    // parse description field
    parseDescription();
    // make sure we release URLs before asking for a new one
    llReleaseURL(externalURL);
    externalURL = "";
    llRequestURL();
}

// parse description field, which contains the type of agent and the energy/money/happiness
parseDescription()
{
    list params = llParseString2List(llGetObjectDesc(), [";"], []);
    
    class =         llList2String(params, 0);
    subtype =       llList2String(params, 1);
    npcEnergy =     llList2Float(params, 2);
    npcMoney =      llList2Float(params, 3);
    npcHappiness =  llList2Float(params, 4);
    home =          llList2String(params, 5);
    updateSetText();
}

// save current parameters to object description and update floating text
setDescription()
{
    llSetObjectDesc(class + ";" + subtype + ";" +
        (string)npcEnergy + ";" +
        (string)npcMoney + ";" +
        (string)npcHappiness + ";" +
        home);
    updateSetText();
}

// update settext with energy, money, happiness
updateSetText()
{
    llSetText("Class: " + class +
            "\nSubType: " + subtype +
            "\nEnergy: " + (string)npcEnergy +
            "\nMoney: " + (string)npcMoney +
            "\nHappiness: " + (string)npcHappiness +
            "\nHome: " + home, <npcEnergy,npcMoney,npcHappiness>, 1.0);
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

    attach(key avatar)
    {
        if (NULL_KEY != avatar)
        {
            init();
        }
        else
        {
            // NPC probably died. We cannot do anything about it, specially if it comes alive again with a different UUID.
            llSay(0, "Register Agent: Going awaaaaaaay...");
        }
    }

    touch_start(integer total_number)
    {
        // possibly this will allow to reconfigure the bot
        
        /*if (llDetectedKey(0) == llGetOwner() || llDetectedGroup(0) || )
        {*/
            updateSetText();
            init();
        /*}
        else
        {
            llSay(0, "Sorry " + llDetectedName(0) + ", you cannot reset this bot!");
        }*/
    }
    
    
    http_response(key request_id, integer status, list metadata, string body)
    {
        if (request_id == registrationRequest)
        {
            if (status == 200)
            {      
                llSay(0, "Register Agent: [Registration request:] " + body);
            }
            else
            {
                llSetText("!!! BROKEN !!!", <1.0,0.0,0.0>, 1.0);
                llSay(0, "Register Agent: Error " +(string)status + ": " + body);
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
                + "&subtype=" + llEscapeURL(subtype)
                + "&class=" + llEscapeURL(class)
                + "&energy=" + llEscapeURL((string)npcEnergy)
                + "&money=" + llEscapeURL((string)npcMoney)
                + "&happiness=" + llEscapeURL((string)npcHappiness)
                + "&home=" + llEscapeURL(home)
                + "&amp;timestamp=" + myTimestamp
                + "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
            
            llSetTimerEvent(3600.0);    // if the registration fails, try later

        }
        else if (method == URL_REQUEST_DENIED)
        {
            llSetText("!!! BROKEN !!!", <1.0,0.0,0.0>, 1.0);
            llSay(0, "Register Agent: Something went wrong, no url. Error was: '" + body + "'");
        }
        else if (method == "POST" || method == "GET")
        {
            // incoming request for bot to do things
            llSay(0, "Register Agent: [Request from server:] " + body);
            
            list params = llParseStringKeepNulls(llUnescapeURL(body), ["&", "="], []);
            string response; // what we return
            key NPC = llGetOwner();
            if (osIsNpc(NPC))
                llSay(0, "Register Agent: Sanity check: This is an NPC with key " + (string)NPC);
            else
                llSay(0, "Register Agent: Sanity check failed: Key " + (string)NPC + " is NOT an NPC");
            
            // llOwnerSay("List parsed: " + (string) params);
            
            string command = llList2String(params, 1);
            
            if (llList2String(params, 0) == "command")
            {
                if (command == "osNpcGetRot")
                {
                    response = (string)osNpcGetRot(NPC);
                }
                else if (command == "osNpcSetRot")
                {
                    osNpcSetRot(NPC, llList2Rot(params, 3));
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
                    osNpcMoveToTarget(NPC, llList2Vector(params, 3), llList2Integer(params, 5));
                    response = "Moving to target " + llList2String(params, 3); 
                }
                else if (command == "osNpcStopMoveToTarget")
                {
                    osNpcStopMoveToTarget(NPC);
                    response = "Stopping.";
                }
                else if (command == "osNpcSit")
                {
                    osNpcSit(NPC, llList2Key(params, 3), llList2Integer(params, 5));
                    response = "Sitting on " +
                        llKey2Name(llList2Key(params, 3))
                        + " (" + llList2Key(params, 3) + ").";
                }
                else if (command == "osNpcStand")
                {
                    osNpcStand(NPC);
                    response = "Standing up.";
                }
                else if (command == "osNpcSay")
                {
                    osNpcSay(NPC, llList2Integer(params, 3), llUnescapeURL(llList2String(params, 5)));
                    response = "Saying \"" +
                        llUnescapeURL(llList2String(params, 5)) + "\".";
                }
                else if (command == "osNpcShout")
                {
                    osNpcShout(NPC, llList2Integer(params, 3), llUnescapeURL(llList2String(params, 5)));
                    response = "Shouting \"" +
                        llUnescapeURL(llList2String(params, 5)) + "\".";
                }
                else if (command == "osNpcWhisper")
                {
                    osNpcWhisper(NPC, llList2Integer(params, 3), llUnescapeURL(llList2String(params, 5)));
                    response = "Whispering \"" +
                        llUnescapeURL(llList2String(params, 5)) + "\".";
                }
                else if (command == "osNpcPlayAnimation")
                {
                    osNpcPlayAnimation(NPC, llList2String(params, 3));
                    response = "Playing animation \"" +
                        llList2String(params, 3) + "\".";
                }               
                else if (command == "osNpcStopAnimation")
                {
                    osNpcStopAnimation(NPC, llList2String(params, 3));
                    response = "Stopping animation \"" +
                        llList2String(params, 3) + "\".";
                }               
                else if (command == "osNpcLoadAppearance")
                {
                    osNpcLoadAppearance(NPC, llList2String(params, 3));
                    response = "Loading appearance \"" +
                        llList2String(params, 3) + "\".";
                }
                else if (command == "osNpcTouch")
                {
                    osNpcTouch(NPC, llList2Key(params, 3), llList2Integer(params, 5));
                    response = "Touching " + llKey2Name(llList2Key(params, 3))
                        + " (" + llList2Key(params, 3) + ").";
                }            
                else if (command == "getMoney")
                {
                    response = (string)npcMoney;
                }
                else if (command == "getHappiness")
                {
                    response = (string)npcHappiness;
                }
                else if (command == "getEnergy")
                {
                    response = (string)npcEnergy;
                }
                else if (command == "getHome")
                {
                    response = home;
                }
                else if (command == "getClass")
                {
                    response = class;
                }
                else if (command == "getSubType")
                {
                    response = subtype;
                }
                else if (command == "setMoney")
                {
                    npcMoney = llList2Float(params, 3);
                    response = "Setting Money to: " + (string)npcMoney;
                    setDescription();
                }
                else if (command == "setHappiness")
                {
                    npcHappiness = llList2Float(params, 3);
                    response = "Setting Happiness to: " + (string)npcHappiness;
                    setDescription();
                }
                else if (command == "setEnergy")
                {
                    npcEnergy = llList2Float(params, 3);
                    response = "Setting Energy to: " + (string)npcEnergy;
                    setDescription();                    
                }
                else if (command == "setHome")
                {
                    home = llList2String(params, 3);
                    response = "Setting Home to: " + home;
                    setDescription();
                }
                else if (command == "setClass")
                {
                    class = llList2String(params, 3);
                    response = "Setting Class to: " + class;
                    setDescription();
                }
                else if (command == "setSubType")
                {
                    subtype = llList2String(params, 3);
                    response = "Setting SubType to: " + subtype;
                    setDescription();
                }
                else
                {
                    response = "";
                    llHTTPResponse(id, 405, "Register Agent:  Unknown engine command " + llList2String(params, 0) + ".");
                }
            }
            
            if (response) 
            {
                llSay(0, "Register Agent: Sending back response to " + 
                    command + " '" +
                    response + "'...");
                llHTTPResponse(id, 200, response);
            }
            else
                llSay(0, "Register Agent: ERROR: No response or no command found!");
        }       
        else
        {
            llHTTPResponse(id, 405, "Method unsupported");
        }
    }
        
    changed(integer c)
    {
        // Region changed, get a new PermURL
        if (c & (CHANGED_REGION | CHANGED_REGION_START | CHANGED_TELEPORT | CHANGED_OWNER ) )
        {
            init();
        }
    }
    
    timer()
    {
        llSetText("Register Agent: Timed out, trying again to\nregister agent...", <1.0,0.0,0.0>, 1.0);
        llSetTimerEvent(0.0);
        init();
    }
}
</code></pre>
{{ end }}