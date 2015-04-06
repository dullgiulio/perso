package main

import (
	"flag"
	"os"
)

func findCached(header, value string, cache *caches) {
	cacheReq := cacheRequest{
		name: "to",
		str:  value,
		data: make(chan mailFiles),
	}

	cache.requestCh <- cacheReq
	data := <-cacheReq.data

	data.WriteTo(os.Stdout)
}

func main() {
	flag.Parse()

	email := flag.Arg(0)
	root := flag.Arg(1)

	caches := newCaches(root)
	caches.initCachesString("from")
	caches.initCachesString("to")
	go caches.run()

	findCached("to", email, caches)

	caches.cancel()
}
