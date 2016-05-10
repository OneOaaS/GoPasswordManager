/**
 * Created by hanchen on 4/20/16.
 */

myApp.controller('listController', ['$scope', '$http', '$routeParams', 'AuthService', 'Pass', 'PublicKey', 'User',
    function ($scope, $http, $routeParams, AuthService, Pass, PublicKey, User) {

        $scope.dirs = [];
        $scope.files = [];
        $scope.isDir = false;
        $scope.isFile = false;
        $scope.file = {};
        $scope.user = User.me();
        $scope.contents = '';

        $scope.fileForm = {};

        $scope.pathParts = [
            { name: 'root', path: '/' },
        ]

        $scope.loadPath = function () {
            // set defaults
            $scope.dirs = [];
            $scope.files = [];
            $scope.isDir = false;
            $scope.isFile = false;
            $scope.file = {};
            $scope.contents = '';

            $scope.pathParts = [
                { name: 'root', path: '/' },
            ]

            var path = $routeParams.path;
            if (!path) {
                path = '.';
                $scope.isDir = true;
                $scope.path = '/';
                $scope.fileForm.path = path;
            } else {
                var pathParts = path.replace(/\/+/g, '/').split('/');
                var pathStr = '';
                for (var i = 0; i < pathParts.length; i++) {
                    pathStr += '/' + pathParts[i];
                    $scope.pathParts.push({
                        name: pathParts[i],
                        path: pathStr,
                    });
                }
                $scope.path = pathStr;
                $scope.fileForm.path = pathStr;
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
                    $scope.file = data;
                $scope.user.getPrivateKeyIds().then(function (myKeys) {
                    for (var i = 0; i < data.recipients.length; i++) {
                        if (data.recipients[i] in myKeys) {
                            $scope.havePermissions = true;
                            break;
                        }
                    }
                });
                }
                else {
                    $scope.isFile = true;
                    $scope.isDir = false;
                    $scope.file = data;
                    $scope.contents = '';
                }
            });
        }

        // load path
        $scope.loadPath();

        $scope.addFile = function () {
            var keys = $scope.file.recipients.join(",");
            PublicKey.get({ ids: keys }).$promise.then(function (keys) {
                var newKeys = [],
                    k = Object.keys(keys),
                    dkeys = [];
                // transform map
                for (var i = 0; i < k.length; i++) {
                    if (k[i].indexOf("$") !== 0) { // skip angular info
                        newKeys.push(keys[k[i]]);
                    }
                }

                // decode keys
                for (var i = 0; i < newKeys.length; i++) {
                    var ikeys = openpgp.key.readArmored(atob(newKeys[i].armored)).keys;
                    if (ikeys.length > 0) {
                        dkeys.push(ikeys[0]);
                    }
                }

                var options = {
                    data: $scope.fileForm.password,
                    publicKeys: dkeys,
                    armor: false
                }

                openpgp.encrypt(options).then(function (message) {
                    var data = btoa(String.fromCharCode.apply(null, message.message.packets.write()));
                    var path = $scope.fileForm.path + '/' + $scope.fileForm.name + '.gpg';
                    path = path.replace(/\/+/g, '/'); // remove repeated slashes

                    var pass = new Pass({ path: path, contents: data, message: 'commit from web frontend' });
                    pass.$save().then(function () {
                        alert('Success!');
                        var idx = _.findIndex($scope.files, function (file) { return decodeURIComponent(file.name) === $scope.fileForm.name; });
                        if (idx < 0) {
                            $scope.files.push({
                                name: $scope.fileForm.name,
                                path: path,
                                type: 'file'
                            });
                        }
                        $scope.fileForm = { path: $scope.path }; // clear contents
                    }, function () {
                        alert('Fail!');
                    });
                });
            })
        };

        $scope.decryptFile = function () {
            if (!$scope.file || !$scope.file.contents) {
                return;
            }

            var msg = openpgp.message.read(b64ToU8($scope.file.contents));
            $scope.user.getPrivateKeyIds().then(function (keys) {
            var msgKeys = msg.getEncryptionKeyIds();
            var key = null;
            for (var i = 0; i < msgKeys.length; i++) {
                var id = msgKeys[i].toHex().toUpperCase();
                if (id in keys) {
                    key = keys[id];
                    break;
                }
            }
            if (!key) {
                alert('cannot decrypt this file: not permitted');
                return;
            }

            var passphrase = prompt('passphrase for key ' + key.primaryKey.getKeyId().toHex().toUpperCase());
            if (!key.decrypt(passphrase)) {
                alert('failed to decrypt key: invalid password');
                return;
            }

            var options = {
                message: msg,
                    privateKey: key
                };

            openpgp.decrypt(options).then(function (plaintext) {
                $scope.contents = plaintext.data;
                $scope.$apply(); // force update?
                });
            });
        };

        function b64ToU8(str) {
            var raw = atob(str); // raw binary contents
            var buf = new Uint8Array(raw.length);
            for (var i = 0; i < raw.length; i++) {
                buf[i] = raw.charCodeAt(i);
            }
            return buf;
        }
    }]);

myApp.controller('userController', ['$scope', '$q', '$http', 'AuthService', 'User', 'Pass', 'Reader', 'UserPrivateKey', 'UserPublicKey',
    function ($scope, $q, $http, AuthService, User, Pass, Reader, UserPrivateKey, UserPublicKey) {
        $scope.user = User.me();
        $scope.keyForm = {};
        $scope.editKeyForm = {};
        $scope.selectedKeyId = null;

        $scope.addKey = function () {
            if (!$scope.keyForm.key) {
                return;
            }

            Reader.readFile($scope.keyForm.key).then(function (data) {
                var result = openpgp.key.readArmored(data);
                if (!result.keys || result.keys.length < 1) {
                    // TODO: display error message
                    return;
                }

                var key = result.keys[0];
                if (!key.isPrivate()) {
                    // TODO: display error message
                    return;
                }

                // upload private key
                var privk = new UserPrivateKey({ userId: $scope.user.id, body: data });
                var pubk = new UserPublicKey({ userId: $scope.user.id, body: key.toPublic().armor() });
                var promises = $q.all([privk.$save(), pubk.$save()]);
                promises.then(function () {
                    $scope.user = User.me();
                });
            });
        };

        $scope.selectFile = function () {
            if ($scope.keyForm.key) {
                $scope.keyForm.keyFileName = $scope.keyForm.key.name;
            }
        };

        $scope.editSelectKey = function (key) {
            $scope.selectedKeyId = key.key;
        };

        $scope.editSelectFile = function () {
            if ($scope.editKeyForm.key) {
                $scope.editKeyForm.keyFileName = $scope.editKeyForm.key.name;
            }
        };

        $scope.editKey = function () {
            if (!$scope.editKeyForm.key || !$scope.selectedKeyId) {
                return;
            }

            Reader.readFile($scope.editKeyForm.key).then(function (data) {
                UserPrivateKey.update({ userId: $scope.user.id, keyId: $scope.selectedKeyId }, { body: data })
                    .$promise.then(function () {
                        $scope.user = User.me();
                    });
            });
        };

        $scope.deleteKey = function (key) {
            if (!confirm("Are you sure?")) {
                return;
            }

            var promises = [
                UserPublicKey.delete({ userId: $scope.user.id, keyId: key.key }).$promise,
                UserPrivateKey.delete({ userId: $scope.user.id, keyId: key.key }).$promise
            ];

            $q.all(promises).then(function () {
                $scope.user = User.me();
            });
        };
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

angular.module('myApp').filter('decodeUri', function () {
    return function (input) {
        return decodeURIComponent(input);
    };
});
