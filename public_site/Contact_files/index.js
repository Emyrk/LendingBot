var app = angular.module('lendingApp', ['ngMaterial']);

app.controller('indexController', ['$scope', '$http', '$log',
	function($scope, $http, $log) {
		var indexScope = $scope;

		indexScope.login = function() {
			indexScope.loginError = '';
			indexScope.attemptingLogin = true;
			$http({
				method: 'POST',
				url: '/login',
				data: $.param({
					email: indexScope.login.email,
					pass: indexScope.login.pass,
					twofa: indexScope.login.twofa,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res, status, headers, config) => {
				//success
				$log.info("login: Success.");

				$('#hiddenLoginForm').submit();
				window.location = '/dashboard';
			}, (err, status, headers, config) => {
				//error
				$log.error("login: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				indexScope.loginError = err.data.error;
			})
			.then(() => {
				indexScope.attemptingLogin = false;
				indexScope.login.email = "";
				indexScope.login.pass = "";
				indexScope.login.twofa = "";
			});
		}

		indexScope.cancelLogin = function() {
			indexScope.attemptingLogin = false;
			indexScope.login.email = "";
			indexScope.login.pass = "";
			indexScope.login.twofa = "";
		}

		indexScope.register = function() {
			indexScope.registerError = "";
			indexScope.attemptingRegister = true;
			$http(
			{
				method: 'POST',
				url: '/register',
				data: $.param({
					email: indexScope.register.email,
					pass: indexScope.register.pass,
					code: indexScope.register.code,
				}),
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				withCredentials: true
			})
			.then((res) => {
				//success
				$log.info("register: Success.");
				$('#regPassword').val("");
				$('#regPassword2').val("");
				$('#registerForm').submit();
				window.location = '/dashboard';
			}, (err) => {
				//error
				$log.error("register: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				indexScope.registerError = err.data.error;
			})
			.then(() => {
				indexScope.register.pass = "";
				indexScope.register.pass2 = "";
				indexScope.attemptingRegister = false;
			});
		}

		indexScope.cancelRegister = function() {
			indexScope.attemptingRegister = false;
			indexScope.register.email = "";
			indexScope.register.pass = "";
			indexScope.register.pass2 = "";
		}

		indexScope.validAuth = function() {
			$http(
			{
				method: 'GET',
				url: '/validauth',
				data : {},
				withCredentials: true,
				headers: {
					'Cache-Control' : 'no-cache'
				}
			})
			.then((res) => {
				//success
				indexScope.isLoggedIn = res.data.data;
			}, (err) => {
				//error
				//$log.error("validAuth: Error: [" + JSON.stringify(err) + "] Status [" + err.status + "]");
				indexScope.isLoggedIn = false;
			});
		}

		//--init
		indexScope.attemptingLogin = false;
		indexScope.attemptingRegister = false;
		init_validator()
		indexScope.validAuth();
		//
	}]);



/* VALIDATOR */

function init_validator () {
	
	if( typeof (validator) === 'undefined'){ return; }
	console.log('init_validator'); 
	
	  // initialize the validator function
	  validator.message.date = 'not a real date';

      // validate a field on "blur" event, a 'select' on 'change' event & a '.reuired' classed multifield on 'keyup':
      $('.loginRegisterForm')
      .on('blur', 'input[required], input.optional, select.required', validator.checkField)
      .on('change', 'select.required', validator.checkField)
      .on('keypress', 'input[required][pattern]', validator.keypress);

      $('.multi.required').on('keyup blur', 'input', function() {
      	validator.checkField.apply($(this).siblings().last()[0]);
      });

      $('.loginRegisterForm').submit(function(e) {
      	e.preventDefault();
      	var submit = true;

        // evaluate the form using generic validaing
        if (!validator.checkAll($(this))) {
        	submit = false;
        }

        // if (submit)
        // 	this.submit();

        return false;
    });
      
  };