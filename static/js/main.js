'use strict';

angular.module('hey', ['ui.router.state', 'ui.router', 'ngResource'])

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

.service( 'gobot', function ($resource) {
    return $resource('/', {}, {
        'notify': { method: 'post', url: '/send_notif'}
    });
})

.controller('HomeCtrl', function($scope, $http, gobot) {
    $scope.boo = 'thanks';
    $scope.send_notif = function() {
        gobot.notify();
    };
    $scope.send_storm = function() {
        gobot.storm();
    };

});
