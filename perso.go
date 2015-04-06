package main

import (
	"flag"
	"fmt"
	"log"
)

func findCached(header, value string, cache *caches) {
	cacheReq := cacheRequest{
		name: "to",
		str:  value,
		data: make(chan []mailFile),
	}

	cache.requestCh <- cacheReq
	data := <-cacheReq.data

	fmt.Printf("%v\n", data)
}

func main() {
	flag.Parse()

	email := flag.Arg(0)
	root := flag.Arg(1)

	cacheReadyCh := make(chan bool)

	caches := newCaches(root)
	caches.initCachesString("from")
	caches.initCachesString("to")
	go caches.run(cacheReadyCh)

	// Wait for the cache to warm up.
	<-cacheReadyCh

	ready := true
	/*
		ready := false

		select {
		case ready = <-cacheReadyCh:
		default: // XXX: Is this needed?
		}
	*/

	if !ready {
		log.Print("Cannot use cache now")
	} else {
		findCached("to", email, caches)
	}

	caches.cancel()
}
