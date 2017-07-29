{{ define "lsl-bot-controller" }}
<p>This is the master bot controller script, which allows controlling all agents, including creating new ones.</p>
<p>Copy the below code to a script called <code>bot controller.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
// Handles registration with the external database
// Send inventory and deal with external commands
// This combines the old listen channel with the new HTTP-based command, since we need to be able to clone avatars
//	to notecards (which was only present on the old code)

string registrationURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-position/";
string externalURL; // this is what we'll get from SL to get incoming connections
string webServerURLupdateInventory = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-inventory/";
key registrationRequest;	// used to track down the request for registration
key updateRequest;	  // used to track down the request for registration
key serverKey; // for inventory updates
key httpRequestKey;
list npcNames;
integer howManyNPCs = 0;
string deleteAvatarURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-position/";
key deleteRequest;
key npc; // used on chat commands
integer LSLSignaturePIN = {{.LSLSignaturePIN}};
integer listenChannel = 10;

init()
{
	llSetText("Registering bot controller...", &lt;1.0,0.0,0.0&gt;, 1.0);
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
		//llOwnerSay("On state_entry");
		llListen(listenChannel,"",NULL_KEY,"");
		llSetText("Listening on " + listenChannel, &lt;0, 255, 0&gt;, 1);
		llOwnerSay("Say /" + (string)listenChannel + " help for commands");
		llSetTimerEvent(0.0);
	}
	
	on_rez(integer what)
	{
		//llOwnerSay("On on_rez");
		init();
	}

	touch_start(integer total_number)
	{
		// llOwnerSay("On touch_start");
		// just re-register
		
		if (llDetectedKey(0) == llGetOwner() || llDetectedGroup(0))
		{
			if (llDetectedTouchFace(0) != 0)
				init();
		}
	}
	
	listen(integer channel, string name, key id, string msg)
	{
		if (msg != "")
		{
			list commands = llParseString2List(msg, [ " " ], []);
			string msg0 = llList2String(commands, 0);
			string msg1 = llList2String(commands, 1);			 
			string msg2 = llList2String(commands, 2);
			string msg3 = llList2String(commands, 3);
 
			if (msg0 == "create")
			{
				if (msg1 != "")
				{
					string notecardName = msg1;
					string FirstName = msg2;
					string LastName = msg3;
					
					if (FirstName == "") FirstName = "Jane";
					if (LastName == "") LastName = "Doe";
 
					npc = osNpcCreate(FirstName, LastName, llGetPos() + &lt;5, 5, 0&gt;, notecardName, OS_NPC_SENSE_AS_AGENT);
					
					npcNames += [npc];
 
					llOwnerSay("Created npc " + (string) npc + " from notecard " + notecardName);
				}
				else
				{
					llOwnerSay("Usage: create <notecard-name> <FirstName> <LastName>");
				}
			}  
			else if (msg0 =="createm")
			{
				// msg1 says number of NPCs to be created
				string notecardName = msg2;
				if (notecardName == "") notecardName = "appearance";
					
				// osOwnerSaveAppearance(notecardName);
				vector pos = llGetPos();
				integer i;
				npcNames = []; // reset list
				key newNPC;
				float angle;
				rotation npcCurrRot;
				rotation npcNewRot;
				
				for (i = 0; i < (integer)msg1; i++)
				{
					// add some randomness
					float xRand = llFrand(30) - 30;
					float yRand = llFrand(30) - 30;
					newNPC = osNpcCreate("John-" + (string)i, "Doe", pos + &lt;xRand, yRand, 0&gt;, notecardName);
					npcNames += [newNPC];
					llSleep(1);
					npcCurrRot = osNpcGetRot(newNPC);
					angle = llAtan2( llVecMag(pos % (pos + &lt;xRand, yRand, 0&gt;)), pos * (pos + &lt;xRand, yRand, 0&gt;) );
					npcNewRot = npcCurrRot * llEuler2Rot(&lt;0.0, 0.0, angle&gt;);
					osNpcSetRot(newNPC, npcNewRot);
					llOwnerSay("NPC <" + newNPC + "> created, old rot: " 
						+ (string)npcCurrRot + ", new rot: " + (string)npcNewRot);
				}
			}
			else if (msg0 == "remove" && npc != NULL_KEY)
			{
				integer npcToKill = llListFindList(npcNames, npc);
				
				if (npcToKill == -1) {
					llOwnerSay("Remove: NPC key '" + (string)npc + "' not found");
				} else {	   
					osNpcSay(npc, "You will pay for this with your liiiiiivvveeessss!!!.....");
					osNpcRemove(npc);
					
					// inform server to delete this bot from
					//	database
					string myTimestamp = llGetTimestamp();
					deleteRequest = llHTTPRequest(deleteAvatarURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"],
						"request=delete"
						+ "&npc=" + (string)npc
						+ "&amp;timestamp=" + myTimestamp
						+ "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
				}
			}
			else if (msg0 == "forceremove")
			{
				llInstantMessage(id, "Trying to remove " +
					msg1 + "(" + llKey2Name((key)msg1) + ")");
				osNpcRemove((key)msg1);
				string myTimestamp = llGetTimestamp();
				deleteRequest = llHTTPRequest(deleteAvatarURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"],
					  "request=delete"
						+ "&npc=" + msg1
						+ "&amp;timestamp=" + myTimestamp
						+ "&signature=" + llMD5String((string)llGetKey() + myTimestamp, 9876));
			}
			else if (msg0 == "removeall")
			{
				integer i;
				
				for (i = 0; i < llGetListLength(npcNames); i++)
				{
					osNpcRemove(llList2Key(npcNames, i));
					string myTimestamp = llGetTimestamp();
				deleteRequest = llHTTPRequest(deleteAvatarURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"],
					   "request=delete"
						+ "&npc=" + llList2String(npcNames, i)
						+ "&amp;timestamp=" + myTimestamp
						+ "&signature=" + llMD5String((string)llGetKey() + myTimestamp, 9876));
				}
				llOwnerSay("All NPCs removed");
			}
			else if (msg0 == "say" && npc != NULL_KEY)
			{
				integer npcToSay = llListFindList(npcNames, npc);
				
				if (npcToSay == -1) {
					llOwnerSay("Say: NPC key " + (string)npc + "not found");
				} else {
					osNpcSay(llList2Key(npcNames, npcToSay), "I am your worst Nightmare!!!!");
				}
			}	
			else if (msg0 == "move")
			{
				integer npcToMove = llListFindList(npcNames, npc);
				if (msg1 != "" && msg2 != "" && npc != NULL_KEY && npcToMove != -1)
				{	  
					key npcMoving = llList2Key(npcNames, npcToMove);
							 
					vector delta = &lt;(integer)msg1, (integer)msg2, 0&gt;;
 
					if (msg3 != "")
					{
						delta.z = (integer)msg3;
					}
 
					osNpcMoveTo(npcMoving, osNpcGetPos(npc) + delta);					 
				}							 
				else
				{
					llOwnerSay("Usage: move <x> <y> [<z>]");
				}
			}	
			else if (msg0 == "moveto")
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (msg1 != "" && msg2 != "" && npc != NULL_KEY && npcToMove != -1)
				{				 
					vector pos = &lt;(integer)msg1, (integer)msg2, 0&gt;;
 
					if (msg3 != "")
					{
						pos.z = (integer)msg3;
					}
 
					osNpcMoveTo(npc, pos);					  
				}							 
				else
				{
					llOwnerSay("Usage: move <x> <y> [<z>]");
				}
			}			 
			else if (msg0 == "movetarget" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1)
				{
					llOwnerSay("MoveTarget: NPC key " + (string)npc + " not found");
				} else {
					osNpcMoveToTarget(npc, llGetPos() + &lt;9,9,5&gt;, OS_NPC_FLY|OS_NPC_LAND_AT_TARGET);
				}
			}
			else if (msg0 == "movetargetnoland" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("MoveTargetNoLand: NPC key " + (string)npc + " not found");
				} else {
					osNpcMoveToTarget(npc, llGetPos() + &lt;9,9,5&gt;, OS_NPC_FLY);
				}
			}			 
			else if (msg0 == "movetargetwalk" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("MoveTargetWalk: NPC key " + (string)npc + " not found");
				} else {
					osNpcMoveToTarget(npc, llGetPos() + &lt;9,9,0&gt;, OS_NPC_NO_FLY);	   
				}
			}
			else if (msg0 == "rot" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Rot: NPC key " + (string)npc + " not found");
				} else {
					vector xyz_angles = &lt;0,0,90&gt;; // This is to define a 1 degree change
					vector angles_in_radians = xyz_angles * DEG_TO_RAD; // Change to Radians
					rotation rot_xyzq = llEuler2Rot(angles_in_radians); // Change to a Rotation				   
					rotation rot = osNpcGetRot(npc);
					osNpcSetRot(npc, rot * rot_xyzq);
				}
			}
			else if (msg0 == "rotabs" && msg1 != "")
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Rotabs: NPC key " + (string)npc + " not found");
				} else {
				vector xyz_angles = &lt;0, 0, (integer)msg1&gt;;
				vector angles_in_radians = xyz_angles * DEG_TO_RAD; // Change to Radians
				rotation rot_xyzq = llEuler2Rot(angles_in_radians); // Change to a Rotation				   
				osNpcSetRot(npc, rot_xyzq); 
				}				
			}
			else if (msg0 == "animate" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Animate: NPC key " + (string)npc + " not found");
				} else {
					osAvatarPlayAnimation(npc, "stabbed+die_2");
					llSleep(3);
					osAvatarStopAnimation(npc, "stabbed+die_2");
				}
			}	
			else if (msg0 == "getrot" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Get rotation: NPC key " + (string)npc + " not found");
				} else {
					llSay(0, "Rotation is: " + (string)osNpcGetRot(npc));
				}
			}	
			else if (msg0 == "save" && msg1 != "" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Save: NPC key " + (string)npc + " not found");
				} else {
					osNpcSaveAppearance(npc, msg1);
					llOwnerSay("Saved appearance " + msg1 + " to " + npc);
				}		
			}
			else if (msg0 == "load" && msg1 != "" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Load appearance: NPC key " + (string)npc + " not found");
				} else {
					osNpcLoadAppearance(npc, msg1);
					llOwnerSay("Loaded appearance " + msg1 + " to " + npc);
				}
			}
			else if (msg0 == "clone")
			{
				if (msg1 != "")
				{
					osOwnerSaveAppearance(msg1);
					llOwnerSay("Cloned your appearance to " + msg1);
				}
				else
				{
					llOwnerSay("Usage: clone <notecard-name-to-save>");
				}
			}
			else if (msg0 == "stop" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Stop: NPC key " + (string)npc + " not found");
				} else {
					osNpcStopMoveToTarget(npc);
				}
			}
			else if (msg0 == "sit" && msg1 != "" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Sit: NPC key " + (string)npc + " not found");
				} else {
					osNpcSit(npc, msg1, OS_NPC_SIT_NOW);
				}
			}	
			else if (msg0 == "stand" && npc != NULL_KEY)
			{
				integer npcToMove = llListFindList(npcNames, npc);
				
				if (npcToMove == -1) {
					llOwnerSay("Stand: NPC key " + (string)npc + " not found");
				} else {
					osNpcStand(npc);
				}
			}
			else if (msg0 == "swarm")
			{
				// is list empty?
				
				if (npcNames == []) llOwnerSay("Swarm: no NPCs");
				
				// go through the list
				integer i;
				vector currPos = llGetPos();
				vector npcCurrPos;
				vector npcFuturePos;
				key currNPC;
				
				for (i = 0; i < llGetListLength(npcNames); i++)
				{
					currNPC = llList2Key(npcNames, i);
					npcCurrPos = osNpcGetPos(currNPC);
					
					
					// calculate intermediate point (quarter the distance)
					npcFuturePos = ((currPos - npcCurrPos) / 4.0) + npcCurrPos;
					
					osNpcMoveToTarget(currNPC, npcFuturePos, OS_NPC_NO_FLY); 
					osNpcSay(currNPC, "Moving from " + (string)currPos + " to " 
						+ (string)npcFuturePos);
				}
				llOwnerSay("One moving interaction finished");
			}
			else if (msg0 == "help")
			{
				llOwnerSay("Commands are:");
				llOwnerSay("create <notecard-name> <FirstName> <LastName> - Create NPC from a stored notecard");
				llOwnerSay("createm <N> <notecard-name> - Create N NPCs from a notecard");		 
				llOwnerSay("remove - Remove current NPC");
				llOwnerSay("forceremove <uuid> - Remove NPC with key <uuid>");
				llOwnerSay("removeall - Remove all NPCs");	
				llOwnerSay("clone <notecard-name> - Clone own appearance to a notecard");
				llOwnerSay("load <notecard-name>  - Load appearance on notecard to current npc");
				llOwnerSay("save <notecard-name>  - Save appearance of current NPC to notecard");
				llOwnerSay("animate");
				llOwnerSay("move");
				llOwnerSay("moveto <x> <y> <z> - move to absolute position");
				llOwnerSay("movetarget");
				llOwnerSay("movetargetnoland");
				llOwnerSay("movetargetwalk");
				llOwnerSay("rot");
				llOwnerSay("getrot");
				llOwnerSay("say");
				llOwnerSay("sit <target-uuid>");
				llOwnerSay("stop");
				llOwnerSay("stand");
				llOwnerSay("swarm");
			}
			else
			{
				llOwnerSay("I don't understand [" + msg + "]");
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
			}
			else
			{
				llSetText("!!! BROKEN !!!", &lt;1.0,0.0,0.0&gt;, 1.0);
				llOwnerSay("Error " +(string)status + ": " + body);
			}
		}
	}
	http_request(key id, string method, string body)
	{
		//llOwnerSay("Entering http_request for registration...");
		if (method == URL_REQUEST_GRANTED)
		{
			externalURL = body;
			string myTimestamp = llGetTimestamp();
			string formRequest = "permURL=" + llEscapeURL(externalURL)
				+ "&objecttype=" + llEscapeURL("Bot Controller")
				+ "&amp;timestamp=" + myTimestamp
				+ "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN);
			// llOwnerSay("Registration URL is " + registrationURL + " Form request is: " + formRequest);
			
			registrationRequest = llHTTPRequest(registrationURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"], 
				formRequest);
			
			llSetTimerEvent(3600.0);	// if the registration fails, try later

		}
		else if (method == URL_REQUEST_DENIED)
		{
			llSetText("!!! BROKEN !!!", &lt;1.0,0.0,0.0&gt;, 1.0);
			llOwnerSay("Something went wrong, no url. " + body);
		}
		else if (method == "POST" || method == "GET")
		{
			// incoming request for bot to do things
			//llSay(0, "[Request from server:] " + body);
			
			list params = llParseStringKeepNulls(llUnescapeURL(body), ["&", "="], []);
			string response; // what we return
			
			// first parameter will always be be npc=
			
			key NPC = llList2String(params, 1);
			//if (osIsNpc(NPC))
			//	  llSay(0, "Sanity check: This is an NPC with key " + (string)NPC);
			//else
			//	  llSay(0, "Sanity check failed: Key " + (string)NPC + " is NOT an NPC");
			
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
					response = "Moving to target " + llList2String(params, 5); 
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
					rotation npcNewRot = npcCurrRot * llEuler2Rot(&lt;0.0, 0.0, angle&gt;);
					osNpcSetRot(newNPC, npcNewRot);					
					osNpcSetProfileImage(newNPC, "botimage"); // this texture MUST be inside the content for this to work!
					osNpcSetProfileAbout(newNPC, "Hello! I'm just a friendly bot passing by! Please ignore me!");

					response = "New NPC: " + (string)newNPC;
				}
				else if (command == "osNpcRemove")
				{
					osNpcRemove(NPC);
					response = "Removing " + llKey2Name(NPC);
				}
				else if (command == "ping")
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
		llSetText("Timed out, trying again to\nregister bot controller...", &lt;1.0,0.0,0.0&gt;, 1.0);
		llOwnerSay("Timed out, trying again to register bot controller...");
		llSetTimerEvent(0.0);
		init();
	}
}

