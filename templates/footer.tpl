{{ define "footer" }}
    <!-- jQuery -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/jquery/jquery.min.js"></script> 

    <!-- Bootstrap Core JavaScript 
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/bootstrap/js/bootstrap.min.js"></script> -->
    <!-- Our modified Bootstrap -->
    <script src="{{.URLPathPrefix}}/lib/bootstrap/js/bootstrap.min.js"></script>
    
    <!-- Bootstrap-Dialog -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.34.7/js/bootstrap-dialog.min.js"></script>
    
    <!-- Metis Menu Plugin JavaScript -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/metisMenu/metisMenu.min.js"></script>

    <!-- Custom Theme JavaScript -->
    <script src="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/dist/js/sb-admin-2.js"></script>
    
	{{ if .Gravatar }}
	<script src="https://secure.gravatar.com/js/gprofiles.js"></script>
	<script src="{{.URLPathPrefix}}/lib/gobot-js/wpgroho.js"></script>
	<div style="display:none">
		<div class="grofile-hash-map-{{ .GravatarHash }}">
		</div>
	</div>
	{{ end }}

</body>

</html>
{{ end }}