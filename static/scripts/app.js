$(document).on("ready", function () {
    $('#whoWon').hide();
    $('#countDownBox').hide();

    $('.reset').on('click', function (e) {
        window.location.reload(true);
    });

    $('img.horse-option').click(function () {
        $('.horse-selected').removeClass('horse-selected'); // removes the previous selected class
        $(this).addClass('horse-selected'); // adds the class to the clicked image
    });
    // todo: submit answer directly when clicking on an option
    $('.alternative').click(function () {
        $('.alt-selected').removeClass('alt-selected'); // removes the previous selected class
        $(this).addClass('alt-selected'); // adds the class to the clicked image
    });

// Initializes game with players' chosen preferences
    $('.submit').on('click', function (e) {
        if (!document.getElementById("nameEntryOne").value || $('.horse-selected')[0].id == undefined){
          return
        }

        $('#selections').hide();

        let player = {
            PlayerDetailsResp: {
                Name: document.getElementById("nameEntryOne").value,
                Icon: $('.horse-selected')[0].id
            }
        };
        connection.send(JSON.stringify(player))
        $('#startGameBox').show()
    });

    $('#start-game-btn').on('click', function(e){
        $.get("/start", function() {
          console.log( "game started" );
        })
        $('#startGameBox').hide()
    });


    $('.alternative').on('click', function (e) {
        let selected = $('.alt-selected').attr('id');
        let response = 0;

        if (selected === "questionAlt1") {
            response = "1"
        } else if (selected === "questionAlt2") {
            response = "2"
        } else if (selected === "questionAlt3") {
            response = "3"
        }

        let message = {
            PlayerResponse: {
                Response: response
            }
        };
        connection.send(JSON.stringify(message));
        // todo: instead of hiding, server will give feedback of right/wrong -> change colour
        // $('#question-area').hide()
    });

});
//End of document onReady


//Helper Functions
function displayWinner(win, pic) {
    $('#whoWon').show();
    victory.play();
    victory.currentTime = 0;
        
    $('#winnerName').text(win);
    $('#winPic').html("<img src=" + pic + ">");
}

function gameReset() {
  window.location.reload(true);
}

//Variables to initialize
const API_IP = "52.63.119.7:8080";

var snd = new Audio('./bugle.wav');
var victory = new Audio('./victory.mp3');

window.WebSocket = window.WebSocket || window.MozWebSocket;

var connection = new WebSocket('ws://'+ API_IP + '/game');

connection.onerror = function (error) {
    console.log(error);
};

var showCountdown = function () {
    $('#startGameBox').hide()
    $('#countDownBox').show();
    snd.play();
    snd.currentTime = 0;
        
    setTimeout(function() {$('#num').text("3");}, 500);
    setTimeout(function() {$('#num').text("2");}, 1500);
    setTimeout(function() {$('#num').text("1");}, 2500);
    setTimeout(function() {$('#num').text("Go!");}, 3500);
    setTimeout(function() {$('#countDownBox').hide();}, 4000);
};

var showQuestion = function (question) {
    $('#question-area').show()
    $('.alternative').css('background-color', 'white');

    $('#questionWord').text(question.WordToGuess);
    $('#questionAlt1').text("1: " + question.Definitions[0]);
    $('#questionAlt2').text("2: " + question.Definitions[1]);
    $('#questionAlt3').text("3: " + question.Definitions[2]);
    $('#question-area').show();
};

// Updates the placement of all the horses
var updateGame = function (summary) {
    for (let i = 0; i < summary.PlayerStates.length; i++) {
        const player = summary.PlayerStates[i];
        const name = player.Name;
        const active = player.Active;  // todo: use this to represent a disconnected player mid-game
        const horse = player.Icon;

        const track = $('table.track' + i);

        // Display player name
        $('#player' + i + 'Name').text(name);

        // Display the player's chosen horse
        track.children().children().children().children('img').attr('src', 'images/' + horse + '.png');

        const targetPoints = 500;
        const maxPosition = 60;
        let position = Math.floor(player.Score / targetPoints * maxPosition);
        position = Math.min(position, maxPosition);

        track.children().children().children().children('img').addClass('empty');
        track.children().children().children().children('img').eq(position).removeClass('empty');
    }
};

var endGame = function (summary) {
    $('#question-area').hide()
    displayWinner(summary.Winner, "images/" + summary.Icon + ".png")
};

var showError = function(message) {
    $('#errorBox').show()
    $('#errorMessage').text(message.Message)
}

var showResult = function(correct) {
    if (correct){
        $('.alt-selected').css('background-color', 'green')
    } else {
        $('.alt-selected').css('background-color', 'red')
    }
}

connection.onmessage = function (wsMessage) {
    try {
        console.log("Received: " + wsMessage.data);
        let data = JSON.parse(wsMessage.data);

        if (data.hasOwnProperty('Welcome')) {
            // todo: should display "waiting for other players". Can display target score?

        } else if (data.hasOwnProperty('Error')) {
            showError(data.Error)

        } else if (data.hasOwnProperty('AboutToStart')) {
            showCountdown(data.AboutToStart)

        } else if (data.hasOwnProperty('PresentQuestion')) {
            showQuestion(data.PresentQuestion)

        } else if (data.hasOwnProperty('RoundSummary')) {
            updateGame(data.RoundSummary)

        } else if (data.hasOwnProperty('Summary')) {
            endGame(data.Summary)

        } else if (data.hasOwnProperty('PlayerResult')) {
            showResult(data.PlayerResult.Correct)
        }

    } catch (e) {
        console.log(e);
        console.log('Unexpected JSON: ', wsMessage.data);
    }
};
