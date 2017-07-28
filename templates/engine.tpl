{{ define "engine" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper">
			<div class="container-fluid">
				<div class="row">			
				<h1 class="page-header">{{.Title}}</h1>
					<div class="col-lg-9">
						<div class="panel panel-default">
							<div class="panel-heading">Send manual commands to engine</div> 
							<div class="panel-body">
									<form role="form" id="formEngine" action="javascript:void(0);" onsubmit="formEngineSubmit()">
										<fieldset id="formEngineFieldSet">
										<div class="col-lg-9">						
											<div class="form-group">
												<label>Destination</label>
												<select class="form-control" name="Destination" id="Destination" size="1">
													<option value="00000000-0000-0000-0000-000000000000" selected="selected" disabled="disabled">Please choose a destination cube</option>
				{{.DestinationOptions}}
												</select>
												<label>Agent</label>
												<select class="form-control" name="Agent" id="Agent" size="1">
													<option value="00000000-0000-0000-0000-000000000000" selected="selected" disabled="disabled">Please choose an agent</option>
				{{.AgentOptions}}
												</select>
											</div> <!-- /.form-group -->
										</div> <!-- ./col-lg-9 -->
										<div class="col-lg-3">
											<button type="submit" class="btn btn-outline btn-success"><i class="fa fa-check"></i>&nbsp;Submit</button>
											<button type="reset" class="btn btn-outline btn-warning"><i class="fa fa-trash-o"></i>&nbsp;Reset</button>
										</div> <!-- ./col-lg-3 -->
										</fieldset>
									</form>
							</div> <!-- ./panel-body -->
						</div> <!-- ./panel -->
					</div> <!-- ./col-lg-9 -->
					<div class="col-lg-3">
						<div class="panel panel-default">					
							<div class="panel-heading">Engine master control</div> 
							<div class="panel-body">
									<button type="button" id="clearLog" class="btn btn-default btn-circle btn-xl"><i class="fa fa-trash-o" onclick="clearLog()"></i></button>
									<button type="button" id="startEngine" class="btn btn-success btn-circle btn-xl"><i class="fa fa-check-circle" onclick="startEngine()"></i></button>
									<button type="button" id="stopEngine" class="btn btn-danger btn-circle btn-xl"><i class="fa fa-times-circle" onclick="stopEngine()"></i></button>
							</div> <!-- ./panel-body -->
						</div> <!-- ./panel -->	
					</div> <!-- ./col-lg-3 -->									
				</div> <!-- ./row -->
				<div id="alertMessage" class="alert alert-warning alert-dismissable" hidden style="display: none;">
                	<button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
					<p id="message">No valid active agents or destination cubes found.</p>
                </div>
				<div class="row">
					<div class="col-lg-12">
						<style type="text/css">
						#log {
						    background: white;
						    margin: 0;
						    padding: 0.5em 0.5em 0.5em 0.5em;
						    //position: absolute;
						    top: 0.5em;
						    left: 0.5em;
						    right: 0.5em;
						    bottom: 3em;
						    overflow: auto;
						}
						</style>
						<div id="engineResponse" name="engineResponse" contenteditable="true"></div>
						<!-- websockets will fill this in -->
						<script type="text/javascript">
							const wsuri = "ws://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/wsEngine/";
							var conn = null; // WebSocket connection, must be sort-of-global-y (20170703)
							
							function formEngineConfig(enableStatus) {
								document.getElementById("formEngine").disabled = enableStatus;
								document.getElementById("formEngineFieldSet").disabled = enableStatus;
							}
							
							/*
							 * WebSocket handler, called only when the window has loaded
							 *  and every time we suspect that the connection was closed (20170703)
							 */
							function start() {
								if (window["WebSocket"]) {
									// start by preparing the scrollable response element
									var log = document.getElementById("engineResponse");
									log.height = 400;
									log.scrollTop = log.scrollHeight; // scroll to bottom - http://web.archive.org/web/20080821211053/http://radio.javaranch.com/pascarello/2005/12/14/1134573598403.html

									// now deal with the WebSocket
									conn = new WebSocket(wsuri);
									console.log("Attempting to connect to WebSocket on " + wsuri);
									conn.binaryType = "arraybuffer";
									conn.onopen = function() {
										console.log("Now connected to " + wsuri);
										formEngineConfig(false); // enable form, we can accept ws connections
										// when this is started, send a message to the server to tell that
	            						//  we are ready to receive messages:
	            						var msg = {
											type: "status",
											subtype: "ready",
											text: "Client is ready now",
											id: ""
										};

										 // Send the msg object as a JSON-formatted string.
										conn.send(JSON.stringify(msg));
										setInterval(check, 5000); // check if connection is active every five seconds
            						}
									conn.onclose = function(evt) {
										//var msg = JSON.parse(evt.data);
										console.log("Connection closed; data received: " + evt.data);
										log.innerHTML += 'Connection closed - trying to reconnect<br />';
										log.scrollTop = log.scrollHeight;
										check(); // this should disable form as well
									}
									// Handler to deal with messages coming from the server (20170702)
									conn.onmessage = function(evt) {
										// see https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_client_applications
										var msg;
										
										if (typeof evt.data == "string") {
											msg = JSON.parse(evt.data);
										} else if (evt.data instanceof ArrayBuffer) {
											// Note: sometimes the websocket data comes as an ArrayBuffer, when the data is structured; sometimes it doesn't. To figure out both cases, we force the WebSocket to return an ArrayBuffer and extract the relevant JSON strings from there: https://developers.google.com/web/updates/2014/08/Easier-ArrayBuffer-String-conversion-with-the-Encoding-API (20170702)
											
									    	// The decode() method takes a DataView as a parameter, which is a wrapper on top of the ArrayBuffer.
									        var dataView = new DataView(evt.data);
									        // The TextDecoder interface is documented at http://encoding.spec.whatwg.org/#interface-textdecoder
									        var decoder = new TextDecoder("utf-8");
									        var decodedString = decoder.decode(dataView);
											
											// Now we should have a string in JSON
											msg = JSON.parse(decodedString);
										} else {
											// I have no idea how to decode any other type
											console.log("Unexpected type of event data, attempt to create a message out of it, even if it may fail.");
											msg = evt.data;
										}
										
										var logTxt = "";
										// check for message type from the server, right now it may be:
										//  status: just add that message to the scrollable element
										//  htmlControl: to turn on/off certain elements of the form
										//   (to disallow submission of new data while the engine
										//    has been manually ran for one Agent/Destination pair)
										switch (msg.type) {
											case "status":
												logTxt = msg.text;
												break;
											case "htmlControl":
												switch (msg.subtype) {
													case "disable":
														document.getElementById(msg.id).disabled = true;
														logTxt = "<em>Element " + msg.id + " disabled</em><br />";
														break;
													case "enable":
														document.getElementById(msg.id).disabled = false;
														logTxt = "<em>Element " + msg.id + " enabled</em><br />";
														break;
													default:
														logTxt = "Unknown subtype '" + msg.subtype + "'<br />";
														break;
												};
												break;
											default:
												logTxt = "Unknown type '" + msg.type + "' with text '" +
													msg.text + "'<br />";
												break;
										};
										// console.log('Received from server: (' + msg.type + '): "' +
										//	msg.text + '"; writing on log: "' + logTxt + '"');
										log.innerHTML += logTxt;
										log.scrollTop = log.scrollHeight;
									}
									conn.onerror = function(err) {
										console.log("Error from WebSocket: " + err.data);
										log.innerHTML += "Error from WebSocket: " + err.data + "<br />";
										log.scrollTop = log.scrollHeight;
										check();
            						}
            						

            					} else {
									log.innerHTML += "<b>Your browser does not support WebSockets.</b><br />";
									log.scrollTop = log.scrollHeight;
								}
							}
							
							/*
							 *	Main handling starts here, when we know that everything has been loaded
							 */
							window.onload = function () {
								// Disable form if we have no destination cubes or active agents
								if ("{{ .DestinationOptions }}" == "" || "{{ .AgentOptions }}" == "") {
									formEngineConfig(true);
									// we might give an explanation here, e.g. enable a hidden field
									//  and put an answer there
									document.getElementById("alertMessage").style.display = 'block';
								}
								if (!conn || conn.readyState === WebSocket.CLOSED) {
									// if we have no valid connection yet, do not allow the form to be submitted (20170703)
									formEngineConfig(true);
								}
								
								start();
							};
							
							window.onunload = function () {
								if (conn.readyState === WebSocket.OPEN) {							
									console.log("Cool, we are able to send a message informing the server that we are gone!");
									var msg = {
											type: "status",
											subtype: "gone",
											text: "",
											id: ""
										};
										
									conn.send(JSON.stringify(msg));
								}
							}
							
							// see https://stackoverflow.com/questions/3780511/reconnection-of-client-when-server-reboots-in-websocket
							function check() {
								// check if connection is ready; if not, try to start it again
								if(!conn || conn.readyState === WebSocket.CLOSED) {
									formEngineConfig(true); // disable form for now
									start();
								}
							}
							
							function formEngineSubmit() {
								// make sure we have a valid connection to the server
								console.log("formEngine was submitted!");
								if (!conn || conn.readyState === WebSocket.CLOSED) {
									console.log("Connection closed while sending formEngine data; trying to restart...");
									formEngineConfig(true); // to be sure the form remains disabled
									start();				//  we could call check() here... (20170703)
									return false;
								} else if (conn.readyState === WebSocket.OPEN) {							
									console.log("Connection OK while sending formEngine data - sending...");
									var msg = {
											type: "formSubmit",
											subtype: "",
											text: document.forms["formEngine"]["Destination"].value + '|' + 
												document.forms["formEngine"]["Agent"].value,
											id: ""
										};
										
									conn.send(JSON.stringify(msg));
									return true;
								} else {
									console.log("Connection either starting or closing, not ready yet for sending — what to do?");
									// probably needs to be addressed somehow (20170703)
									return false;
								}
							}
							
							// This is just to get us an empty log again
							function clearLog() {
								var log = document.getElementById("engineResponse");
								log.innerHTML = "<p class='text-muted'><small><i>(log cleared on request)</i></small></p>";
								log.scrollTop = log.scrollHeight;
								window.setTimeout(function () {
									//log.innerHTML = "";
									log.scrollTop = log.scrollHeight;
								},5000);								
							}
							
							function startEngine() {
								if (conn.readyState === WebSocket.OPEN) {
									var msg = {
											type: "engineControl",
											subtype: "start",
											text: "Start Engine NOW!",
											id: "startEngine"
										};
									conn.send(JSON.stringify(msg));
									return true;
								}
								document.getElementById("message").innerHTML = "WebSocket not connected!";
								document.getElementById("alertMessage").style.display = 'block';
								return false;
							}
							
							function stopEngine() {
								if (conn.readyState === WebSocket.OPEN) {
									var msg = {
											type: "engineControl",
											subtype: "stop",
											text: "Stop Engine NOW!",
											id: "stopEngine"
										};
									conn.send(JSON.stringify(msg));
									return true;
								}
								document.getElementById("message").innerHTML = "WebSocket not connected!";
								document.getElementById("alertMessage").style.display = 'block';
								return false;
							}
						 </script>
						 <noscript>Look, if you don't even bother to turn on JavaScript, you will get nothing.</noscript>
						{{ if .Content }}
						{{ .Content }}
						{{ end }}
						{{ if .ButtonText }}
						<a href="{{.URLPathPrefix}}{{ .ButtonURL }}">
							<button type="button" class="btn btn-outline btn-primary btn-lg">{{ .ButtonText }}</button>
						</a>
						{{ end }}
					</div>
					<!-- /.col-lg-12 -->
				</div>
				<!-- /.row -->
			</div>
			<!-- /.container-fluid -->
		</div>
		<!-- /#page-wrapper -->

	</div>
	<!-- /#wrapper -->
{{ template "footer" .}}
{{ end }}