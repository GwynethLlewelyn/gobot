{{ define "login" }}
{{ template "header" .}}
	<div class="container">
		<div class="row">
			<div class="col-md-4 col-md-offset-4">
				<div class="login-panel panel panel-default">
					<div class="panel-heading">
						<h3 class="panel-title">Please Sign In</h3>
					</div>
					<div class="panel-body">
						<form role="form" action="{{.URLPathPrefix}}/admin/login/" method="post">
							<fieldset>
								<div class="form-group input-group">
									<span class="input-group-addon">@</span>
									<input class="form-control" placeholder="E-mail" name="email" type="email" autofocus required>
								</div>
								<div class="form-group input-group">
									<span class="input-group-addon">
									 	<i class="fa fa-lock"></i>
									</span>
									<input class="form-control" placeholder="Password" name="password" type="password" value="" required>
								</div>
								<!-- Change this to a button or input when using this as a form -->
								<input type="submit" value="Login" class="btn btn-lg btn-success btn-block">
							</fieldset>
						</form>
					</div>
				</div>
			</div>
		</div>
	</div>
{{ template "footer" .}}
{{ end }}