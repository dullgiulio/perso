package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	conf := newConfig()
	conf.parseFlags()

	// Provides help text based on user configuration
	help := newHelp(conf.keys)

	// Keep track of what is searcheable
	indexer := newMailIndexer(conf.keys)

	// Handle all requests to the cache (searching or adding)
	caches := newCaches(indexer, conf.root)
	go caches.run()

	// First crawl. HTTP listener won't start before
	crawler := newCrawler(indexer, caches, conf.root)
	crawler.scan()

	if conf.interval > 0 {
		// Keep crawling for new or deleted messages
		go func() {
			for {
				<-time.After(time.Duration(conf.interval))
				crawler.scan()
			}
		}()
	}

	// Handle all HTTP requests here
	handler := newHttpHandler(help, caches, conf, indexer)
	log.Fatal(http.ListenAndServe(conf.listen, handler))
}
