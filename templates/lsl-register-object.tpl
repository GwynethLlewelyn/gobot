{{ define "lsl-register-object" }}
<p>Copy the below code to a script called <code>GoBot Registering Button.lsl</code> and put it inside a cube.</p>

<pre><code class="language-javascript">
// Handles registration with the external database
// Send current configuration for class, type, and rates for energy/money/happiness
// To-do: Accepts remote calls to trigger animations, send IMs to people etc

string registrationURL = &quot;http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/register-position/&quot;;
string externalURL; // this is what we'll get from SL to get incoming connections
string webServerURLupdateInventory = &quot;http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/update-inventory/&quot;;
string processCubeURL = &quot;http://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/process-cube/&quot;;
key processCubeRequest;
key registrationRequest;    // used to track down the request for registration
key updateRequest;    // used to track down the request for registration
key serverKey; // for inventory updates
key httpRequestKey;
integer LSLSignaturePIN = 9876;
string type = &quot;money&quot;;
string class = &quot;peasant&quot;;
float rateEnergy;   // cube configuration parameters; touch on the face
float rateMoney;
float rateHappiness;
vector touchST; // for the sliders

init()
{
    llSetText(&quot;Registering position...&quot;, &lt;1.0,0.0,0.0&gt;, 1.0);
    llOwnerSay(&quot;Registering position...&quot;);
    // parse description field
    parseDescription();

    // release URLs before requesting a new one
    llReleaseURL(externalURL);
    externalURL = &quot;&quot;;
    llRequestURL();
}

// parse description field, which contains the type of box and the class it applies to
parseDescription()
{
    list params = llParseString2List(llGetObjectDesc(), [&quot;;&quot;], []);

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
    llSetText(&quot;Energy: &quot; + (string)rateEnergy +
            &quot;\nMoney: &quot; + (string)rateMoney +
            &quot;\nHappiness: &quot; + (string)rateHappiness, &lt;1.0,1.0,1.0&gt;, 1.0);
}

