{{ define "engine" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper">
			<div class="container-fluid">
				<div class="row">
					<div class="col-lg-12">
						<h1 class="page-header">{{.Title}}</h1>
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
							window.onload = function () {
								var conn = null;
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