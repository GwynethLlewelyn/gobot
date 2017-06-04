{{ define "commands" }}
{{ template "header" .}}
	<div id="wrapper">

{{ template "navigation" .}}

		<!-- Page Content -->
		<div id="page-wrapper mychange">
			<div class="row">
				<div class="col-lg-12">
					<h1 class="page-header">{{.Title}}</h1>
				</div>
				<!-- /.col-lg-12 -->
			</div>
			<!-- /.row -->
			<div class="row">
				<div class="col-lg-12">
					<div class="panel panel-default">
						<div class="panel-heading">
							Basic Form Elements
						</div>
						<div class="panel-body">
							<form role="form" action="{{.URLPathPrefix}}/admin/commands/exec/" method="post">
							<div class="row">
								<div class="col-lg-6">					
									<div class="form-group">
										<label>Enter your command PermURL</label>
										<select class="form-control" name="PermURL" id="PermURL" size="1">
											<option value="0" selected="selected" disabled="disabled">Please choose an avatar</option>
											{{.AvatarPermURLOptions}}
										</select>
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->
								<div class="col-lg-6">					
									<div class="form-group">
										<label>Commands</label>
										<select class="form-control" name="command" id="command" size="1">
											<option value="0" selected="selected" disabled="disabled">Please choose</option>
											<optgroup label="Get and Set">
												<option value="osNpcGetRot">osNpcGetRot</option>
												<option value="osNpcSetRot">osNpcSetRot</option>
												<option value="osNpcGetPos">osNpcGetPos</option>
												<option value="osNpcGetOwner">osNpcGetOwner</option>
											</optgroup>
											<optgroup label="Movement">
												<option value="osNpcMoveToTarget">osNpcMoveToTarget</option>
												<option value="osNpcStopMoveToTarget">osNpcStopMoveToTarget</option>
											</optgroup>
											<optgroup label="Sitting and Standing">
												<option value="osNpcSit">osNpcSit</option>
												<option value="osNpcStand">osNpcStand</option>
											</optgroup>
											<optgroup label="Communication">
												<option value="osNpcSay">osNpcSay</option>
												<option value="osNpcShout">osNpcShout</option>
												<option value="osNpcWhisper">osNpcWhisper</option>
											</optgroup>
											<optgroup label="Animations">
												<option value="osNpcPlayAnimation">osNpcPlayAnimation</option>
												<option value="osNpcStopAnimation">osNpcStopAnimation</option>
											</optgroup>
											<optgroup label="Appearance">
												<option value="osNpcLoadAppearance">osNpcLoadAppearance</option>
											</optgroup>						   
											<optgroup label="Touch">
												<option value="osNpcTouch">osNpcTouch</option>
											</optgroup>
											<optgroup label="Engine Get Values">
												<option value="getMoney">getMoney</option>
												<option value="getHappiness">getHappiness</option>
												<option value="getEnergy">getEnergy</option>
												<option value="getHome">getHome</option>
												<option value="getClass">getClass</option>
												<option value="getSubType">getSubType</option>
											</optgroup>									  
											<optgroup label="Engine Set Values">
												<option value="setMoney">setMoney</option>
												<option value="setHappiness">setHappiness</option>
												<option value="setEnergy">setEnergy</option>
												<option value="setHome">setHome</option>
												<option value="setClass">setClass</option>
												<option value="setSubType">setSubType</option>
											</optgroup>	
										</select>
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->									  
							</div> <!-- /.row -->
							<div class="row">
								<div class="col-lg-6">					
									<div class="form-group">
										<label>Parameter #1 Type</label>
										<select class="form-control" name="param1" id="param1" size="1">
											<option value="string" selected="selected">string</option>
											<option value="key">key</option>
											<option value="int">int</option>
											<option value="float">float</option>
											<option value="vector">vector</option>
										</select>
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->
								<div class="col-lg-6">					
									<div class="form-group input-group">
										<label>Parameter #1 Data</label>
										<input class="form-control" type="text" name="data1" id="data1" placeholder="Enter data">
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->
							</div> <!-- /.row -->
							<div class="row">
								<div class="col-lg-6">					
									<div class="form-group">
										<label>Parameter #2 Type</label>
										<select class="form-control" name="param2" id="param2" size="1">
											<option value="string" selected="selected">string</option>
											<option value="key">key</option>
											<option value="int">int</option>
											<option value="float">float</option>
											<option value="vector">vector</option>
										</select>
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->
								<div class="col-lg-6">					
									<div class="form-group input-group">
										<label>Parameter #2 Data</label>
										<input class="form-control" type="text" name="data2" id="data2" placeholder="Enter data">
									</div> <!-- /.form-group -->
								</div> <!-- /.col-lg.6 -->
							</div> <!-- /.row -->
							<button type="submit" class="btn btn-default">Submit</button>
							<button type="reset" class="btn btn-default">Reset</button>
						</form>
						</div>
						<!-- /.panel-body -->
					</div>
					<!-- /.panel -->
				</div>
				<!-- /.col-lg-12 -->
			</div>
			<!-- /.row -->
		</div>
		<!-- /#page-wrapper -->

	</div>
	<!-- /#wrapper -->
{{ template "footer" .}}
{{ end }}