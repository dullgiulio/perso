package main

import (
	"flag"
	"os"
)

func findCached(header, value string, cache *caches) {
	cacheReq := cacheRequest{
		header: "to",
		value:  value,
		index:  2,
		limit:  1,
		data:   make(chan mailFiles),
	}

	cache.requestCh <- cacheReq
	data := <-cacheReq.data

	data.WriteTo(os.Stdout)
}

func main() {
	flag.Parse()

	email := flag.Arg(0)
	root := flag.Arg(1)

	indexKeys := []string{"", "from", "to"}

	caches := newCaches(root)
	for i := range indexKeys {
		caches.initCachesString(indexKeys[i])
	}

	go caches.run()

	indexer := newMailIndexer(indexKeys)
	crawler := newCrawler(indexer, caches, root)
	crawler.scan()

	findCached("to", email, caches)

	caches.cancel()
}
