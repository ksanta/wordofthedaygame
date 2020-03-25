# Word of the day game
A simple game that will give you a word and prompt you to guess the correct definition. All the words have been
previously featured as a "word of the day" on a popular dictionary website and are scraped from the website and then 
stored into a local cache file.

This Go application demonstrates the use of goroutines, channels, website scraping, some file manipulation and prompting
the user for input.

# Install and play the game
The simplest way to play the game is to use Docker. This way, you don't even need to have any Go build tools installed.

```shell script
docker build -t ksanta/wordofthedaygame .

# This step can take a few minutes as it scrapes the dictionary website for all the words of the day.
docker run -it --name wordofthedaygame ksanta/wordofthedaygame

# Once the container is created with all the words pre-cached, run this to play the game
docker start -i wordofthedaygame
```

# Development
This is an optional step and only needed during local development. Skip this if all you want to do is to play the game.
Go 1.13 is required to build the binary.
```shell script
go build
````
