"use strict";
(function() {
    var apiLocation = "http://localhost:8080/api";
    angular.module("myApp")
        .factory("User", ["$resource", function ($resource) {
            var User = $resource(apiLocation+"/user/:id", null, {
                'update': { method: 'PATCH' },
                'me': { method: 'GET', url: apiLocation+'/me' }
            });
            
            return User;
        }]);
})();