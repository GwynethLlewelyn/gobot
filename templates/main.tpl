{{ define "main" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper">
			<div class="container-fluid">
				<div class="row">
					<div class="col-lg-12">
						{{ if .Title }}
						<h1 class="page-header">{{.Title}}</h1>
						{{ end }}
						<div class="col-lg-6">
							{{ if .Agents }}
							<div class="panel panel-default">
								<div class="panel-heading">
									Statistics
								</div>
								<!-- /.panel-heading -->
								<div class="panel-body"> 	
									<ul>
										<li>{{.Agents}}</li>
										<li>{{.Inventory}}</li>
										<li>{{.Positions}}</li>
										<li>{{.Obstacles}}</li>
									 </ul>
								</div>
								<!-- /.panel-body -->
							</div>
							<!-- /.panel -->
							{{ end }}
							{{ if .ObstaclePieChart }}
							<div class="panel panel-default">
								<div class="panel-heading">
									Obstacles (and Phantom objects)
								</div>
								<!-- /.panel-heading -->
								<div class="panel-body">
									<div class="flot-chart">
										<div class="flot-chart-content" id="flot-pie-chart"></div>
									</div>
								</div>
								<!-- /.panel-body -->
							</div>
							<!-- /.panel -->
							{{ end }}
						</div> <!-- ./col-lg-6 -->
						<div class="col-lg-6">
							{{ if .SetCookie }}
							<div class="panel panel-default">
								<div class="panel-heading">
									User
								</div>
								<!-- /.panel-heading -->
								<div class="panel-body">            
									{{ if .Gravatar }}
									<div style="float:left;">
										<a href="https://gravatar.com/{{ .GravatarHash }}">
											<img class="avatar avatar-{{ .GravatarSize }} photo" src="{{ .Gravatar }}" height="{{ .GravatarSize }}" width="{{ .GravatarSize }}" alt="{{ .SetCookie }}">
										</a>
									</div>
									{{ end }}
									<div style="float:right;">
										Welcome, {{ .SetCookie }}
									</div>
								</div>
								<!-- /.panel-body -->
							</div>
							<!-- /.panel -->								
							{{ end }}
							{{ if .MapURL }}
							<div class="panel panel-default">
								<div class="panel-heading">
									In-world map
								</div>
								<!-- /.panel-heading -->
								<div class="panel-body"> 								
									<div id="map"></div>
									<script type="text/javascript">
										// create a few cute markers
										var agentMarker = L.VectorMarkers.icon({
											icon: 'android',
											markerColor: 'green'
		  								});
										var objectMarker = L.VectorMarkers.icon({
											icon: 'cubes',
											markerColor: 'blue'
		  								});
		  								var positionMarker = L.VectorMarkers.icon({
											icon: 'codepen',
											markerColor: 'red'
		  								});
		
										// set up the map
										var map = L.map('map', {
											attributionControl: false,
											crs: L.CRS.Simple,
											minZoom: -3
										});
		
										var bounds = [[0,0], [255,255]];
										var image = L.imageOverlay({{ .MapURL }}, bounds).addTo(map);
										map.fitBounds(bounds);
										{{ .MapMarkers }}
										map.setView([128, 128], 1);
									</script>
								</div>
								<!-- /.panel-body -->
							</div>
							<!-- /.panel -->		
							{{ end }}
						</div> <!-- ./col-lg-6 -->
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