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
                        {{ if .Agents }}
                        <div class="col-lg-6">
	                        <h2>Statistics</h2>
	                        <ul>
		                        <li>
		                        	{{.Agents}}
		                        </li>
		                        <li>
		                        	{{.Inventory}}
		                        </li>
		                        <li>
		                        	{{.Positions}}
		                        </li>
		                        <li>
		                        	{{.Obstacles}}
		                        </li>
	                        </ul>
							{{ if .SetCookie }}
							<p>Welcome, {{ .SetCookie }}</p>
							{{ end }}
                        </div> <!-- ./col-lg-6 -->
                        {{ end }}

                        {{ if .Content }}
                        {{ .Content }}
                        {{ end }}
                        {{ if .MapURL }}
						<div class="col-lg-6">
							<h2>In-world map</h2>
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
						</div> <!-- ./col-lg-6 -->
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