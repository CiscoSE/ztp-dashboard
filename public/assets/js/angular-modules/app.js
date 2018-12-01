/* Global variables */

var appModule = angular.module('appModule',['ngRoute','ngAnimate'])

/*  Filters    */

// Tells if an object is instance of an array type. Used primary within ng-templates
appModule.filter('isArray', function() {
  return function (input) {
    return angular.isArray(input);
  };
});


// Add new item to list checking first if it has not being loaded and if it is not null.
// Used primary within ng-templates
appModule.filter('append', function() {
  return function (input, item) {
    if (item){
        for (i = 0; i < input.length; i++) {
            if(input[i] === item){
                return input;
            }
        }
        input.push(item);
    }
    return input;
  };
});

// Remove item from list. Used primary within ng-templates
appModule.filter('remove', function() {
  return function (input, item) {
    input.splice(input.indexOf(item),1);
    return input;
  };
});

// Capitalize the first letter of a word
appModule.filter('capitalize', function() {

  return function(token) {
      return token.charAt(0).toUpperCase() + token.slice(1);
   }
});

// Replace any especial character for a space
appModule.filter('removeSpecialCharacters', function() {

  return function(token) {
      return token.replace(/#|_|-|$|!|\*/g,' ').trim();
   }
});


function property(){
    function parseString(input){
        return input.split(".");
    }

    function getValue(element, propertyArray){
        var value = element;

        _.forEach(propertyArray, function(property){
            value = value[property];
        });

        return value;
    }

    return function (array, propertyString, target){
        var properties = parseString(propertyString);

        return _.filter(array, function(item){
            return getValue(item, properties).toUpperCase().startsWith(target.toUpperCase());
        });
    }
}

appModule.filter('property', property);

appModule.filter('int', function() {
    return function(input) {
       return parseInt(input, 10);
    }
});


/*  Configuration    */

// Application routing
appModule.config(function($routeProvider, $locationProvider){
    // Maps the URLs to the templates located in the server
    $routeProvider
        .when('/', {templateUrl: '/ng/home'})
        .when('/home', {templateUrl: '/ng/home'})
        .when('/devices', {templateUrl: '/ng/devices'})
        .when('/devices/detail', {templateUrl: '/ng/devices/detail'})
        .when('/settings', {templateUrl: '/ng/settings'})
    $locationProvider.html5Mode(true);
});


// To avoid conflicts with other template tools such as Jinja2, all between {a a} will be managed by ansible instead of {{ }}
appModule.config(['$interpolateProvider', function($interpolateProvider) {
  $interpolateProvider.startSymbol('{a');
  $interpolateProvider.endSymbol('a}');
}]);

/* Factories */

// The notify factory allows services to notify to an specific controller when they finish operations
appModule.factory('NotifyingService' ,function($rootScope) {
    return {
        subscribe: function(scope, event_name, callback) {
            var handler = $rootScope.$on(event_name, callback);
            scope.$on('$destroy', handler);
        },

        notify: function(event_name) {
            $rootScope.$emit(event_name);
        }
    };
});

/*  Controllers    */

// App controller is in charge of managing all services for the application
appModule.controller('AppController', function($scope, $location, $http){

    $scope.error = "";
    $scope.success = "";
    $scope.loading = false;
    
    $scope.clearError = function(){
        $scope.error = "";
    };
    $scope.clearSuccess = function(){
        $scope.success = "";
    };
    $scope.go = function (path) {
        $location.path(path);
    };
});
