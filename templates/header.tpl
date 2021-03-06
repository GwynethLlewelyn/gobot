{{ define "header" }}
<!DOCTYPE html>
<html lang="en">
<head>

	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="{{.Title}}">
	<meta name="author" content="Gwyneth Llewelyn">

	<title>{{.Title}}</title>

	<!-- Google Web Fonts -->
	<link href="https://fonts.googleapis.com/css2?family=Cantarell&family=Cardo&display=swap" type="text/css" rel="stylesheet">

	<!-- Bootstrap Core CSS 
	<link href="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/bootstrap/css/bootstrap.min.css" rel="stylesheet"> -->
	<!-- our modified bootstrap -->
	<link href="{{.URLPathPrefix}}/lib/bootstrap/css/bootstrap.min.css" rel="stylesheet" type="text/css">

	<!-- Bootstrap-Dialog -->
<!--	<link href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.34.7/css/bootstrap-dialog.min.css" rel="stylesheet" type="text/css"> -->
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.35.4/css/bootstrap-dialog.min.css" integrity="sha256-wstTM1F5dOf7cgnlRHIW3bmoRAAGh6jL7tMIvqTuFZE=" crossorigin="anonymous" />

	<!-- MetisMenu CSS -->
	<link href="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/vendor/metisMenu/metisMenu.min.css" rel="stylesheet" type="text/css">

	<!-- Custom CSS -->
	<link href="{{.URLPathPrefix}}/lib/startbootstrap-sb-admin-2/dist/css/sb-admin-2.css" rel="stylesheet" type="text/css">

	<!-- Custom Fonts -->
	<!--<link href="../vendor/font-awesome/css/font-awesome.min.css" rel="stylesheet" type="text/css">-->
	<link href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" type="text/css">
	
	{{ if .MapURL }}
	<!-- Call Leaflet.js to deal with maps -->
	<!-- Bumped to Leaflet 1.5.1 (20191116); Leaflet + extra markers need to be in the header or else this will bomb (20191117) -->
	<link rel="stylesheet" href="https://unpkg.com/leaflet@1.5.1/dist/leaflet.css"
		integrity="sha512-xwE/Az9zrjBIphAcBb3F6JVqxf46+CDLwfLMHloNu6KEQCAWi6HcDUbeOfBIptF7tcCzusKFjFw2yuvEpDL9wQ=="
		crossorigin=""/>
	<script src="https://unpkg.com/leaflet@1.5.1/dist/leaflet.js"
		integrity="sha512-GffPMF3RvMeYyc1LWMHtK8EbPv0iNZ8/oTtHPx9/cc2ILxQ+u905qIwdpULaqDkyBKgOaB57QTMg7ztg8Jm2Og=="
		crossorigin="">
	</script>

	<!-- This is to get cute markers on Leaflet maps -->
	<link rel="stylesheet" href="{{.URLPathPrefix}}/lib/Leaflet.vector-markers/dist/leaflet-vector-markers.css">
	<script src="{{.URLPathPrefix}}/lib/Leaflet.vector-markers/dist/leaflet-vector-markers.min.js"></script>
	{{ end }}
	
	{{ if .gobotJS }}
	<!-- Call agGrid -->
	<script src="{{.URLPathPrefix}}/lib/ag-grid/dist/ag-grid.min.js"></script>
	<script src="{{.URLPathPrefix}}/lib/gobot-js/{{.gobotJS}}"></script>
	{{ end }}

	{{ if .Gravatar }}
	<!-- I have no idea if this is really needed! -->
	<link rel="stylesheet" href="https://secure.gravatar.com/css/services.css" type="text/css">
	<link rel="stylesheet" href="{{.URLPathPrefix}}/lib/gravatar-profile.css" type="text/css">
	{{ end }}
	
	{{ if .LSL }}
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.6.0/themes/prism.min.css" type="text/css">
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.6.0/prism.min.js"></script>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.6.0/plugins/toolbar/prism-toolbar.min.css" type="text/css">
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.6.0/plugins/toolbar/prism-toolbar.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.6.0/plugins/copy-to-clipboard/prism-copy-to-clipboard.min.js"></script>
	{{ end }}

	<!-- Our own overrides -->
	<link href="{{.URLPathPrefix}}/lib/gobot.css" rel="stylesheet" type="text/css">

	<!-- HTML5 Shim and Respond.js IE8 support of HTML5 elements and media queries -->
	<!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
	<!--[if lt IE 9]>
		<script src="https://oss.maxcdn.com/libs/html5shiv/3.7.0/html5shiv.js"></script>
		<script src="https://oss.maxcdn.com/libs/respond.js/1.4.2/respond.min.js"></script>
	<![endif]-->

	<!-- stupid favicons, updated with new logo 20170621 -->
	<link rel="apple-touch-icon" sizes="180x180" href="{{.URLPathPrefix}}/apple-touch-icon.png">
	<link rel="icon" type="image/png" sizes="32x32" href="{{.URLPathPrefix}}/favicon-32x32.png">
	<link rel="icon" type="image/png" sizes="16x16" href="{{.URLPathPrefix}}/favicon-16x16.png">
	<link rel="manifest" href="{{.URLPathPrefix}}/manifest.json">
	<link rel="mask-icon" href="{{.URLPathPrefix}}/safari-pinned-tab.svg" color="#5bbad5">
	<meta name="apple-mobile-web-app-title" content="Gobot">
	<meta name="application-name" content="Gobot">
	<meta name="msapplication-TileColor" content="#00a300">
	<meta name="msapplication-TileImage" content="{{.URLPathPrefix}}/mstile-144x144.png">
	<meta name="theme-color" content="#ffffff">
	<!-- favicons end here -->
</head>
<body>
{{ if .Gravatar }}
<!-- Gravatar Hovercards are sneaky, they add their own CSS at the top of the header! -->
<style>
.gcard {
	z-index: 1000;
}
.emptyPlaceholder {
	z-index: 1000;
}
</style>
{{ end }}
<span id="URLPathPrefix" hidden>{{.URLPathPrefix}}</span>
{{ end }}