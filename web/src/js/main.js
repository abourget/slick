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

    $scope.tabularasa = function() {
        if (confirm("Are you SURE you want to Tabula Rasa Asana tasks for everyone ?")) {
            $http.post('/plugins/tabularasa');
        }
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

})

.directive("deployer", function() {
    return {
        restrict: "E",
        templateUrl: "deployer.html",
        link: function($scope, $element, $attrs) {
            $scope.messages = [];
            $scope.branch = '';
            $scope.env = 'stage';
            $scope.clear_on_deploy = true;
            $scope.tag_updt = true;
            $scope.advanced = false;
            $scope.custom_tags = '';

            var ws = new WebSocket("ws://" + window.location.host + "/plugins/deployer.ws");
            ws.onopen = function(ev) {
                $scope.$apply(function() {
                    $scope.messages.push("[Websocket connected]");
                });
            };
            ws.onerror = function(ev) {
                $scope.$apply(function() {
                    console.log(ev);
                    $scope.messages.push("[Websocket error]");
                });
            };
            ws.onmessage = function(event) {
                $scope.$apply(function() {
                    $scope.messages.push(event.data);
                });
            }
            $scope.deploy = function() {
                var branch = $scope.branch;
                if ($scope.env == 'prod') {
                    branch = '';
                }

                if ($scope.clear_on_deploy) {
                    $scope.messages = [];
                }

                if (!$scope.tag_updt) {
                    alert("Hmm.. thought we'd have 'updt_streambed' by default ?");
                    return;
                }
                var tags = "updt_streambed";
                if ($scope.tag_config) {
                    tags += ",global_config,restart_streambed";
                }
                if ($scope.custom_tags) {
                    tags += "," + $scope.custom_tags;
                }

                ws.send(JSON.stringify({
                    environment: $scope.env,
                    branch: branch,
                    tags: tags,
                    initiatedBy: USER.name
                }));
            }
        }
    };
})
;
