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
							{{ if and .SetCookie (not .LSL) }}
							<div class="panel panel-default">
								<div class="panel-heading">
									User
								</div>
								<!-- /.panel-heading -->
								<div class="panel-body">            
									{{ if .Gravatar }}
									<div style="float:left;">
										<a href="https://gravatar.com/{{ .GravatarHash }}" title="{{ .SetCookie }}">
											<img class="avatar avatar-{{ .GravatarSize }} photo" src="{{ .Gravatar }}" srcset="https://secure.gravatar.com/avatar/{{ .GravatarHash }}?s={{ .GravatarTwiceSize }}&amp;d=mm&amp;r=r 2x" height="{{ .GravatarSize }}" width="{{ .GravatarSize }}" alt="{{ .SetCookie }}">
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
										
										/*
										var mapMinZoom = 1
										var mapMaxZoom = 6
										
										L.Projection.Direct = {
											project: function (latlng) {
												return new L.Point(latlng.lat*256, latlng.lng*256);
											},
											unproject: function (point) {
												return new L.LatLng(point.x/256, point.y/256);
											}
										};
										
										L.CRS.OpenSim=L.extend({},L.CRS,{
											projection: L.Projection.Direct,
											transformation:new L.Transformation(1,0,1,0),
										
											scale: function (zoom) {
												return 1;// OpenSim zoom
											}
										});
										
										var map = L.map('map',{
											attributionControl: false,
											minZoom: mapMinZoom,
											maxZoom: mapMaxZoom,
											crs: L.CRS.OpenSim
										});
										
										map.setView([3646, 3645], 1);

										L.tileLayer('http://opensim.betatechnologies.info:8002/map-{z}-{x}-{y}-objects.jpg', {
											maxZoom: mapMaxZoom,
											continuousWorld: true,
											noWrap:true,
											tileSize:256,
											crs: L.CRS.OpenSim,
											attribution: 'opensim',
											id: 'opensim',
										}).addTo(map);
										
										map.panTo([3646, 3645]);
										*/
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
						{{ if .LSL }}
							{{ if eq .LSL "lsl-register-object" }}
							{{ template "lsl-register-object" . }}
							{{ end }}
							{{ if eq .LSL "lsl-bot-controller" }}
							{{ template "lsl-bot-controller" . }}
							{{ end }}					
						<script>hljs.initHighlightingOnLoad();</script>
						{{ end }}
						{{ if .ButtonText }}
						<a href="{{.URLPathPrefix}}{{ .ButtonURL }}">
							<button id={{.ButtonID}} type="button" class="btn btn-outline btn-primary btn-lg">{{ .ButtonText }}</button>
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