# twitch-downloader

Easily download twitch VODs and Clips.

## Usage

![Uage](doc/usage.gif?raw=true)

## Download

You can download the latest release for Windows, Macos and Linux here:
https://github.com/jybp/twitch-downloader/releases

## Flags

|&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Flag&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| Description |
| --- | --- |
| `-url` | The URL of the twitch VOD or Clip to download. |
| `-q` | Quality of the video to download. Omit this flag to print the available qualities.<br>Use "best" to automatically select the highest quality. |
| `-o` | Path where the video will be downloaded. Example: `-o my-video.ts`. (optional) |
| `-start` | Specify "start" to download a subset of the VOD. Example: 1h23m45s (optional) |
| `-end` | Specify "end" to download a subset of the VOD. Example: 1h34m56s (optional) |
| `-client-id` | Use a specific twitch.tv API client ID. Using any other client id other than twitch own client id might not work. (optional) |
| `-v` | Verbose errors. (optional) |

## Build from source

1. Install the latest version of Go https://golang.org/
2. Clone the git repository
3. Run `go build -ldflags "-X main.defaultClientID=kimne78kx3ncx6brgo4mv6wki5h1ko" ./cmd/twitchdl`