state read_inventory
{
	state_entry()
	{
		llSetText("Sending to webserver - 0%", &lt;0.3, 0.7, 0.2&gt;, 1.0);
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
		llSetText("Checking inventory...", &lt;1.0,1.0,0.0&gt;, 1.0);
		
		for (i = 0; i < length; i++)
		{
			itemName = llGetInventoryName(INVENTORY_ALL, i);
			
			if (llGetInventoryType(itemName) != INVENTORY_SCRIPT && llGetInventoryType(itemName) != INVENTORY_TEXTURE) // skip script, skip textures
			{
				myTimeStamp = llGetTimestamp();
				
				httpBody =	"name=" + llEscapeURL(itemName) + 
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
			
				llSetText("Sending to webserver - " + (string) ((integer)((float)i/(float)length*100)) + "%", &lt;0.3, 0.7, 0.2&gt;, 1.0);
			}
		}
		state default;
	}
	
	http_response(key request_id, integer status, list metadata, string body)
	{ 
		llSetText("", &lt;0.0,0.0,0.0&gt;, 1.0);
		
		if (request_id == httpRequestKey)
		{
			if (status != 200)
			{
				llOwnerSay("HTTP Error " + (string)status + ": " + body);
			}
			else 
			{
				//llOwnerSay("Web-server reply: " + body); 
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
		//llOwnerSay("state_exit inventory");
		llSetTimerEvent(0.0);
	}
}
</code></pre>
{{ end }}