/// <reference path="typings/tsd.d.ts" />
(function () {
    "use strict";

    var apiLocation = "http://localhost:8080/api";

    angular.module("myApp")
        .factory("User", ["$q", "$resource", "UserPublicKey", "UserPrivateKey", UserService])
        .factory("UserPublicKey", ["$resource", UserPublicKeyService])
        .factory("UserPrivateKey", ["$resource", UserPrivateKeyService])
        .factory("PublicKey", ["$resource", PublicKey]);

    function UserService($q, $resource, UserPublicKey, UserPrivateKey) {
        var User = $resource(apiLocation + "/user/:userId", null, {
            'update': { method: 'PATCH' },
            'me': { method: 'GET', url: apiLocation + '/me' }
        });
        angular.extend(User.prototype, {
            getPublicKeys: function () {
                var deferred = $q.defer();
                this.$promise.then(function (user) {
                    deferred.resolve(UserPublicKey.query({ userId: user.id }));
                });
                return deferred.promise;
            },
            getPublicKey: function (id) {
                var deferred = $q.defer();
                this.$promise.then(function (user) {
                    deferred.resolve(UserPublicKey.query({ userId: user.id, keyId: id }));
                });
                return deferred.promise;
            },
            getPrivateKeys: function () {
                var deferred = $q.defer();
                this.$promise.then(function (user) {
                    deferred.resolve(UserPrivateKey.query({ userId: user.id }));
                });
                return deferred.promise;
            },
            getPrivateKey: function (id) {
                var deferred = $q.defer();
                this.$promise.then(function (user) {
                    deferred.resolve(UserPrivateKey.query({ userId: user.id, keyId: id }));
                });
                return deferred.promise;
            }
        });
        return User;
    }

    function UserPublicKeyService($resource) {
        var UserPublicKey = $resource(apiLocation + "/user/:userId/publicKey/:keyId");
        return UserPublicKey;
    }

    function UserPrivateKeyService($resource) {
        var UserPrivateKey = $resource(apiLocation + "/user/:userId/privateKey/:keyId", null, {
            'update': { method: 'PUT' }
        });
        return UserPrivateKey;
    }

    function PublicKey($resource) {
        var PublicKey = $resource(apiLocation + "/publicKey/:keyId");
        return PublicKey;
    }
})();