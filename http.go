package main

import (
	"io"
	"log"
	"net/http"
)

type httpHandler struct {
	help    *help
	cache   *caches
	indexer *mailIndexer
}

func newHttpHandler(help *help, cache *caches, indexer *mailIndexer) *httpHandler {
	return &httpHandler{
		help:    help,
		cache:   cache,
		indexer: indexer,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/help" {
		h.help.write(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if err := h.writeFromURL(r.URL.Path, io.Writer(w)); err != nil {
		switch err.(type) {
		case errorRedirect:
			http.Redirect(w, r, err.Error(), 307)
		default:
			log.Print(r.URL.Path, ": ", err)
			http.NotFound(w, r)
		}
	}
}

func (h *httpHandler) writeFromURL(url string, w io.Writer) error {
	cacheReq, err := makeCacheRequest(url)
	if err != nil {
		return err
	}

	if !h.indexer.keys.has(cacheReq.header) {
		return errInvalidURL
	}

	cacheReq.match = h.indexer.keys.keyType(cacheReq.header)

	// Copy request object as it will be modified
	h.cache.requestCh <- *cacheReq
	data := <-cacheReq.data

	if len(data) == 0 {
		return err
	}

	data.writeTo(w)
	return nil
}
