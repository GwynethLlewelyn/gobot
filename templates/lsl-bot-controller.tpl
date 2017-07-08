{{ define "lsl-bot-controller" }}
<p>Copy the below code to a script called <code>bot controller.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
key npc;
integer listenChannel = 10;
list npcNames;
integer howManyNPCs = 0;
string deleteAvatarURL = "http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-agent/";
key deleteRequest;
string LSLSignaturePIN = "{{.LSLSignaturePIN}}";

default
{
    // NPC manipulator adapted by justincc 0.0.3 released 20121025
    state_entry()
    {
        llListen(listenChannel,"",NULL_KEY,"");
        llSetText("Listening on " + listenChannel, <0, 255, 0>, 1);
        llOwnerSay("Say /" + (string)listenChannel + " help for commands");
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
 
                    npc = osNpcCreate(FirstName, LastName, llGetPos() + <5, 5, 0>, notecardName, OS_NPC_SENSE_AS_AGENT);
                    
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
                    newNPC = osNpcCreate("John-" + (string)i, "Doe", pos + <xRand, yRand, 0>, notecardName);
                    npcNames += [newNPC];
                    llSleep(1);
                    npcCurrRot = osNpcGetRot(newNPC);
                    angle = llAtan2( llVecMag(pos % (pos + <xRand, yRand, 0>)), pos * (pos + <xRand, yRand, 0>) );
                    npcNewRot = npcCurrRot * llEuler2Rot(<0.0, 0.0, angle>);
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
                    //  database
                    string myTimestamp = llGetTimestamp();
                    deleteRequest = llHTTPRequest(deleteAvatarURL, [HTTP_METHOD, "POST", HTTP_MIMETYPE, "application/x-www-form-urlencoded"],
                        "request=delete"
                        + "&npc=" + (string)npc
                        + "&timestamp=" + myTimestamp
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
                        + "&timestamp=" + myTimestamp
                        + "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
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
                        + "&timestamp=" + myTimestamp
                        + "&signature=" + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
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
                             
                    vector delta = <(integer)msg1, (integer)msg2, 0>;
 
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
                    vector pos = <(integer)msg1, (integer)msg2, 0>;
 
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
                    osNpcMoveToTarget(npc, llGetPos() + <9,9,5>, OS_NPC_FLY|OS_NPC_LAND_AT_TARGET);
                }
            }
            else if (msg0 == "movetargetnoland" && npc != NULL_KEY)
            {
                integer npcToMove = llListFindList(npcNames, npc);
                
                if (npcToMove == -1) {
                    llOwnerSay("MoveTargetNoLand: NPC key " + (string)npc + " not found");
                } else {
                    osNpcMoveToTarget(npc, llGetPos() + <9,9,5>, OS_NPC_FLY);
                }
            }            
            else if (msg0 == "movetargetwalk" && npc != NULL_KEY)
            {
                integer npcToMove = llListFindList(npcNames, npc);
                
                if (npcToMove == -1) {
                    llOwnerSay("MoveTargetWalk: NPC key " + (string)npc + " not found");
                } else {
                    osNpcMoveToTarget(npc, llGetPos() + <9,9,0>, OS_NPC_NO_FLY);     
                }
            }
            else if (msg0 == "rot" && npc != NULL_KEY)
            {
                integer npcToMove = llListFindList(npcNames, npc);
                
                if (npcToMove == -1) {
                    llOwnerSay("Rot: NPC key " + (string)npc + " not found");
                } else {
                    vector xyz_angles = <0,0,90>; // This is to define a 1 degree change
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
                vector xyz_angles = <0, 0, (integer)msg1>;
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
        if (request_id == deleteRequest)
        {
            if (status == 200)
            {      
                llSay(0, "[Delete request:] " + body);
            }
            else
            {
                llSay(0, "Error " +(string)status + ": " + body);
            }
        }    
    }
}
</code></pre>
{{ end }}