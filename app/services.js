angular.module('myApp').factory('AuthService',
    ['$q', '$timeout', '$http',
        function ($q, $timeout, $http) {

            function isLoggedIn() {
                console.log("isloggedin?");
                console.log(user);
                if(user) {
                    return true;
                } else {
                    return false;
                }
            }
            
            function getUserStatus() {
                return $http.get('/user/status')
                    // handle success
                    .success(function (data) {
                        if(data.status){
                            user = true;
                        } else {
                            user = false;
                        }
                    })
                    // handle error
                    .error(function (data) {
                        user = false;
                    });
            }
            
            function getUserId(){
                return $http.get('/user/id')
            }

            function getUserInfo(){
                return $http.get('/user/info')
            }

            function login(username, password) {

                // create a new instance of deferred
                var deferred = $q.defer();

                // send a post request to the server
                $http.post('/login',
                    {username: username, password: password})
                    // handle success
                    .success(function (data, status) {
                        if(status === 200){
                            user = true;
                            deferred.resolve();
                        } else {
                            user = false;
                            deferred.reject();
                        }
                    })
                    // handle error
                    .error(function (data) {
                        user = false;
                        deferred.reject();
                    });

                // return promise object
                return deferred.promise;
            }

            function logout() {

                // create a new instance of deferred
                var deferred = $q.defer();

                // send a get request to the server
                $http.get('/user/logout')
                    // handle success
                    .success(function (data) {
                        user = false;
                        deferred.resolve();
                    })
                    // handle error
                    .error(function (data) {
                        user = false;
                        deferred.reject();
                    });

                // return promise object
                return deferred.promise;
            }

            function register(fullname, username, password) {

                // create a new instance of deferred
                var deferred = $q.defer();

                // send a post request to the server
                $http.post('/user/register',
                    {fullname: fullname, username: username, password: password})
                    // handle success
                    .success(function (data, status) {
                        if(status === 200 && data.status){
                            deferred.resolve();
                        } else {
                            deferred.reject();
                        }
                    })
                    // handle error
                    .error(function (data) {
                        deferred.reject();
                    });

                // return promise object
                return deferred.promise;
            }

            function edit(userData) {
                console.log('server');
                console.log(userData);
                return $http.put('/user', userData, {
                    headers: {'Content-Type': 'application/json'}
                });
            }

            // create user variable
            var user = null;

            // return available functions for use in the controllers
            return ({
                isLoggedIn: isLoggedIn,
                getUserId: getUserId,
                getUserInfo: getUserInfo,
                login: login,
                logout: logout,
                register: register,
                edit: edit
            });

        }]);