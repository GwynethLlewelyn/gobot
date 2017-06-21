{{ define "engine" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper">
			<div class="container-fluid">
				<div class="row">			
				<h1 class="page-header">{{.Title}}</h1>
					<form role="form" id="formEngine" action="javascript:void(0);" onsubmit="formEngineSubmit()">
						<fieldset id="formEngineFieldSet">
						<div class="col-lg-6">						
							<div class="form-group">
								<label>Destination</label>
								<select class="form-control" name="Destination" id="Destination" size="1">
									<option value="0" selected="selected" disabled="disabled">Please choose a destination cube</option>
{{.DestinationOptions}}
								</select>
								<label>Agent</label>
								<select class="form-control" name="Agent" id="Agent" size="1">
									<option value="0" selected="selected" disabled="disabled">Please choose an agent</option>
{{.AgentOptions}}
								</select>
							</div> <!-- /.form-group -->
						</div> <!-- ./col-lg-6 -->
						<div class="col-lg-6">
							<button type="submit" class="btn btn-default">Submit</button>
							<button type="reset" class="btn btn-default">Reset</button>
						</div> <!-- ./col-lg-6 -->
						</fieldset>
					</form>
				</div> <!-- ./row -->
				<div id="alertMessage" class="alert alert-warning alert-dismissable" hidden style="display: none;">
                	<button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
					<p id="message">No valid active agents or destination cubes found.</p>
                </div>
				<div class="row">
					<div class="col-lg-12">
						Results from the engine:
						<hr />
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
							var conn = null;
							
							window.onload = function () {
								// Disable form if we have no destination cubes or active agents
								if ("{{ .DestinationOptions }}" == "" || "{{ .AgentOptions }}" == "") {
									document.getElementById("formEngine").disabled = true;
									document.getElementById("formEngineFieldSet").disabled = true;
									// we might give an explanation here, e.g. enable a hidden field
									//  and put an answer there
									document.getElementById("alertMessage").style.display = 'block';
								} 
								
								// now deal with the WebSocket
								var log = document.getElementById("engineResponse");
								log.height = 400;
								log.scrollTop = log.scrollHeight; // scroll to bottom - http://web.archive.org/web/20080821211053/http://radio.javaranch.com/pascarello/2005/12/14/1134573598403.html
								var wsuri = "ws://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/wsEngine/";
								
							    function start() {
									if (window["WebSocket"]) {
										conn = new WebSocket(wsuri);
										conn.onopen = function() {
											console.log("connected to " + wsuri);
	            						}
										conn.onclose = function(evt) {
											console.log("Connection closed; data received: " + evt.data);
											log.innerHTML += 'Connection closed - trying to reconnect<br />';
											log.scrollTop = log.scrollHeight;
											check();
										}
										conn.onmessage = function(evt) {
											// see https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_client_applications
											
											console.log('Got an update: "' + evt.data + '"');
											log.innerHTML += evt.data;
											log.scrollTop = log.scrollHeight;
										}
										conn.onerror = function(err) {
											console.log("Error from WebSocket: " + err.data)
											log.innerHTML += "Error from WebSocket: " + err.data + "<br />";
											log.scrollTop = log.scrollHeight;
											check();
	            						}
	            						conn.send("Client is ready now");
									} else {
										log.innerHTML += "<b>Your browser does not support WebSockets.</b><br />";
										log.scrollTop = log.scrollHeight;
									}
								}
								
								// see https://stackoverflow.com/questions/3780511/reconnection-of-client-when-server-reboots-in-websocket
								function check() {
									if(!conn || conn.readyState === WebSocket.CLOSED) start();
								}
								
								start();
								
								setInterval(check, 5000);
							};
							
							function formEngineSubmit() {
								// make sure we have a valid connection to the server
								console.log("formEngine was submitted!");
								if (!conn || conn.readyState === WebSocket.CLOSED) {
									start();
								} else {
									conn.send("Ready to launch engine!");
								}
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