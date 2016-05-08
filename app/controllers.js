/**
 * Created by hanchen on 4/20/16.
 */

myApp.controller('listController', ['$scope', '$http', '$routeParams', 'AuthService', 'Pass', 'PublicKey',
    function ($scope, $http, $routeParams, AuthService, Pass, PublicKey) {

        $scope.dirs = [];
        $scope.files = [];
        $scope.isDir = false;
        $scope.isFile = false;
        $scope.file = {};

        $scope.pathParts = [
            { name: 'root', path: '/' },
        ]

        var path = $routeParams.path;
        if (!path) {
            path = '.';
            $scope.isDir = true;
        } else {
            var pathParts = path.split('/');
            var pathStr = '';
            for (var i = 0; i < pathParts.length; i++) {
                pathStr += '/' + pathParts[i];
                $scope.pathParts.push({
                    name: pathParts[i],
                    path: pathStr,
                });
            }
        }

        Pass.get({ path: path }).$promise.then(function (data) {
            if (data.hasOwnProperty('children')) {
                // we have a directory
                $scope.isDir = true;
                $scope.isFile = false;
                for (var i = 0; i < data.children.length; i++) {
                    switch (data.children[i].type) {
                        case 'dir':
                            $scope.dirs.push(data.children[i]);
                            break;
                        case 'file':
                            $scope.files.push(data.children[i]);
                            break;
                    }
                }
            }
            else {
                $scope.isFile = true;
                $scope.isDir = false;
                $scope.file = data;

                var raw = atob(data.contents); // raw binary contents
                var buf = new Uint8Array(raw.length);
                for (var i = 0; i < raw.length; i++) {
                    buf[i] = raw.charCodeAt(i);
                }
                var message = openpgp.message.read(buf);
                $scope.message = message;
                $scope.recipients = [];

                var recipients = message.getEncryptionKeyIds();
                for (var i = 0; i < recipients.length; i++) {
                    var recipient = recipients[i].toHex().toUpperCase();
                    // trim beginning zeros
                    var zbegin = recipient.search(/[^0]/);
                    if (zbegin > 0) {
                        recipient = recipient.substr(zbegin);
                    }
                    $scope.recipients.push(recipient);
                    PublicKey.get({ keyId: recipient }).$promise.then(function (r) {
                        var idx = $scope.recipient.indexOf(recipient);
                        if (idx >= 0) {
                            $scope.recipients[idx] = r.user;
                        }
                    });
                }
            }
        });

        $scope.addDir = function () {
            // does something
        };

        $scope.addFile = function () {
            // does something
        };
    }]);

myApp.controller('userController', ['$scope', '$http', 'AuthService', 'User', 'Pass',
    function ($scope, $http, AuthService, User, Pass) {

        $scope.user = User.me();

        $scope.addKey = function () {
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
angular.module('myApp').filter('arrayToString', function () {
    return function (array) {
        return _.join(array, ', ');
    }
});