default
{
    state_entry()
    {
        parseDescription();
        llSetTimerEvent(3600.0); // this will hopefully force an update every hour
    }

    on_rez(integer what)
    {
        init();
    }

    touch_start(integer total_number)
    {
        // this has to be redone: touching will allow to set the cube&acute;s class, type, etc.

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
                llWhisper(PUBLIC_CHANNEL, &quot;Sorry, your viewer does not support touched faces.&quot;);
            else if (llDetectedTouchFace(0) == 0)
            {
                touchST = llDetectedTouchST(0);

                if (touchST != TOUCH_INVALID_TEXCOORD)
                {
                    // happiness: 0 &lt;= 0.33
                    // money: &gt; 0.33 &lt;= 0.67
                    // energy: &gt; 0.67 &lt;= 1.0

                    if (touchST.y &lt;= 0.33)
                    {
                        rateHappiness = (2 * touchST.x) - 1.0;
                    }
                    else if (touchST.y &gt; 0.33 &amp;&amp; touchST.y &lt;= 0.67)
                    {
                        rateMoney = (2 * touchST.x) - 1.0;
                    }
                    else if (touchST.y &gt; 0.67)
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
        if ((llDetectedKey(0) == llGetOwner() || llDetectedGroup(0)) &amp;&amp; llDetectedTouchFace(0) == 0)
        {
         // save to description
            llSetObjectDesc(type + &quot;;&quot; + class + &quot;;&quot; + (string)rateEnergy + &quot;;&quot; + (string)rateMoney
                + &quot;;&quot; + (string)rateHappiness);

            // if permURL is empty, do a full registration
            if (externalURL == &quot;&quot;)
            {
                init();
            }
            else
            {
                llInstantMessage(llDetectedKey(0), &quot;Updating information...&quot;);
                string myTimestamp = llGetTimestamp();
                updateRequest = llHTTPRequest(registrationURL, [HTTP_METHOD, &quot;POST&quot;, HTTP_MIMETYPE, &quot;application/x-www-form-urlencoded&quot;],
                    &quot;permURL=&quot; + llEscapeURL(externalURL)
                    + &quot;&amp;objecttype=&quot; + llEscapeURL(type)
                    + &quot;&amp;objectclass=&quot; + llEscapeURL(class)
                    + &quot;&amp;rateenergy=&quot; + llEscapeURL((string)rateEnergy)
                    + &quot;&amp;ratemoney=&quot; + llEscapeURL((string)rateMoney)
                    + &quot;&amp;ratehappiness=&quot; + llEscapeURL((string)rateHappiness)
                    + &quot;&amp;timestamp=&quot; + myTimestamp
                    + &quot;&amp;signature=&quot; + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
            }
        }
    }

    http_response(key request_id, integer status, list metadata, string body)
    {
        if (request_id == registrationRequest || request_id == updateRequest || request_id == processCubeRequest)
        {
            if (status == 200)
            {
                llOwnerSay(body);
                // new registration? switch to inventory reading
                if (request_id == registrationRequest)
                    state read_inventory;
                // if it is just an update, no need to do anything else for now
                updateSetText();
            }
            else
            {
                llSetText(&quot;!!! BROKEN !!!&quot;, &lt;1.0,0.0,0.0&gt;, 1.0);
                llOwnerSay(&quot;Error &quot; +(string)status + &quot;: &quot; + body);
                llSetTimerEvent(3600.0);
            }
        }
    }
    http_request(key id, string method, string body)
    {
        if (method == URL_REQUEST_GRANTED)
        {
            externalURL = body;

            string myTimestamp = llGetTimestamp();
            registrationRequest = llHTTPRequest(registrationURL, [HTTP_METHOD, &quot;POST&quot;, HTTP_MIMETYPE, &quot;application/x-www-form-urlencoded&quot;],
                &quot;permURL=&quot; + llEscapeURL(externalURL)
                + &quot;&amp;objecttype=&quot; + llEscapeURL(type)
                + &quot;&amp;objectclass=&quot; + llEscapeURL(class)
                + &quot;&amp;rateenergy=&quot; + llEscapeURL((string)rateEnergy)
                + &quot;&amp;ratemoney=&quot; + llEscapeURL((string)rateMoney)
                + &quot;&amp;ratehappiness=&quot; + llEscapeURL((string)rateHappiness)
                + &quot;&amp;amp;timestamp=&quot; + myTimestamp
                + &quot;&amp;signature=&quot; + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));

            llSetTimerEvent(3600.0);    // if the registration fails, try later

        }
        else if (method == URL_REQUEST_DENIED)
        {
            llSetText(&quot;!!! BROKEN !!!&quot;, &lt;1.0,0.0,0.0&gt;, 1.0);
            llOwnerSay(&quot;Something went wrong, no url. &quot; + body);
            llSetTimerEvent(3600.0);
        }
        else if (method == &quot;POST&quot; || method == &quot;GET&quot;)
        {
            // NOTE(gwyneth): This will allow the web-based interface to send commands to each and every object.
            //  Right now, it is just used to see if the objects are alive; this is needed by the Garbage Collector,
            //  it sends a 'command=ping' and expects a 'pong'; if not, it assumes that the object is 'dead' and cleans up. (20170729)

            list params = llParseStringKeepNulls(llUnescapeURL(body), [&quot;&amp;&quot;, &quot;=&quot;], []);
            string response; // what we return

            string commandTag = llList2String(params, 0);
            string command = llList2String(params, 1);

            if (commandTag == &quot;command&quot;)
            {
                if (command == &quot;ping&quot;)
                {
                     response = &quot;pong&quot;;
                }
                else
                {
                    response = &quot;&quot;;
                    llHTTPResponse(id, 405, &quot;Unknown engine command &quot; + command + &quot;.&quot;);
                }
            }

            if (response)
            {
                //llSay(0, &quot;Sending back response for &quot; +
                //      command + &quot; '&quot; +
                //      response + &quot;'...&quot;);
                llHTTPResponse(id, 200, response);
            }
            else
                llSay(0, &quot;ERROR: No response or no command found!&quot;);
        }
        else
        {
            llHTTPResponse(id, 405, &quot;Method unsupported&quot;);
        }
    }

    changed(integer c)
    {
        // Region changed, get a new PermURL
        if (c &amp; (CHANGED_REGION | CHANGED_REGION_START | CHANGED_TELEPORT ) )
        {
            init();
        }
        // Deal with inventory changes
        else if (c &amp; CHANGED_INVENTORY)
        {
            state read_inventory;
        }
        else if (c &amp; CHANGED_LINK)
        {
            key av = llAvatarOnSitTarget();
            if (av) // evaluated as true if key is valid and not NULL_KEY
            {
                string myTimestamp = llGetTimestamp();
                processCubeRequest = llHTTPRequest(processCubeURL, [HTTP_METHOD, &quot;POST&quot;, HTTP_MIMETYPE, &quot;application/x-www-form-urlencoded&quot;],
                  &quot;avatar=&quot; + (string)av
                + &quot;&amp;amp;timestamp=&quot; + myTimestamp
                + &quot;&amp;signature=&quot; + llMD5String((string)llGetKey() + myTimestamp, LSLSignaturePIN));
            }
        }
    }

    timer()
    {
        llSetText(&quot;Trying again to register cube/position...&quot;, &lt;1.0,0.0,0.0&gt;, 1.0);
        llSetTimerEvent(0.0);
        init();
    }
}

state read_inventory
{
    state_entry()
    {
        llSetText(&quot;Sending to webserver - 0%&quot;, &lt;0.3, 0.7, 0.2&gt;, 1.0);
                // now prepare this line for sending to web server

        string httpBody;
        string itemName;
        string myTimeStamp;
        integer i;
        integer length = llGetInventoryNumber(INVENTORY_ALL);
        serverKey = llGetKey();

        llSetTimerEvent(360.00); // timeout if the web server is too slow in responding

        // Now add the new items.
        // This needs two passes: on the first one, we will skip textures
        // The second pass will add them later
        llSetText(&quot;Checking inventory...&quot;, &lt;1.0,1.0,0.0&gt;, 1.0);

        for (i = 0; i &lt; length; i++)
        {
            itemName = llGetInventoryName(INVENTORY_ALL, i);

            if (llGetInventoryType(itemName) != INVENTORY_SCRIPT &amp;&amp; llGetInventoryType(itemName) != INVENTORY_TEXTURE) // skip script, skip textures
            {
                myTimeStamp = llGetTimestamp();

                httpBody =  &quot;name=&quot; + llEscapeURL(itemName) +
                            &quot;&amp;amp;timestamp=&quot; + myTimeStamp +
                            &quot;&amp;permissions=&quot; + (string) llGetInventoryPermMask(itemName, MASK_NEXT) +
                            &quot;&amp;itemType=&quot; + (string) llGetInventoryType(itemName) +
                            &quot;&amp;signature=&quot; + llMD5String((string)serverKey + myTimeStamp, LSLSignaturePIN);
                llSleep(1.0);

                httpRequestKey = llHTTPRequest(webServerURLupdateInventory,
                                [HTTP_METHOD, &quot;POST&quot;,
                                 HTTP_MIMETYPE,&quot;application/x-www-form-urlencoded&quot;],
                                httpBody);
                //llOwnerSay(&quot;Object &quot; + (string) i + &quot;: &quot; + httpBody);
                if (httpRequestKey == NULL_KEY)
                    llOwnerSay(&quot;Error contacting webserver on item #&quot; + (string)i);

                llSetText(&quot;Sending to webserver - &quot; + (string) ((integer)((float)i/(float)length*100)) + &quot;%&quot;, &lt;0.3, 0.7, 0.2&gt;, 1.0);
            }
        }
        state default;
    }

    http_response(key request_id, integer status, list metadata, string body)
    {
        llSetText(&quot;&quot;, &lt;0.0,0.0,0.0&gt;, 1.0);

        if (request_id == httpRequestKey)
        {
            if (status != 200)
            {
                llOwnerSay(&quot;HTTP Error &quot; + (string)status + &quot;: &quot; + body);
            }
            else
            {
                llOwnerSay(&quot;Web-server reply: &quot; + body);
                if (body == &quot;closed&quot;)
                    state default;
            }
        }
    }

    timer()
    {
        // HTTP server does not work, go to default state for now
        llOwnerSay(&quot;Web server did not reply after 6 minutes - not updated - trying again later&quot;);
        llSetTimerEvent(0.0);
        init();
    }

    state_exit()
    {
        llSetTimerEvent(0.0);
    }
}
</code></pre>
{{ end }}