package main

import (
	"flag"
	"fmt"
	"log"
)

func findCached(header, value string, cache *caches) {
	cacheReq := cacheRequest{
		name: "to",
		str:  "imps.dev@kuehne-nagel.com",
		data: make(chan []mailFile),
	}

	cache.requestCh <- cacheReq
	data := <-cacheReq.data

	fmt.Printf("%v\n", data)
}

func main() {
	flag.Parse()
	root := flag.Arg(0)

	cacheReadyCh := make(chan bool)

	caches := newCaches(root)
	caches.initCachesString("from")
	caches.initCachesString("to")
	caches.initCachesTime("date")
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
		findCached("to", "dullgiulio@gmail.com", caches)
	}

	caches.cancel()
}
