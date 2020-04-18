const API_IP = location.host;

var snd = new Audio('./bugle.wav');
var victory = new Audio('./victory.mp3');

$(document).ready(function () {
    $('#whoWon').hide();
    $('#countDownBox').hide();

    $('.reset').click(function () {
        window.location.reload(true);
    });

    $('img.horse-option').click(function () {
        $('.horse-selected').removeClass('horse-selected'); // removes the previous selected class
        $(this).addClass('horse-selected'); // adds the class to the clicked image
    });

    $('.definition').click(function () {
        $(this).addClass('alt-selected'); // adds the class to the clicked image

        const response = $(this).data('option')
        let message = {
            PlayerResponse: {
                Response: response
            }
        };
        connection.send(JSON.stringify(message));
        $('.definition').css("pointer-events", "none")
    });

    // Initialises game with players' chosen preferences
    $('.submit').on('click', function () {
        if (!document.getElementById("nameEntryOne").value || $('.horse-selected')[0].id == undefined) {
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

    $('#start-game-btn').on('click', function (e) {
        $.get("http://" + API_IP + "/start", function () {
            console.log("game started");
        });
        $('#startGameBox').hide()
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

//Variables to initialize
window.WebSocket = window.WebSocket || window.MozWebSocket;
var connection = new WebSocket('ws://' + API_IP + '/game');

connection.onerror = function (error) {
    console.log(error);
};

var showCountdown = function () {
    $('#startGameBox').hide()
    $('#countDownBox').show();
    snd.play();
    snd.currentTime = 0;

    setTimeout(function () {
        $('#num').text("3");
    }, 500);
    setTimeout(function () {
        $('#num').text("2");
    }, 1500);
    setTimeout(function () {
        $('#num').text("1");
    }, 2500);
    setTimeout(function () {
        $('#num').text("Go!");
    }, 3500);
    setTimeout(function () {
        $('#countDownBox').hide();
    }, 4000);
};

var showQuestion = function (question) {
    let definitions = $('.definition')
    definitions.removeClass('alt-selected'); // removes the previous selected class
    definitions.css('background-color', 'white')
    definitions.css('pointer-events', 'auto')

    $('#word-to-guess').text(question.WordToGuess);
    $('#definition0').text(question.Definitions[0]);
    $('#definition1').text(question.Definitions[1]);
    $('#definition2').text(question.Definitions[2]);

    $('#question-area').show();
};

// Updates the placement of all the horses
var updateGame = function (summary) {
    for (let i = 0; i < summary.PlayerStates.length; i++) {
        const player = summary.PlayerStates[i];
        const name = player.Name;

        let horseIcon = player.Icon;
        if (!player.Active) {
            horseIcon = "dead"
        }

        let track = $('#track' + i)

        // Create a new track if this is a new player
        if (track.length === 0) {
            track = $('#template-track')
                .clone()
                .appendTo('#tracks')
                .attr('id', 'track' + i)
                .css('display', 'flex')

            track.find('img')
                .attr('id', 'horse' + i)

            track.find('span')
                .attr('id', 'player' + i)
        }

        // Set the player name
        const playerSpan = $('#player' + i)
        playerSpan.text(name);

        // Set the horse icon
        const horse = $("#horse" + i)
        horse.attr('src', 'images/' + horseIcon + '.png')

        // Set the horse position
        const targetPoints = 500;
        const maxPosition = 100;
        let position = Math.floor(player.Score / targetPoints * maxPosition);
        position = Math.min(position, maxPosition);

        horse.animate({left: position + "%"}, "slow");
    }
};

var endGame = function (summary) {
    $('#question-area').hide()
    displayWinner(summary.Winner, "images/" + summary.Icon + ".png")
};

var showError = function (message) {
    $('#errorBox').show()
    $('#errorMessage').text(message.Message)
}

// showResult lets the player know which answer was correct
var showResult = function (playerResult) {
    // Make the correct answer green
    $('.definition[data-option=' + playerResult.CorrectAnswer + ']')
        .css('background-color', 'green')

    // If the player guessed wrong, make the selected definition red
    if (!playerResult.Correct) {
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
            showResult(data.PlayerResult)
        }

    } catch (e) {
        console.log(e);
        console.log('Unexpected JSON: ', wsMessage.data);
    }
};
