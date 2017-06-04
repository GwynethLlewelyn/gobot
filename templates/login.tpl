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
								<div class="checkbox">
									<label>
										<input name="remember" type="checkbox" value="Remember Me">Remember Me
									</label>
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
	<script type="text/javascript">
		// see https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/HTML5/Constraint_validation
		function checkInput() {
			var email		= document.getElementByName("email").value,
				password	= document.getElementByName("password").value;
				
			console.log("Email status: " . email.CheckValidity("Invalid Email"));
			console.log("Password status: " . password.CheckValidity("Invalid Password"));
		}
		
		window.onload = function () {
			document.getElementByName("email").oninput = checkInput;
			document.getElementByName("password").oninput = checkInput;
		}
	</script>
{{ template "footer" .}}
{{ end }}