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
                        <div id="objectGrid" style="height: 440px;" class="ag-fresh"></div>
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