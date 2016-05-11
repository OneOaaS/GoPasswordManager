/**
 * Created by hanchen on 4/20/16.
 */

myApp.controller('listController', ['$scope', '$http', '$q', '$routeParams', '$route', 'AuthService', 'Pass', 'PublicKey', 'User', 'PassPerm',
    function ($scope, $http, $q, $routeParams, $route, AuthService, Pass, PublicKey, User, PassPerm) {

        $scope.dirs = [];
        $scope.files = [];
        $scope.isDir = false;
        $scope.isFile = false;
        $scope.file = {};
        $scope.user = User.me();
        $scope.contents = '';
        $scope.permissionKey = null;

        $scope.fileForm = {};
        $scope.permissionForm = {};

        buildPath();
        loadPath();

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

                encryptMessage($scope.fileForm.password, dkeys).then(function (data) {
                    var path = $scope.fileForm.path + '/' + $scope.fileForm.name + '.gpg';
                    path = path.replace(/\/+/g, '/'); // remove repeated slashes

                    var pass = new Pass({ path: path, contents: data, message: 'commit from web frontend' });
                    pass.$save().then(function () {
                        alert('Success!');
                        loadPath(); // just reload the path
                        $scope.fileForm = { path: $scope.path }; // clear contents
                    }, function () {
                        alert('Fail!');
                    });
                });
            })
        };

        $scope.deleteFile = function (file) {
            if (!confirm('Are you sure?')) {
                return;
            }

            Pass.delete({ path: decodeURIComponent(file.path) }).$promise.then(function () {
                // just reload the view
                loadPath().then(null, function (err) {
                    if (err.status && err.status === 404) {
                        // deleting file cleared folder, so redirect up a level
                        if ($scope.pathParts.length > 1) {
                            $route.updateParams({ path: $scope.pathParts[$scope.pathParts.length - 2].path });
                        }
                    }
                });
            });
        }

        $scope.decryptFile = function () {
            if (!$scope.file || !$scope.file.contents || !$scope.permissionKey) {
                return;
            }

            if (!decryptPermissionKey()) {
                alert('invalid password');
                return;
            }

            decryptMessage($scope.file.contents, $scope.permissionKey).then(function (plaintext) {
                $scope.contents = plaintext;
                $scope.$apply(); // force update?
            });
        };

        $scope.addPermission = function () {
            if (!$scope.permissionForm.keyId || !$scope.permissionKey) {
                return;
            }

            if (!decryptPermissionKey()) {
                alert('invalid password');
                return;
            }

            var newKeyId = $scope.permissionForm.keyId;
            PassPerm.get({ path: $scope.path }).$promise.then(function (perms) {
                if (perms.access.indexOf(newKeyId) >= 0) {
                    // already have permissions
                    return;
                }
                var access = perms.access;
                access.push(newKeyId);
                reencrypt(perms.change, access, $scope.permissionKey).then(function () {
                    alert('Success!');
                    loadPath(); // refresh the path
                    $scope.permissionForm = {};
                }, function () {
                    alert('Fail!');
                });
            });
        };

        $scope.deletePermission = function (recipient) {
            if (!$scope.permissionKey) {
                return;
            }

            if (!confirm('Are you sure?')) {
                return;
            }

            PassPerm.get({ path: $scope.path }).$promise.then(function (perms) {
                var idx = perms.access.indexOf(recipient);
                if (idx < 0) {
                    alert("recipient doesn't exist");
                    return;
                }
                var access = perms.access;
                access.splice(idx, 1);
                if (access.length == 0) {
                    alert('cannot remove last key');
                    return;
                }
                if (!decryptPermissionKey()) {
                    alert('invalid password');
                    return;
                }
                reencrypt(perms.change, access, $scope.permissionKey).then(function () {
                    alert('Success!');
                    loadPath(); // reload the path
                }, function () {
                    alert('Fail');
                });
            });
        };

        function buildPath() {
            $scope.pathParts = [
                { name: 'root', path: '/' },
            ]

            var path = $routeParams.path || '/';
            path.replace(/\/+/g, '/'); // clean path

            if (path !== '/') {
                if (path[0] === '/') path = path.substr(1);
                var pathParts = path.split('/');
                for (var i = 0; i < pathParts.length; i++) {
                    $scope.pathParts.push({
                        name: pathParts[i],
                        path: '/' + _.take(pathParts, i + 1).join('/')
                    });
                }
            }

            $scope.path = path;
            $scope.fileForm.path = path;
        }

        function loadPath() {
            return Pass.get({ path: $scope.path }).$promise.then(function (data) {
                if (data.hasOwnProperty('children')) {
                    // we have a directory
                    $scope.isDir = true;
                    $scope.isFile = false;
                    $scope.files = [];
                    $scope.dirs = [];
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
                }
                else {
                    $scope.isFile = true;
                    $scope.isDir = false;
                    $scope.file = data;
                    $scope.contents = '';
                }

                $scope.user.getPrivateKeyIds().then(function (myKeys) {
                    for (var i = 0; i < data.recipients.length; i++) {
                        if (data.recipients[i] in myKeys) {
                            $scope.permissionKey = myKeys[data.recipients[i]];
                            break;
                        }
                    }
                });
            });
        }

        function decryptPermissionKey() {
            if (!$scope.permissionKey.getEncryptionKeyPacket().isDecrypted) {
                var passphrase = prompt('enter passphrase for key id ' + $scope.permissionKey.primaryKey.getKeyId().toHex().toUpperCase());
                return $scope.permissionKey.decrypt(passphrase);
            }
            return true;
        }

        function reencrypt(files, pubKeys, privKey) {
            return PublicKey.get({ ids: pubKeys.join(',') }).$promise.then(function (keys) {
                for (var i = 0; i < pubKeys; i++) {
                    if (!(pubKeys[i] in keys)) {
                        return $q.reject("could not find all keys");
                    }
                }
                var keyArr = [];
                for (var k in keys) {
                    if (keys.hasOwnProperty(k) && k[0] !== '$') {
                        var rKeys = openpgp.key.readArmored(atob(keys[k].armored)).keys || [];
                        Array.prototype.push.apply(keyArr, rKeys);
                    }
                }
                var promises = [];
                for (var i = 0; i < files.length; i++) {
                    var path = files[i];
                    (function (path) {
                        promises.push(Pass.get({ path: decodeURIComponent(path) }).$promise.then(function (pass) {
                            console.log('reencrypting ' + path);
                            return reencryptMessage(pass.contents, privKey, keyArr).then(function (contents) {
                                var obj = {};
                                obj[path] = contents;
                                return obj;
                            });
                        }));
                    })(path);
                }
                return $q.all(promises).then(function (results) {
                    var merged = {};
                    for (var i = 0; i < results.length; i++) {
                        angular.merge(merged, results[i]);
                    }
                    var path = $scope.path;
                    if (path === '/') path = '.';
                    return PassPerm.save({ path: path }, { access: pubKeys, files: merged }).$promise;
                })
            });
        }

        function getKeyFromMessage(msg, keys) {
            var msgKeys = msg.getEncryptionKeyIds();
            for (var i = 0; i < msgKeys.length; i++) {
                var id = msgKeys[i].toHex().toUpperCase();
                if (id in keys) {
                    return keys[id];
                }
            }
            return null;
        }

        function contentsToMessage(contents) {
            return openpgp.message.read(b64ToU8(contents));
        }

        function reencryptMessage(contents, key, newKeys) {
            return decryptMessage(contents, key).then(function (message) {
                return encryptMessage(message, newKeys);
            });
        }

        function decryptMessage(contents, key) {
            if (!(contents instanceof openpgp.message.Message)) {
                contents = contentsToMessage(contents);
            }
            var options = {
                message: contents,
                privateKey: key
            };
            return openpgp.decrypt(options).then(function (plaintext) {
                return plaintext.data;
            });
        }

        function encryptMessage(data, keys) {
            if (!angular.isArray(keys)) {
                keys = [keys];
            }
            var options = {
                data: data,
                publicKeys: keys,
                armor: false
            };
            return openpgp.encrypt(options).then(function (message) {
                return btoa(String.fromCharCode.apply(null, message.message.packets.write()));
            });
        }

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
