{{ define "main" }}
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
						<p id="#engineResponse"><!-- websockets will fill this in --></p>
							<script type="text/javascript">
								(function() {
									var data = document.getElementById("engineResponse");
									var conn = new WebSocket("ws://{{.Host}}/ws?lastMod={{.LastMod}}");
									conn.onclose = function(evt) {
										data.textContent = 'Connection closed';
									}
									conn.onmessage = function(evt) {
										console.log('file updated');
										data.textContent = evt.data;
									}
								})();
							 </script>
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