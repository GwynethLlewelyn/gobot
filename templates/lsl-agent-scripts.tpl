{{ define "lsl-agent-scripts" }}
<p>These three scripts go inside a transparent box, worn by the agent's avatar.</p>
<p>Copy the first code to a script called <code>register agent.lsl</code> and put it inside the transparent box. When the agent's avatar is rezzed, it will use this script to register with the engine, and be able to receive commands from it.</p>

<pre><code class="language-javascript">// Handles agent registration with the external database
// Send current configuration for agent class, subtype, and energy/money/happiness
// On the first version it will only work in OpenSimulator and call osNPC functions

string registrationURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-agent/";
string externalURL; // this is what we'll get from SL to get incoming connections
key registrationRequest;	// used to track down the request for registration
key httpRequestKey;
string class = "peasant";
string subtype = "publican";
string home = ""; // place where NPC will return when energy exhausted
float npcEnergy = 1.0;	  // start with max energy and happiness
float npcMoney = 0.0; // but no money
float npcHappiness = 1.0;
integer LSLSignaturePIN = {{.LSLSignaturePIN}};

init()
{
	llSetObjectName("Bot Controller - " + llKey2Name(llGetOwner())); 
	llSetText("Registering agent...", &lt;1.0,0.0,0.0&gt;, 1.0);
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
	
	class =			llList2String(params, 0);
	subtype =		llList2String(params, 1);
	npcEnergy =		llList2Float(params, 2);
	npcMoney =		llList2Float(params, 3);
	npcHappiness =	llList2Float(params, 4);
	home =			llList2String(params, 5);
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
			"\nHome: " + home, &lt;npcEnergy,npcMoney,npcHappiness&gt;, 1.0);
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
				llSetText("!!! BROKEN !!!", &lt;1.0,0.0,0.0&gt;, 1.0);
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
			
			llSetTimerEvent(3600.0);	// if the registration fails, try later

		}
		else if (method == URL_REQUEST_DENIED)
		{
			llSetText("!!! BROKEN !!!", &lt;1.0,0.0,0.0&gt;, 1.0);
			llSay(0, "Register Agent: Something went wrong, no url. Error was: '" + body + "'");
		}
		else if (method == "POST" || method == "GET")
		{
			// incoming request for bot to do things
			//llSay(0, "Register Agent: [Request from server:] " + body);
			
			list params = llParseStringKeepNulls(llUnescapeURL(body), ["&", "="], []);
			string response; // what we return
			key NPC = llGetOwner();
			//if (osIsNpc(NPC))
			//	  llSay(0, "Register Agent: Sanity check: This is an NPC with key " + (string)NPC);
			//else
			//	  llSay(0, "Register Agent: Sanity check failed: Key " + (string)NPC + " is NOT an NPC");
			
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
				else if (command == "ping")
				{
					response = "pong";
				}
				else
				{
					response = "";
					llHTTPResponse(id, 405, "Register Agent:  Unknown engine command " + command + ".");
				}
			}
			
			if (response) 
			{
				//llSay(0, "Register Agent: Sending back response to " + 
				//	  command + " '" +
				//	  response + "'...");
				llHTTPResponse(id, 200, response);
			}
			else
			{
				llSay(0, "Register Agent: ERROR: No response or no command found!'" + command + "' found.");
				llHTTPResponse(id, 404, "No response or no command '" + command + "' found.");
			}	
		}		
		else
		{
			llHTTPResponse(id, 405, "Method '" + method + "' unsupported.");
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
		llSetText("Register Agent: Timed out, trying again to\nregister agent...", &lt;1.0,0.0,0.0&gt;, 1.0);
		llSetTimerEvent(0.0);
		init();
	}
}
</code></pre>

<p>This next script will use llCastRay to try to figure out objects around the agent at a narrow scope. llCastRay will be a bit more precise to tell the avatar if it's an object really in front of them.</p>
<p>Copy it, name it <code>llCastRay detector script 3.1.lsl</code> and drop it inside the transparent box as with the script before.</p>
	
