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
							<form role="form">
							<div class="row">
								<div class="col-lg-6">
									
										<div class="form-group">
											<label>Enter your command PermURL</label>
											<select class="form-control" name="PermURL" id="PermURL" size="1">
												<option value="0" selected="selected" disabled="disabled">Please choose an avatar</option>
												<option>2</option>
												<option>3</option>
												<option>4</option>
												<option>5</option>
											</select>
										
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
										
										<div class="form-group">
											<label>Selects</label>
											<select multiple class="form-control">
												<option>1</option>
												<option>2</option>
												<option>3</option>
												<option>4</option>
												<option>5</option>
											</select>
										</div> <!-- /.form-group -->
										<button type="submit" class="btn btn-default">Submit Button</button>
										<button type="reset" class="btn btn-default">Reset Button</button>
									</form>
								</div>
								<!-- /.col-lg-6 (nested) -->
								<div class="col-lg-6">
									<h1>More Form States</h1>
									<form role="form">
										 <div class="form-group">
											<label>Selects</label>
											<select multiple class="form-control">
												<option>1</option>
												<option>2</option>
												<option>3</option>
												<option>4</option>
												<option>5</option>
											</select>
										</div> <!-- /.form-group -->									
									</form>
								</div>
								<!-- /.col-lg-6 (nested) -->
							</div>
							<!-- /.row (nested) -->
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