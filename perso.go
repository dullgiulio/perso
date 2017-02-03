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
		go func() {
			notify, err := newNotify(conf.root)
			if err != nil {
				log.Fatal("inotify setup error: ", err)
			}

			// Keep crawling for new or deleted messages
			crawler.run(
				notify.eventsChannel(),
				notify.errorsChannel(),
				time.Tick(time.Duration(conf.interval)),
			)
		}()
	}

	// Handle all HTTP requests here
	handler := newHttpHandler(help, caches, conf, indexer)
	srv := &http.Server{
		Addr:         conf.listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(handler.run(srv))
}
