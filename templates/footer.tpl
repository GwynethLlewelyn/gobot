{{ define "footer" }}
    <!-- jQuery -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/jquery/jquery.min.js"></script> 

    <!-- Bootstrap Core JavaScript 
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/bootstrap/js/bootstrap.min.js"></script> -->
    <!-- Our modified Bootstrap -->
    <script src="{{.URLPathPrefix}}/lib/bootstrap/js/bootstrap.min.js"></script>
    
    <!-- Bootstrap-Dialog -->
    <!-- <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.34.7/js/bootstrap-dialog.min.js"></script> -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.35.4/js/bootstrap-dialog.min.js" integrity="sha256-IpgnbT7iaNM6j9WjtXKI8VMJ272WM9VvFYkZdu1umOA=" crossorigin="anonymous"></script>
    
    <!-- Metis Menu Plugin JavaScript -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/metisMenu/metisMenu.min.js"></script>

    <!-- Custom Theme JavaScript -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/dist/js/sb-admin-2.js"></script>
    
	{{ if .Gravatar }}
	<!-- this does not work yet -->
	<script src="{{.URLPathPrefix}}/lib/gobot-js/gprofiles.js"></script>
	<script src="{{.URLPathPrefix}}/lib/gobot-js/wpgroho.js"></script>
	{{ end }}
	
	{{ if .MapURL }}
	<script src="https://unpkg.com/leaflet@1.5.1/dist/leaflet.js"
		integrity="sha512-GffPMF3RvMeYyc1LWMHtK8EbPv0iNZ8/oTtHPx9/cc2ILxQ+u905qIwdpULaqDkyBKgOaB57QTMg7ztg8Jm2Og=="
		crossorigin="">
	</script>
	
	<!-- This is to get cute markers on Leaflet maps -->
	<script src="{{.URLPathPrefix}}/lib/Leaflet.vector-markers/dist/leaflet-vector-markers.min.js"></script>
	{{ end }}

	{{ if .ObstaclePieChart }}
    <!-- Flot Charts JavaScript -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot/excanvas.min.js"></script>
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot/jquery.flot.js"></script>
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot/jquery.flot.pie.js"></script>
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot/jquery.flot.resize.js"></script>
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot/jquery.flot.time.js"></script>
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/flot-tooltip/jquery.flot.tooltip.min.js"></script>
    <script type="text/javascript">
	    //data for Flot Pie Chart
		$(function() {
		
		    var data = [{
		        label: "Objects",
		        data: {{ .obstaclesCnt }}
		    }, {
		        label: "Phantom",
		        data: {{ .phantomCnt }}
		    }];
		
		    var plotObj = $.plot($("#flot-pie-chart"), data, {
		        series: {
		            pie: {
		                show: true,
		                tilt: 0.35,
		                label: {
			                show: true,
			                radius: 3/4,
			                /*formatter: labelFormatter,*/
			                background: {
								color: '#000',
								opacity: 0.7
			                }
            			}
		            }
		        },
		        grid: {
		            hoverable: true
		        },
		        tooltip: true,
		        tooltipOpts: {
		            content: "%s: %n (%p.0%)", // show percentages, rounding to 2 decimal places
		            shifts: {
		                x: 20,
		                y: 0
		            },
		            defaultTheme: false
		        }
		    });
		
		});
    </script>
    {{ end }}

</body>

</html>
{{ end }}