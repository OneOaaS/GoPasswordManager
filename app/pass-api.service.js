(function () {
    "use strict";

    var apiLocation = "http://localhost:8080/api";
    
    angular.module("myApp")
        .factory("User", ["$resource", UserService])
        .factory("UserPublicKey", ["$resource", UserPublicKeyService]);

    function UserService($resource) {
        var User = $resource(apiLocation + "/user/:id", null, {
            'update': { method: 'PATCH' },
            'me': { method: 'GET', url: apiLocation + '/me' }
        });

        return User;
    }
    
    function UserPublicKeyService($resource) {
        var UserPublicKey = $resource(apiLocation+"/user/:userId/publicKey/:keyId")
        return UserPublicKey;
    }
})();