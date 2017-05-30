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
                        Header path prefix is: {{.URLPathPrefix}} and<br />
						my content is: <br />
                        {{.Content}}
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