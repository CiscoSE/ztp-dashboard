/* Global variables */

var appModule = angular.module('appModule', ['ngRoute', 'ngAnimate'])

/*  Filters    */

// Tells if an object is instance of an array type. Used primary within ng-templates
appModule.filter('isArray', function () {
    return function (input) {
        return angular.isArray(input);
    };
});


// Add new item to list checking first if it has not being loaded and if it is not null.
// Used primary within ng-templates
appModule.filter('append', function () {
    return function (input, item) {
        if (item) {
            for (i = 0; i < input.length; i++) {
                if (input[i] === item) {
                    return input;
                }
            }
            input.push(item);
        }
        return input;
    };
});

// Remove item from list. Used primary within ng-templates
appModule.filter('remove', function () {
    return function (input, item) {
        input.splice(input.indexOf(item), 1);
        return input;
    };
});

// Capitalize the first letter of a word
appModule.filter('capitalize', function () {

    return function (token) {
        return token.charAt(0).toUpperCase() + token.slice(1);
    }
});

// Replace any especial character for a space
appModule.filter('removeSpecialCharacters', function () {

    return function (token) {
        return token.replace(/#|_|-|$|!|\*/g, ' ').trim();
    }
});


function property() {
    function parseString(input) {
        return input.split(".");
    }

    function getValue(element, propertyArray) {
        var value = element;

        _.forEach(propertyArray, function (property) {
            value = value[property];
        });

        return value;
    }

    return function (array, propertyString, target) {
        var properties = parseString(propertyString);

        return _.filter(array, function (item) {
            return getValue(item, properties).toUpperCase().startsWith(target.toUpperCase());
        });
    }
}

appModule.filter('property', property);

appModule.filter('int', function () {
    return function (input) {
        return parseInt(input, 10);
    }
});


/*  Configuration    */

// Application routing
appModule.config(function ($routeProvider, $locationProvider) {
    // Maps the URLs to the templates located in the server
    $routeProvider
        .when('/', { templateUrl: '/ng/home' })
        .when('/home', { templateUrl: '/ng/home' })
        .when('/devices', { templateUrl: '/ng/devices' })
        .when('/devices/detail', { templateUrl: '/ng/devices/detail' })
        .when('/settings', { templateUrl: '/ng/settings' })
        .when('/configs', { templateUrl: '/ng/configs' })
        .when('/configs/detail', { templateUrl: '/ng/configs/detail' })
        .when('/images', { templateUrl: '/ng/images' })
        .when('/images/detail', { templateUrl: '/ng/images/detail' })
    $locationProvider.html5Mode(true);
});


// To avoid conflicts with other template tools such as Jinja2, all between {a a} will be managed by ansible instead of {{ }}
appModule.config(['$interpolateProvider', function ($interpolateProvider) {
    $interpolateProvider.startSymbol('{a');
    $interpolateProvider.endSymbol('a}');
}]);

/* Factories */

// The notify factory allows services to notify to an specific controller when they finish operations
appModule.factory('NotifyingService', function ($rootScope) {
    return {
        subscribe: function (scope, event_name, callback) {
            var handler = $rootScope.$on(event_name, callback);
            scope.$on('$destroy', handler);
        },

        notify: function (event_name) {
            $rootScope.$emit(event_name);
        }
    };
});

/*  Controllers    */

// App controller is in charge of managing all services for the application
appModule.controller('AppController', function ($scope, $location, $http) {

    // Common variables
    $scope.error = "";
    $scope.success = "";
    $scope.serverUrl = $location.protocol() + "://" + $location.host() + ":" + $location.port()

    // Configuration variables
    $scope.configs = [];
    $scope.configsLoading = false;
    $scope.configAction = "create"
    $scope.currentConfig = {}

    // Device variables
    $scope.deviceTypes = [];
    $scope.deviceTypesLoading = false;
    $scope.devices = [];
    $scope.currentDevice = {};
    $scope.devicesLoading = false;
    $scope.deviceAction = 'create'

    // Image variables
    $scope.images = [];
    $scope.imagesLoading = false;
    $scope.imageAction = "create"
    $scope.currentImage = {}
    $scope.imageFile = {};

    // Settings variables
    $scope.settings = {};
    $scope.settingsLoading = false;

    // Common functions
    $scope.clearError = function () {
        $scope.error = "";
    };
    $scope.clearSuccess = function () {
        $scope.success = "";
    };
    $scope.go = function (path) {
        $location.path(path);
    };

    // Configurations

    // Get all the configs in the database
    $scope.getConfigs = function () {
        $scope.configsLoading = true;
        $http
            .get('/api/configs')
            .then(function (response, status, headers, config) {
                $scope.configs = response.data;
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
            })
            .finally(function () {
                $scope.configsLoading = false;
            })
    };

    $scope.getConfigs()

    // Creates a request to add a new ZTP config to the database
    $scope.submitConfig = function () {
        $scope.clearError();
        $scope.clearSuccess();

        if (!($scope.currentConfig.name && $scope.currentConfig.deviceType && $scope.currentConfig.configuration)) {
            $scope.error = "Please complete all fields";
            return;
        }
        $scope.configsLoading = true;
        $http
            .post('/api/configs', $scope.currentConfig)
            .then(function (response, status, headers, config) {
                $scope.success = "Configuration added"
                $scope.getConfigs();
                $scope.go('configs')
                $scope.currentConfig = {}
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
                $scope.configsLoading = false;
            })
            .finally(function () {

            })
    };

    // Devices

    $scope.newDevice = function() {
        $scope.currentDevice =  {};
        $scope.deviceAction = "create";
        $scope.go('devices/detail');
    };

    $scope.checkDeviceTypeSelected = function(value, index, array) {
        if (!($scope.currentDevice.deviceType)){
            return false;
        }
        return value.deviceType.name === $scope.currentDevice.deviceType.name;
    };
    $scope.getXrDevices = function(value, index, array) {
        return value.deviceType.name === "iOS-XR";
    };
    $scope.getNxDevices = function(value, index, array) {
        return value.deviceType.name === "NX-OS";
    };

    $scope.getDeviceTypes = function () {
        $scope.deviceTypesLoading = true;
        $http
            .get('/api/devices/types')
            .then(function (response, status, headers, config) {
                $scope.deviceTypes = response.data;
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
            })
            .finally(function () {
                $scope.deviceTypesLoading = false;
            })
    };
    $scope.getDeviceTypes()

    $scope.getDevices = function () {
        $scope.devicesLoading = true;
        $http
            .get('/api/devices')
            .then(function (response, status, headers, config) {
                $scope.devices = response.data;
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
            })
            .finally(function () {
                $scope.devicesLoading = false;
            })
    };
    $scope.getDevices();
    // Refresh devices each 10 seconds. TODO: This should be done with websockets 
    setInterval(function(){ $scope.getDevices(); }, 10000);

    $scope.submitDevice = function () {
        $scope.clearError();
        $scope.clearSuccess();

        if (!($scope.currentDevice.hostname && $scope.currentDevice.serial && $scope.currentDevice.fixedIp && $scope.currentDevice.deviceType && $scope.currentDevice.image && $scope.currentDevice.config)) {
            $scope.error = "Please complete all fields";
            return;
        }
        $scope.devicesLoading = true;

        if ($scope.deviceAction === "edit"){
            $http
            .put('/api/devices', $scope.currentDevice)
            .then(function (response, status, headers, config) {
                $scope.success = "Device added"
                $scope.getDevices();
                $scope.go('devices')
                $scope.currentDevice = {}
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
                $scope.devicesLoading = false;
            })
            .finally(function () {

            })
        }
        else {
            $scope.currentDevice.status = "Configured"
            $http
            .post('/api/devices', $scope.currentDevice)
            .then(function (response, status, headers, config) {
                $scope.success = "Device added"
                $scope.getDevices();
                $scope.go('devices')
                $scope.currentDevice = {}
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
                $scope.devicesLoading = false;
            })
            .finally(function () {

            });
        }
    };

    $scope.removeDevice = function () {
        $scope.clearError();
        $scope.clearSuccess();

        $http
        .delete('/api/devices?serial=' + $scope.currentDevice.serial)
        .then(function (response, status, headers, config) {
            $scope.success = "Device removed"
            $scope.getDevices();
            $scope.go('devices')
            $scope.currentDevice = {}
        })
        .catch(function (response, status, headers, config) {
            $scope.error = response.data
            $scope.devicesLoading = false;
        })
        .finally(function () {

        });
    };



    $scope.selectDevice = function(device){
        $scope.currentDevice = angular.copy(device);
        $scope.deviceAction = 'edit'
        $scope.go('/devices/detail')
    };

    // Images
    $scope.getImages = function () {
        $scope.imagesLoading = true;
        $http
            .get('/api/images')
            .then(function (response, status, headers, config) {
                $scope.images = response.data;
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
            })
            .finally(function () {
                $scope.imagesLoading = false;
            })
    };
    $scope.getImages();

    $scope.saveImageFile = function (files) {
        $scope.imageFile = files[0];
    }

    $scope.submitImage = function () {
        $scope.clearError();
        $scope.clearSuccess();

        if (!($scope.currentImage.name && $scope.currentImage.deviceType && $scope.imageFile)) {
            $scope.error = "Please complete all fields";
            return;
        }
        $scope.imagesLoading = true;

        var fd = new FormData();
        //Add the selected file
        fd.append("file", $scope.imageFile);
        fd.append("deviceType", $scope.currentImage.deviceType.name);
        fd.append("name", $scope.currentImage.name);

        $http.post("/api/images", fd, {
            headers: { 'Content-Type': undefined },
        })
            .then(function (response, status, headers, config) {

                $scope.success = "Image added"
                $scope.getImages();
                $scope.go('images');
                $scope.currentImage = {};
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
                $scope.imagesLoading = false;
            })
            .finally(function () {

            });

    };

    // Settings 
    $scope.getSettings = function () {
        $scope.settingsLoading = true;
        $http
            .get('/api/settings')
            .then(function (response, status, headers, config) {
                $scope.settings = response.data;
            })
            .catch(function (response, status, headers, config) {
                $scope.error = response.data
            })
            .finally(function () {
                $scope.settingsLoading = false;
            })
    };
    $scope.getSettings();

    $scope.submitSettings = function () {
        $scope.clearError();
        $scope.clearSuccess();
        $scope.settingsLoading = true;

        $http
        .post('/api/settings', $scope.settings)
        .then(function (response, status, headers, config) {
            $scope.success = "Settings updated"
            $scope.getSettings();
        })
        .catch(function (response, status, headers, config) {
            $scope.error = response.data;
            $scope.settingsLoading = false;
        })
        .finally(function () {
            
        });
    };

});
