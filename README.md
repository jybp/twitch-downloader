# twitch-downloader

Easily download twitch VODs on Windows, MacOs and Linux with no dependencies whatsoever.

## Usage

![Usage](doc/usage.gif?raw=true)

## Download

You can download the latest release here:
https://github.com/jybp/twitch-downloader/releases

## Flags

|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Flag&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| Description |
| --- | --- |
| `-vod` | The ID or absolute URL of the twitch VOD to download. https://www.twitch.tv/videos/12345 is the VOD with ID "12345". |
| `-q` | Quality of the VOD to download. Omit this flag to print the available qualities. |
| `-o` | Path where the VOD will be downloaded. (optional)|
| `-start` | Specify "start" to download a subset of the VOD. Example: 1h23m45s (optional) |
| `-end` | Specify "end" to download a subset of the VOD. Example: 1h34m56s (optional) |
| `-client-id` | Use a specific twitch.tv API client ID. Using any other client id other than twitch own client id might not work. (optional) |

## Build from source

1. Get a twitch Client ID by registering an application https://dev.twitch.tv/console/apps/create
2. Install the latest version of Go https://golang.org/
3. Clone the git repository
4. Set the current directory to `cmd/twitchdl` and run `go build -ldflags "-X main.defaultClientID=TwitchClientID"`
