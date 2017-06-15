{{ define "engine" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper mychange">
			<div class="container-fluid">
				<div class="row">
					<div class="col-lg-12">
						<h1 class="page-header">{{.Title}}</h1>
						Results from the engine:
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
						<div id="engineResponse"></div>
						<!-- websockets will fill this in -->
						<script type="text/javascript">
							window.onload = function () {
								var conn;
								var log = document.getElementById("engineResponse");
								console.log('My log div is', log);
								function appendLog(item) {
							        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
							        log.appendChild(item);
							        if (doScroll) {
							            log.scrollTop = log.scrollHeight - log.clientHeight;
							        }
							    }
								if (window["WebSocket"]) {
									var conn = new WebSocket("ws://{{.Host}}{{.ServerPort}}{{.URLPathPrefix}}/wsEngine/");
									conn.onclose = function(evt) {
										appendLog('Connection closed');
									}
									conn.onmessage = function(evt) {
										console.log('Got an update' + evt.data);
										appendLog(evt.data);
									}
								} else {
									appendLog("<b>Your browser does not support WebSockets.</b>");
								}
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