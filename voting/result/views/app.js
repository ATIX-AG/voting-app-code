var app = angular.module('catsvsdogs', []);
var socket = io.connect({transports:['polling']});

var bg1 = document.getElementById('background-stats-1');
var bg2 = document.getElementById('background-stats-2');

app.controller('statsCtrl', function($scope){
  $scope.aPercent = 50;
  $scope.bPercent = 50;
  $scope.aOption = 'Imperative';
  $scope.bOption = 'Declarative';

  const updateScores = function(){
    socket.on('scores', function (json) {
       var a = parseInt(json.a || 0);
       var b = parseInt(json.b || 0);

       var percentages = getPercentages(a, b);

       bg1.style.width = percentages.a + "%";
       bg2.style.width = percentages.b + "%";

       $scope.$apply(function () {
         $scope.aPercent = percentages.a;
         $scope.bPercent = percentages.b;
         $scope.total = a + b;
       });
    });
  };

  const updateOptions = function(){
    socket.on('options', function (json) {
       $scope.$apply(function () {
         $scope.aOption = json.a;
         $scope.bOption = json.b;
       });
    });
  };

  var init = function() {
    document.body.style.opacity=1;
    updateOptions();
    updateScores();
  };
  socket.on('message',function(data) {
    init();
  });
});

function getPercentages(a, b) {
  var result = {};

  if (a + b > 0) {
    result.a = Math.round(a / (a + b) * 100);
    result.b = 100 - result.a;
  } else {
    result.a = result.b = 50;
  }

  return result;
}
