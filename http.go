package main

import (
	"io"
	"log"
	"net/http"
)

type errorNotFound string

func (e errorNotFound) Error() string {
	return string(e)
}

var errNotFound = errorNotFound("Not found")

type httpHandler struct {
	help    *help
	cache   *caches
	config  *config
	indexer *mailIndexer
	paths   []string
}

func newHttpHandler(help *help, cache *caches, config *config, indexer *mailIndexer) *httpHandler {
	return &httpHandler{
		help:    help,
		cache:   cache,
		config:  config,
		indexer: indexer,
		paths:   config.keys.all(),
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/latest/0", 307)
		return
	}

	if r.URL.Path == "/help" {
		tmpl := newTemplate()
		if err := tmpl.render(w, h.help); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
		}
		return
	}

	switch err := h.writeList(r.URL.Path, w); err {
	case nil:
		return
	case errNotFound:
		log.Print(r.URL.Path, ": ", err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if err := h.writeFromURL(r.URL.Path, w); err != nil {
		switch err.(type) {
		case errorRedirect:
			http.Redirect(w, r, err.Error(), 307)
		case errorNotFound:
			http.NotFound(w, r)
		default:
			log.Print(r.URL.Path, ": ", err)
			http.NotFound(w, r)
		}
	}
}

func (h *httpHandler) writeList(url string, w http.ResponseWriter) error {
	cr, err := makeCacheListRequest(url)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")

	if !h.indexer.keys.has(cr.header) {
		return errInvalidURL
	}

	h.cache.listCh <- *cr
	data := <-cr.data

	if data == nil || len(data) == 0 {
		return errNotFound
	}

	tmpl := newTemplate()
	tmplWriter := newListTemplate(cr.header, data)

	if err := tmpl.render(w, tmplWriter); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return nil
	}

	return nil
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

	if data == nil || len(data) == 0 {
		return errNotFound
	}

	data.writeTo(w, h.config)
	return nil
}
