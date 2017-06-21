{{ define "top-menu" }}
            <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                    <span class="sr-only">Toggle navigation</span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/agents/"><i class="fa fa-android fa-fw"></i> Agents</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/positions/"><i class="fa fa-codepen fa-fw"></i> Positions</a></span>
                    <span class="icon-bar"><span class="icon-bar"a href="{{.URLPathPrefix}}/admin/inventory/"><i class="fa fa-folder-open-o fa-fw"></i> Content/Inventory</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/objects/"><i class="fa fa-cubes fa-fw"></i> Obstacles (Objects)</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/commands/"><i class="fa fa-android fa-fw"></i>Commands to Bot</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/controller-commands/"><i class="fa fa-codepen fa-fw"></i>Commands to Controller</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/engine/"><i class="fa fa-gears fa-fw"></i> Engine</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/user-management/"><i class="fa fa-users fa-fw"></i> Manage Users</a></span>
                    <span class="icon-bar"><a href="#"><i class="fa fa-gear fa-fw"></i> Settings</a></span>
                    <span class="icon-bar"><a href="{{.URLPathPrefix}}/admin/logout/"><i class="fa fa-sign-out fa-fw"></i> Logout</a></span>
                </button>
                <a class="navbar-brand" href="{{.URLPathPrefix}}/admin/"><img src="{{.URLPathPrefix}}/templates/images/gobot-logo-menunav.png" height="32" alt="{{.Title}}"></a>
            </div>
            <!-- /.navbar-header -->

            <ul class="nav navbar-top-links navbar-right">	            
	            {{ if .SetCookie }}
	            <li id="#username">
                {{ .SetCookie }}
                {{ if .GravatarMenu }}
                <div style="float:left;"><a href="https://gravatar.com/{{ .GravatarHash }}"><img class="avatar avatar-{{ .GravatarSizeMenu }} photo" src="{{ .GravatarMenu }}" height="{{ .GravatarSizeMenu }}" width="{{ .GravatarSizeMenu }}" alt="{{ .SetCookie }}"></a></div>
                {{ end }}
	            </li>
                {{ end }}
                <li>
                    <a href="{{.URLPathPrefix}}/admin/"><i class="fa fa-dashboard fa-fw"></i> Dashboard</a>
                </li>
                <li class="dropdown">
                    <a class="dropdown-toggle" data-toggle="dropdown" href="#">
                    	<i class="fa fa-database fa-fw"></i> Database <i class="fa fa-caret-down"></i>
                    </a>
                    <ul class="dropdown-menu">
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
                    <!-- /.dropdown-menu -->
                </li>
                <li class="dropdown">
                    <a class="dropdown-toggle" data-toggle="dropdown" href="#">
                    	<i class="fa fa-terminal fa-fw"></i> Commands <i class="fa fa-caret-down"></i>
                    </a>
                    <ul class="dropdown-menu">                
						<li>
		                	<a href="{{.URLPathPrefix}}/admin/commands/"><i class="fa fa-android fa-fw"></i> To Bot</a>
		                </li>
						<li>
		                	<a href="{{.URLPathPrefix}}/admin/controller-commands/"><i class="fa fa-codepen fa-fw"></i> To Controller</a>
		                </li>
                    </ul>
                    <!-- /.dropdown-menu -->
                </li>
                <!-- /.dropdown (commands) -->
                <li>
                    <a href="{{.URLPathPrefix}}/admin/engine/"><i class="fa fa-gears fa-fw"></i> Engine</a>
                </li>
                <li class="dropdown">
                    <a class="dropdown-toggle" data-toggle="dropdown" href="#">
                        <i class="fa fa-user fa-fw"></i> <i class="fa fa-caret-down"></i>
                    </a>
                    <ul class="dropdown-menu dropdown-user">
                        <li><a href="{{.URLPathPrefix}}/admin/user-management/"><i class="fa fa-users fa-fw"></i> Manage Users</a>
                        </li>
                        <li><a href="#"><i class="fa fa-gear fa-fw"></i> Settings</a>
                        </li>
                        <li class="divider"></li>
                        <li><a href="{{.URLPathPrefix}}/admin/logout/"><i class="fa fa-sign-out fa-fw"></i> Logout</a>
                        </li>
                    </ul>
                    <!-- /.dropdown-user -->
                </li>
                <!-- /.dropdown -->
            </ul>
            <!-- /.navbar-top-links -->
{{ end }}