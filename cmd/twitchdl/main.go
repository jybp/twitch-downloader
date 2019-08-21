package main

import (
	"flag"
)

// Must be injected at build time using the -ldflags flag.
// go build ./cmd/twitchdl -ldflags "-X cmd/twitchdl.defaultClientID=YourClientID"
var defaultClientID string

func main() {
	var clientID, vodID, quality, output string
	var info bool
	flag.StringVar(&vodID, "vodID", "", `The ID of the VOD to download.
Can be inferred from the URL:
https://www.twitch.tv/videos/123 is the VOD with ID "123".`)
	flag.BoolVar(&info, "info", false, "Print available qualities.")
	flag.StringVar(&quality, "quality", "", "Quality of the VOD to download. Must not be present with the -info flag.")
	flag.StringVar(&output, "output", "", "Output directory for extracted files. Must not be present with the -info flag.")
	flag.StringVar(&clientID, "clientID", "", "optional: Use a specific twitch.tv API client ID.")
	flag.Parse()

	if len(vodID) == 0 || (info && (len(quality) > 0 || len(output) > 0)) {
		flag.PrintDefaults()
		return
	}

	if len(clientID) == 0 {
		clientID = defaultClientID
	}

	panic("TODO")
}
