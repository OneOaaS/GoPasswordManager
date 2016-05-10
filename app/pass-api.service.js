/// <reference path="typings/tsd.d.ts" />
(function () {
    "use strict";

    var apiLocation = "http://localhost:8080/api";

    angular.module("myApp")
        .factory("User", ["$q", "$resource", "UserPublicKey", "UserPrivateKey", UserService])
        .factory("UserPublicKey", ["$resource", UserPublicKeyService])
        .factory("UserPrivateKey", ["$resource", UserPrivateKeyService])
        .factory("PublicKey", ["$resource", PublicKeyService])
        .factory("Pass", ["$resource", PassService])
        .config(["$httpProvider", PassConfig])
        .factory("PassPerm", ["$resource", PassPermService])
        .config(["$httpProvider", PassPermConfig])
        .factory("Reader", ["$q", FileReaderService]);

    function UserService($q, $resource, UserPublicKey, UserPrivateKey) {
        var User = $resource(apiLocation + "/user/:userId", null, {
            'update': { method: 'PATCH' },
            'me': { method: 'GET', url: apiLocation + '/me' }
        });
        angular.extend(User.prototype, {
            getPublicKeys: function () {
                return this.$promise.then(function (user) {
                    return UserPublicKey.query({ userId: user.id }).$promise.then(function (keys) {
                        return keys;
                    });
                });
            },
            getPublicKey: function (id) {
                return this.$promise.then(function (user) {
                    return UserPublicKey.get({ userId: user.id, keyId: id }).$promise.then(function (key) {
                        return key;
                    });
                });
            },
            getPrivateKeys: function () {
                return this.$promise.then(function (user) {
                    return UserPrivateKey.query({ userId: user.id }).$promise.then(function (keys) {
                        return keys;
                    });
                });
            },
            getPrivateKey: function (id) {
                return this.$promise.then(function (user) {
                    return UserPrivateKey.query({ userId: user.id, keyId: id }).$promise.then(function (key) {
                        return key;
                    });
                });
            },
            getPrivateKeyIds: function () {
                var keyMap = {};
                for (var i = 0; i < this.privateKeys.length; i++) {
                    var keys = openpgp.key.readArmored(atob(this.privateKeys[i].armored)).keys || [];
                    for (var j = 0; j < keys.length; j++) {
                        var ids = keys[j].getKeyIds();
                        for (var k = 0; k < ids.length; k++) {
                            var idStr = ids[k].toHex().toUpperCase();
                            keyMap[idStr] = keys[j];
                        }
                    }
                }
                return keyMap;
            }
        });
        return User;
    }

    function UserPublicKeyService($resource) {
        var UserPublicKey = $resource(apiLocation + "/user/:userId/publicKey/:keyId",
            { userId: '@userId', keyId: '@keyId' },
            {
                'save': {
                    method: 'POST',
                    transformRequest: function (data, headers) {
                        return data.body;
                    },
                    headers: { 'Content-Type': 'text/plain' }
                }
            });
        return UserPublicKey;
    }

    function UserPrivateKeyService($resource) {
        var UserPrivateKey = $resource(apiLocation + "/user/:userId/privateKey/:keyId",
            { userId: '@userId', keyId: '@keyId' },
            {
                'update': {
                    method: 'PUT',
                    transformRequest: function (data, headers) {
                        return data.body;
                    },
                    headers: { 'Content-Type': 'text/plain' }
                },
                'save': {
                    method: 'POST',
                    transformRequest: function (data, headers) {
                        return data.body;
                    },
                    headers: { 'Content-Type': 'text/plain' }
                }
            });
        return UserPrivateKey;
    }

    function PublicKeyService($resource) {
        var PublicKey = $resource(apiLocation + "/publicKey/:keyId");
        return PublicKey;
    }

    function PassService($resource) {
        var Pass = $resource(apiLocation + "/pass/:path", { path: '@path' });
        return Pass;
    }

    function PassConfig($httpProvider) {
        // awful hack to rewrite Pass urls and unescape the path
        $httpProvider.interceptors.push(function () {
            return {
                request: function (config) {
                    var pathPattern = "/api/pass/";

                    var uri = document.createElement("a"); // cheap URI parsing
                    uri.href = config.url;

                    if (uri.pathname.indexOf(pathPattern) !== 0) {
                        // not interested in this path
                        return config;
                    }

                    uri.pathname = uri.pathname.replace(/%2F/gi, "/");
                    config.url = uri.href;

                    return config;
                }
            };
        })
    }

    function PassPermService($resource) {
        var PassPerm = $resource(apiLocation + "/passPerm/:path");
        return PassPerm;
    }

    function PassPermConfig($httpProvider) {
        // awful hack to rewrite PassPerm urls and unescape the path
        $httpProvider.interceptors.push(function () {
            return {
                request: function (config) {
                    var pathPattern = "/api/passPerm/";

                    var uri = document.createElement("a"); // cheap URI parsing
                    uri.href = config.url;

                    if (uri.pathname.indexOf(pathPattern) !== 0) {
                        // not interested in this path
                        return config;
                    }

                    uri.pathname = uri.pathname.replace(/%2F/gi, "/");
                    config.url = uri.href;

                    return config;
                }
            };
        })
    }

    function FileReaderService($q) {
        var Reader = {};

        Reader.readFile = function (file) {
            var d = $q.defer();
            var fr = new FileReader();

            fr.onload = function (evt) {
                d.resolve(evt.target.result);
            };

            fr.onerror = function () {
                d.reject(this);
            }

            fr.readAsText(file);

            return d.promise;
        }

        return Reader;
    }
})();