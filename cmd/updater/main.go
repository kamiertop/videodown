package main

import "flag"

var latestDownloadURL = flag.String("download-url", "", "The Latest Download URL.")

func main() {
	flag.Parse()
}
