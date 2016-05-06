/// <reference path="typings/tsd.d.ts" />
(function () {
    "use strict";

    var apiLocation = "http://localhost:8080/api";

    angular.module("myApp")
        .factory("User", ["$resource", "UserPublicKey", "UserPrivateKey", UserService])
        .factory("UserPublicKey", ["$resource", UserPublicKeyService])
        .factory("UserPrivateKey", ["$resource", UserPrivateKeyService])
        .factory("PublicKey", ["$resource", PublicKey]);

    function UserService($resource, UserPublicKey, UserPrivateKey) {
        var User = $resource(apiLocation + "/user/:userId", null, {
            'update': { method: 'PATCH' },
            'me': { method: 'GET', url: apiLocation + '/me' }
        });
        angular.extend(User.prototype, {
            getPublicKeys: function () {
                return UserPublicKey.query({ userId: this.id });
            },
            getPublicKey: function (id) {
                return UserPublicKey.get({ userId: this.id, keyId: id });
            },
            getPrivateKeys: function () {
                return UserPrivateKey.query({ userId: this.id });
            },
            getPrivateKey: function (id) {
                return UserPrivateKey.get({ userId: this.id, keyId: id });
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