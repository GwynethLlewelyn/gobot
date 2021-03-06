{{ define "sidebar-left-menu" }}
            <div class="navbar-default sidebar" role="navigation">
                <div class="sidebar-nav navbar-collapse">
                    <ul class="nav" id="side-menu">
                        <li class="sidebar-search">
                            <div class="input-group custom-search-form">
                                <input type="text" class="form-control" placeholder="Search...">
                                <span class="input-group-btn">
                                    <button class="btn btn-default" type="button">
                                        <i class="fa fa-search"></i>
                                    </button>
                                </span>
                            </div>
                            <!-- /input-group -->
                        </li>
                        <li>
                            <a href="{{.URLPathPrefix}}/admin/"><i class="fa fa-dashboard fa-fw"></i> Dashboard</a>
                        </li>
                        <li>
                            <a href="{{.URLPathPrefix}}/admin/agents/"><i class="fa fa-android fa-fw"></i> Agents</a>
                        </li>
                        <li>
                            <a href="{{.URLPathPrefix}}/admin/positions/"><i class="fa fa-codepen fa-fw"></i> Positions</a>
                        </li>
                        <li>
                            <a href="{{.URLPathPrefix}}/admin/inventory/"><i class="fa fa-folder-open-o fa-fw"></i> Content/Inventory</a>
                        </li>
                        <li>
                            <a href="{{.URLPathPrefix}}/admin/objects/"><i class="fa fa-cubes fa-fw"></i> Obstacles (Objects)</a>
                        </li>
                    </ul>
                </div>
                <!-- /.sidebar-collapse -->
            </div>
            <!-- /.navbar-static-side -->

{{ end }}