package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
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
	if r.URL.Path == "/help" {
		h.help.write(w)
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

	var b bytes.Buffer

	fmt.Fprintln(&b, "<ul>")
	for _, val := range data {
		fmt.Fprintf(&b, `<li><a href="/%s/%s/latest/0">%s</a></li>`, neturl.QueryEscape(cr.header), neturl.QueryEscape(val), val)
	}
	fmt.Fprintln(&b, "</ul>")

	w.Write(template(fmt.Sprintf("Perso - List for %s", cr.header), b.String()))

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