<pre><code class="language-javascript">
// Version 3.1 (for GoBot)
// Just based on object rotation
// Thanks to Lucinda Bulloch for a *working* llCastRay() demo script!
// includes sending it to external webserver
// to-do: see where the missing info comes from!

string sensorURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-sensor/";
key sensorRequest;

particles(key uuid)
{
	// laser beam
	llParticleSystem([
			PSYS_SRC_PATTERN, PSYS_SRC_PATTERN_ANGLE, 
			PSYS_SRC_BURST_PART_COUNT,(integer) 4,	 // adjust for beam strength,
			PSYS_SRC_BURST_RATE,(float) .05,		  
			PSYS_PART_MAX_AGE,(float)  1.2,	 // was .6			
			PSYS_SRC_BURST_SPEED_MIN,(float)1,		  
			PSYS_SRC_BURST_SPEED_MAX,(float) 7.0,	   
			PSYS_PART_START_SCALE,(vector) &lt;0,.1,0&gt;, 
			PSYS_PART_END_SCALE,(vector) &lt;.04,.5,0&gt;,	 
			PSYS_PART_START_COLOR,(vector) &lt;1,0,0&gt;,  
			PSYS_PART_END_COLOR,(vector) &lt;.2,0,0&gt;,   
			PSYS_PART_START_ALPHA,(float)0.5,		   
			PSYS_PART_END_ALPHA,(float)0.00,
			PSYS_SRC_TARGET_KEY, uuid, 
			PSYS_PART_FLAGS,
			PSYS_PART_EMISSIVE_MASK |	  
			PSYS_PART_FOLLOW_VELOCITY_MASK |
			PSYS_PART_FOLLOW_SRC_MASK |	  
			PSYS_PART_INTERP_SCALE_MASK |
			PSYS_PART_TARGET_LINEAR_MASK ]
		);
}

