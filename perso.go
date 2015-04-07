package main

import (
	"flag"
	"log"
	"os"
)

func findCached(url string, cache *caches) {
	cacheReq, err := makeCacheRequest(url)
	if err != nil {
		log.Print("Invalid URL ", url, ": ", err)
		return
	}

	cache.requestCh <- *cacheReq
	data := <-cacheReq.data

	data.WriteTo(os.Stdout)
}

func main() {
	flag.Parse()

	url := flag.Arg(0)
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

	findCached(url, caches)
}
