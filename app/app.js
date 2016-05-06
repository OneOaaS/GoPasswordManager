var myApp = angular.module('myApp', ['ngRoute', 'angular.filter', 'ngResource', 'ui.tree']).constant('_', window._);

myApp.config(function ($routeProvider) {
    $routeProvider
        .when('/login', {
            templateUrl: '/partials/login.html',
            controller: 'loginController',
            access: {restricted: false}
        })
        .when('/logout', {
            controller: 'logoutController',
            access: {restricted: true}
        })
        .when('/register', {
            templateUrl: '/partials/register.html',
            controller: 'registerController',
            access: {restricted: false}
        })
        .when('/user', {
            templateUrl: '/partials/user.html',
            controller: 'userController',
            access: {restricted: true}
        })
        .when('/list', {
            templateUrl: '/partials/list.html',
            controller: 'listController',
            access: {restricted: true}
        })
        .otherwise({
            redirectTo: '/login'
        });
});

// myApp.run(function ($rootScope, $location, $route, AuthService) {
//     $rootScope.$on('$routeChangeStart',
//         function (event, next, current) {
//             AuthService.getUserStatus()
//                 .then(function(){
//                     if (next.access.restricted && !AuthService.isLoggedIn()){
//                         $location.path('/');
//                         $route.reload();
//                     }
//                 });
//         });
// });