detection()
{
	vector start = llGetPos();
//	vector end = start + &lt;10.0, 0.0, 0.0&gt;*llGetCameraRot();
	vector end = start + &lt;10.0, 0.0, 0.0&gt;*llGetRot();
	
	// rotation camRot = llGetCameraRot();
	// vector camPos = llGetCameraPos();
	
	// to-do: point slightly downwards
	
		/*	  
	rotation objectRot = llGetRootRotation() * llEuler2Rot(&lt;0.0, -45.0 * DEG_TO_RAD, 0.0&gt;);	
	rotation castRot = llGetRootRotation() * llEuler2Rot(&lt;0.0, 45.0 * DEG_TO_RAD, 0.0&gt;);	 

	llSetRot(objectRot);			

	llOwnerSay("Debug: Object rot: " + (string) objectRot
		+ " Root rot: " + (string) llGetRootRotation());
	// + " Camera rot: "
	//	  + (string) camRot + " Camera pos: " + (string) camPos);
		*/
		
	// llSetLinkPrimitiveParamsFast(2, [PRIM_ROT_LOCAL, objectRot]); // set prim to point slightly downwards
	 list results = llCastRay(start, end, [
		RC_DATA_FLAGS, RC_GET_ROOT_KEY,
		RC_REJECT_TYPES, RC_REJECT_LAND,
		RC_MAX_HITS, 50,
		RC_DETECT_PHANTOM, TRUE
	]);
 
	integer statusCode = llList2Integer(results, -1);
	
	if (statusCode == RCERR_SIM_PERF_LOW)
	{
		llSay(0, "Sim performance low, cannot use llRayCast()");
	} 
	else if (statusCode == RCERR_CAST_TIME_EXCEEDED)
	{
		llSay(0, "Too many raycasts");
	}
	else if (statusCode == RCERR_UNKNOWN)
	{
	 	llSay(0, "Unknown raycast error");	
	}
	else
	{
					
		//llOwnerSay("Current rotation: " + (string)llGetRot());
		//llOwnerSay("Current camera rotation: " + (string)llGetCameraRot());
		//llOwnerSay("Detected: " + llDumpList2String(results, "|"));
	
		integer hitNum;
		
		for (hitNum = 0; hitNum < statusCode; hitNum++)
		{
			key uuid = llList2Key(results, 2*hitNum);
			vector pos = llList2Vector(results, 2*hitNum+1);
			float dist = llVecDist(start, pos);
			string name;
			
			if (uuid == NULL_KEY)
			{
				// llOwnerSay("Land at " + (string)dist + "m");
			}
			else if (uuid == llGetOwner())
			{
				// llOwnerSay("Self-detected; ignoring");
			}
			else
			{
				name = llKey2Name(uuid);
				//llOwnerSay(name + "<" + (string) uuid + "> at " + (string)dist + "m");
				particles(uuid);
				
				// get a few more details				
				list objectDetailsList = llGetObjectDetails(uuid, [ OBJECT_ROT, OBJECT_VELOCITY, OBJECT_CREATOR, OBJECT_PHANTOM, OBJECT_PRIM_EQUIVALENCE]);
				string type;
				
				// test if it's an avatar or an object
				if (llList2Key(objectDetailsList, 2) == NULL_KEY) // avatars have no creator!
				{
					type = "1";
				}
				else
				{
					type = "12"; // not entirely correct but ok; more checks are needed to figure out the exact type
				}
				
				// now get the bounding box too
				list boundingBox = llGetBoundingBox(uuid);
				
				// Sending remotely
				sensorRequest = llHTTPRequest(sensorURL +
					"?key=" + (string)uuid +
					"&name=" + name +
					"&pos=" + (string)pos +
					"&rot=" + llList2String(objectDetailsList, 0) +
					"&vel=" + llList2String(objectDetailsList, 1) +
					"&phantom=" + llList2String(objectDetailsList, 3) +
					"&prims=" + llList2String(objectDetailsList, 4) +
					"&bblo=" + llList2String(boundingBox, 0) +
					"&bbhi=" + llList2String(boundingBox, 1) +
					"&type=" + type +
					"&origin=castray" +
					"&amp;timestamp=" + llGetTimestamp(),
					[HTTP_METHOD, "GET"], "");
			}
		}
	}
}

default
{
	state_entry()
	{
//		llRequestPermissions(llGetOwner(), PERMISSION_CONTROL_CAMERA|PERMISSION_TRACK_CAMERA);
		llClearCameraParams();
		llSetTimerEvent(5.0);
	}
	
	state_exit()
	{
		llParticleSystem([]);
	}
	
	on_rez(integer start_param)
	{
		llResetScript();
	}

	attach(key id)
	{
		if (id == llGetOwner())
		{
//			  llRequestPermissions(id, PERMISSION_TAKE_CONTROLS|PERMISSION_TRACK_CAMERA);
			llResetScript();
		}
		else
		{
			llParticleSystem([]);
//			  llReleaseControls();
		}
	}	
	
	run_time_permissions(integer perm)
	{
		if (perm & PERMISSION_CONTROL_CAMERA)
		{
			llClearCameraParams();
			llSetTimerEvent(5.0);

			/* llSetCameraParams([
				CAMERA_ACTIVE, 1, // 1 is active, 0 is inactive
				CAMERA_BEHINDNESS_ANGLE, 0.0, // (0 to 180) degrees was 45
				CAMERA_BEHINDNESS_LAG, 0.0, // (0 to 3) seconds was 0.5
				CAMERA_DISTANCE, 0.5, // ( 0.5 to 10) meters was 8
				//CAMERA_FOCUS, &lt;0,0,5&gt;, // region relative position
				CAMERA_FOCUS_LAG, 0.0 , // (0 to 3) seconds was 0.05
				CAMERA_FOCUS_LOCKED, FALSE, // (TRUE or FALSE)
				CAMERA_FOCUS_THRESHOLD, 0.0, // (0 to 4) meters
				CAMERA_PITCH, 0.0, // (-45 to 80) degrees was 20
				//CAMERA_POSITION, &lt;0,0,0&gt;, // region relative position
				CAMERA_POSITION_LAG, 0.1, // (0 to 3) seconds
				CAMERA_POSITION_LOCKED, FALSE, // (TRUE or FALSE)
				CAMERA_POSITION_THRESHOLD, 0.0, // (0 to 4) meters
				CAMERA_FOCUS_OFFSET, &lt;1.0,0.0,1.0&gt; // &lt;-10,-10,-10&gt; to &lt;10,10,10&gt; meters was 3,0,2
				]);
			*/
		}
	}
	
	touch(integer who)
	{
		if (llDetectedKey(0) == llGetOwner())
		{
			llParticleSystem([]);
			detection();
		}
	}
	
	timer()
	{
		llParticleSystem([]);
		//llOwnerSay("Timer" + llGetTimestamp());
		detection();
	}
	
	http_response(key request_id, integer status, list metadata, string body)
	{
		if (request_id == sensorRequest)
		{
			if (status == 200)
			{			
				// llOwnerSay(body);
			}
			else
			{
				llSay(0, "HTTP Error " + (string)status + ": " + body);
			}
		}
	}
}
</code></pre>

