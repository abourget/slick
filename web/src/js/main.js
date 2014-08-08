'use strict';

angular.module('plotbot', ['ui.router.state', 'ui.router'])

.run(function($rootScope, $state, $stateParams) {
    $rootScope.hello = 'world';
})

.config(function($stateProvider, $urlRouterProvider) {
    $urlRouterProvider.otherwise('/');
})

.config(function($locationProvider) {
    $locationProvider.html5Mode(true);
})

.config(function($stateProvider) {
    $stateProvider.state('home', {
        url: '/',
        controller: 'HomeCtrl',
        templateUrl: 'home.html'
    });

})

.controller('HomeCtrl', function($scope, $http) {
    $scope.send_notif = function() {
        $http.post('/send_notif');
    };

    $scope.get_standup = function() {
        $http.get('/plugins/standup.json').success(function(data, status) {
          $scope.standup = data;
        });
    };

    $scope.load_users = function() {
        $http.get('/hipchat/users').success(function(data, status) {
            $scope.users = data.users;
        });
    };

    $scope.load_rooms = function() {
        $http.get('/hipchat/rooms').success(function(data, status) {
            $scope.rooms = data.rooms;
        });
    };

});
