{{ define "objects" }}
{{ template "header" .}}
    <div id="wrapper">

{{ template "navigation" .}}

        <!-- Page Content -->
        <div id="page-wrapper">
            <div class="container-fluid">
                <div class="row">
                    <div class="col-lg-12">
                        <h1 class="page-header">{{.Title}}</h1>
                        Header path prefix is: {{.URLPathPrefix}} and<br />
						my content is: <br />
                        {{.Content}}
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