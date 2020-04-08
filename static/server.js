var WebSocketServer = require('websocket').server;
var http = require('http');

var server = http.createServer(function(request, response) {
  // process HTTP request. Since we're writing just WebSockets
  // server we don't have to implement anything.
});
server.listen(1337, function() { });

// create the server
wsServer = new WebSocketServer({
  httpServer: server
});

// WebSocket server
wsServer.on('request', function(request) {
  var connection = request.accept(null, request.origin);

  // This is the most important callback for us, we'll handle
  // all messages from users here.
  connection.on('message', function(message) {
    if (message.type === 'utf8') {
    	if (message.utf8Data == 'start'){
    		connection.send('{"reply":1')
    	}
    }
  });

  connection.on('close', function(connection) {
    // close user connection
  });

  ticker = 0
  var updateGame = function(){
  	ticker +=1

  	var game={}
  	var players=[]
  	for (var i=0;i<8;i++){
  		var player = {}
  		player['name'] = "Name"+i
  		player['horse'] = 'Horse'+parseInt(i+1)
  		player['score'] = ticker
  		players.push(player)
  	}
  	game['players'] = players
  	message = {}
  	message['action'] = 'update-game'
  	message['game'] = game
  	connection.send(JSON.stringify(message))
  	if (ticker >= 60){
  		ticker = 0
  	}
  }

  var startCountDown = function(){
  	message = {}
  	message['action'] = 'countdown-start'
  	connection.send(JSON.stringify(message))
  }

  var endGame = function(){
  	message = {}
  	message['action'] = 'end-game'
  	winner={}
  	winner['name'] = 'Brendan'
  	winner['horse'] = 'Horse1'
  	message['winner'] = winner
  	connection.send(JSON.stringify(message))
  }

  var sendQuestion = function(){
  	var question = {}
  	question['word'] = 'shenanigans' + ticker
  	question['alt1'] = 'non-sense'
  	question['alt2'] = 'small talk'
  	question['alt3'] = 'engine part'
  	message = {}
  	message['action'] = 'show-question'
  	message['question'] = question

  	connection.send(JSON.stringify(message))

  }

  var t=setInterval(updateGame,500);

  var t=setInterval(sendQuestion,1000);

  setTimeout(startCountDown, 5000)


  setTimeout(endGame, 15000)


});