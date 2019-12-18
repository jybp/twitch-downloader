# twitch-downloader

Easily download twitch VODs on Windows, MacOs and Linux with no dependencies whatsoever.

## Usage

![Usage](doc/usage.gif?raw=true)

## Download

You can download the latest release here:
https://github.com/jybp/twitch-downloader/releases

## Flags

`-client-id` Use a specific twitch.tv API client ID. Usage is optional. 
Note: Using any other client id other than twitch own client id might not work.

`-o` Path where the VOD will be downloaded. Usage is optional.

`-q` Quality of the VOD to download. Omit this flag to print the available qualities.

`-vod` The ID or absolute URL of the twitch VOD to download. https://www.twitch.tv/videos/12345 is the VOD with ID "12345".

## Build from source

1. Get a twitch Client ID by registering an application https://dev.twitch.tv/console/apps/create
2. Install the latest version of Go https://golang.org/
3. Clone the git repository
4. Set the current directory to `cmd/twitchdl` and run `go build -ldflags "-X main.defaultClientID=TwitchClientID"`
