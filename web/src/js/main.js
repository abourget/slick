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
    $scope.show_all = false;

    $scope.send_notif = function() {
        $http.post('/send_notif');
    };

    $scope.send_message = function() {
        $http.post('/send_message', {room: $scope.room, message: $scope.message});
        $scope.room = '';
        $scope.message = '';
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

.directive("deployer", function($timeout) {
    function setupWebsocket($scope) {
        function msg(msg) {
            $scope.$apply(function() {
                $scope.messages.push(msg);
            });
        }

        var ws = new WebSocket("ws://" + window.location.host + "/plugins/deployer.ws");
        ws.onopen = function() {
            msg("[Websocket connected]");
        };
        ws.onerror = function() {
            msg("[Websocket error]");
            $timeout(function() {
                msg("[Websocket reconnecting...]");
                setupWebsocket($scope);
            }, 3000);
        };
        ws.onmessage = function(event) {
            msg(event.data);
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

            if (ws.readyState == 3) {
                msg("[Websocket cannot send, resetting...]");
                setupWebsocket($scope);
                return;
            }
            ws.send(JSON.stringify({
                environment: $scope.env,
                branch: branch,
                tags: tags,
                initiatedBy: USER.name
            }));
        }
    }

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

            setupWebsocket($scope);
        }
    };
})
;
