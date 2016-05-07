/**
 * Created by hanchen on 4/20/16.
 */

myApp.controller('listController', ['$scope', '$http', '$routeParams', 'AuthService',
    function ($scope, $http, $routeParams, AuthService) {

        // HARDCODED DIR PLACEHOLDER
        $scope.isDir = true;
        $scope.dirs = [
            {name: 'bob', path: '/watermelon/bob'},
            {name: 'george', path: '/doodle/george.gdg'}
        ];

        // HARDCODED FILE PLACEHOLDER
        $scope.isFile = true;
        $scope.key = "fjl34rjkargajeio;4tja9wegua4gh4htqh4ulhsie;4jgaoi;34jt34tq34;igj43inwglrhj;5jgiapihr4gjls4g";

        $scope.addDir = function(){
            // does something
        };

        $scope.addFile = function(){
            // does something
        };
    }]);

myApp.controller('userController', ['$scope', '$http', 'AuthService', 'User', 'Pass',
    function ($scope, $http, AuthService, User, Pass) {

        $scope.user = User.me();
        $scope.passwords = Pass.get({path:'.'});
        $scope.passwords.$promise.then(function() {
            $scope.passwords['name'] = '/';
        });
        
        $scope.addKey = function(){
            $scope.user.keys.push($scope.keyForm.key);
        }
    }]);

// angular.module('myApp').controller('userController', ['$scope'],)

angular.module('myApp').controller('loginController',
    ['$scope', '$location', 'AuthService',
        function ($scope, $location, AuthService) {

            $scope.login = function () {

                // initial values
                $scope.error = false;
                $scope.disabled = true;

                // call login from service
                AuthService.login($scope.loginForm.username, $scope.loginForm.password)
                    // handle success
                    .then(function () {
                        $location.path('/user');
                        $scope.disabled = false;
                        $scope.loginForm = {};
                    })
                    // handle error
                    .catch(function () {
                        $scope.error = true;
                        $scope.errorMessage = "Invalid username and/or password";
                        $scope.disabled = false;
                        $scope.loginForm = {};
                    });

            };

        }]);

angular.module('myApp').controller('logoutController',
    ['$scope', '$location', 'AuthService', 'User',
        function ($scope, $location, AuthService, User) {

            $scope.isLoggedIn = AuthService.isLoggedIn();

            $scope.logout = function () {
                console.log("hello logout");
                // call logout from service
                AuthService.logout()
                    .then(function () {
                        $location.path('/login');
                    });

            };

        }]);


// Filter array to be strings for display
angular.module('myApp').filter('arrayToString', function() {
    return function(array){
        return _.join(array, ', ');
    }
});