<p>The last script uses regular sensors to gather even more data around itself. Sensors have a wider range than casting rays, but they basically retrieve objects irrespectively of their size, so we don't know if they are true obstacles or not.</p>
<p>Copy the code below to a script called <code>Sensorama.lsl</code> and place it inside the transparent box as before.</p>
<pre><code class="language-javascript">
// Inside bot attachment
// Senses data and sends it to remote host for processing

string sensorURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-sensor/";
key sensorRequest;

default
{
	state_entry()
	{
		llSay(0, "Attempting to sensorize...");
		llSensorRepeat("", NULL_KEY, AGENT|NPC|ACTIVE|PASSIVE|SCRIPTED, 10.0, PI/2, 5.0);
	}
	
	attach(key where)
	{
		if (where != NULL_KEY)
		{
			llResetScript();
		}
	}
	
	sensor(integer numDetected)
	{
		integer i;
		
		for (i = 0; i < numDetected; i++)
		{
			
			/*llOwnerSay("Detected " +
				llDetectedName(i) + " at " +
				(string)llDetectedPos(i) + " type: " +
				(string)llDetectedType(i));*/
			
			 // get a few more details
				
			list objectDetailsList = llGetObjectDetails(llDetectedKey(i), [OBJECT_PHANTOM, OBJECT_PRIM_EQUIVALENCE]);
			// now get the bounding box too
			list boundingBox = llGetBoundingBox(llDetectedKey(i));
			
			// Sending remotely
			sensorRequest = llHTTPRequest(sensorURL +
				"?key=" + llDetectedKey(i) +
				"&name=" + llEscapeURL(llDetectedName(i)) +
				"&pos=" + (string)llDetectedPos(i) +
				"&rot=" + (string)llDetectedRot(i) +
				"&vel=" + (string)llDetectedVel(i) +
				"&phantom=" + llList2String(objectDetailsList, 0) +
				"&prims=" + llList2String(objectDetailsList, 1) +
				"&bblo=" + llList2String(boundingBox, 0) +
				"&bbhi=" + llList2String(boundingBox, 1) +
				"&origin=sensor" +
				"&type=" + llDetectedType(i) +
				"&amp;timestamp=" + llGetTimestamp(),
				[HTTP_METHOD, "GET"], "");
		}
	}
	
	 http_response(key request_id, integer status, list metadata, string body)
	{
		if (request_id == sensorRequest)
		{
			if (status == 200)
			{			
				// llOwnerSay(body);
			}
			else
			{
				llSay(0, "HTTP Error " + (string)status + ": " + body);
			}
		}
	}
}
</code></pre>
{{ end }}