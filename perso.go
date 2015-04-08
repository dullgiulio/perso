package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	flag.Parse()

	root := flag.Arg(0)

	indexKeys := makeIndexKeys()
	indexKeys.add("", keyTypeAny)

	// TODO: Defaults. Should come from args
	indexKeys.add("from", keyTypeAddr)
	indexKeys.add("to", keyTypeAddr)
	indexKeys.add("x-php-originating-script", keyTypePart)

	crawlInterval := 2 * time.Second

	// Keep track of what is searcheable
	indexer := newMailIndexer(indexKeys)

	// Handle all requests to the cache (searching or adding)
	caches := newCaches(indexer, root)
	go caches.run()

	// First crawl. HTTP listener won't start before
	crawler := newCrawler(indexer, caches, root)
	crawler.scan()

	// Keep crawling for new or deleted messages
	go func() {
		for {
			<-time.After(crawlInterval)
			crawler.scan()
		}
	}()

	// Handle all HTTP requests here
	handler := newHttpHandler(caches, indexer)
	log.Fatal(http.ListenAndServe(":8888", handler))
}
