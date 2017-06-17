{{ define "inventory" }}
{{ template "header" .}}
    <div id="wrapper">

{{ template "navigation" .}}

        <!-- Page Content -->
        <div id="page-wrapper">
            <div class="container-fluid">
                <div class="row">
                    <div class="col-lg-12">
                        <h1 class="page-header">{{.Title}}</h1>
                        <div id="inventoryGrid" style="height: 440px;" class="ag-fresh"></div>
                        <br />
                        <button type="button" class="btn btn-outline btn-primary" onclick="onInsertRow()">Insert Row</button>&nbsp;
                        <button type="button" class="btn btn-outline btn-warning" onclick="onRemoveSelected()">Remove Selected</button>
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