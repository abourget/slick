'use strict';

angular.module('hey', ['ui.router.state', 'ui.router'])

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
    templateUrl: '/static/tpl/home.html'
  });

})

.controller('HomeCtrl', function($scope, $http) {
  $scope.boo = 'thanks';
  $scope.send_notif = function() {
    $http.post('/send_notif');
  };
})

;
