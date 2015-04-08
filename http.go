package main

import (
	"io"
	"log"
	"net/http"
)

type httpHandler struct {
	cache   *caches
	indexer *mailIndexer
}

func newHttpHandler(cache *caches, indexer *mailIndexer) *httpHandler {
	return &httpHandler{
		cache:   cache,
		indexer: indexer,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	h.writeFromURL(r.URL.Path, io.Writer(w))
}

func (h *httpHandler) writeFromURL(url string, w io.Writer) {
	cacheReq, err := makeCacheRequest(url)
	if err != nil {
		log.Print(url, ": ", err)
		return
	}

	if !h.indexer.keys.has(cacheReq.header) {
		log.Print(url, ": ", errInvalidURL)
		return
	}

	cacheReq.match = h.indexer.keys.keyType(cacheReq.header)

	// Copy request object as it will be modified
	h.cache.requestCh <- *cacheReq
	data := <-cacheReq.data

	data.WriteTo(w)
}
