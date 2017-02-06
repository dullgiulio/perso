package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type errorNotFound string

func (e errorNotFound) Error() string {
	return string(e)
}

var errNotFound = errorNotFound("Not found")

type httpHandler struct {
	helpTmpl *help
	cache    *caches
	config   *config
	indexer  *mailIndexer
	paths    []string
}

func newHttpHandler(help *help, cache *caches, config *config, indexer *mailIndexer) *httpHandler {
	return &httpHandler{
		helpTmpl: help,
		cache:    cache,
		config:   config,
		indexer:  indexer,
		paths:    config.keys.all(),
	}
}

func (h *httpHandler) run(srv *http.Server) error {
	r := mux.NewRouter()
	r.HandleFunc("/", h.forward("latest/0"))
	r.HandleFunc("/help", h.help)
	r.HandleFunc("/latest/{selector}", h.messages("", false))
	r.HandleFunc("/oldest/{selector}", h.messages("", true))
	for key := range h.config.keys {
		if key == "" {
			continue
		}
		prefix := "/" + key
		r.HandleFunc(prefix, h.list(key))
		r.HandleFunc(prefix+"/", h.forward(""))
		r.HandleFunc(prefix+"/{value}", h.forward("/latest/0"))
		r.HandleFunc(prefix+"/{value}/latest", h.forward("/0"))
		r.HandleFunc(prefix+"/{value}/oldest", h.forward("/0"))
		r.HandleFunc(prefix+"/{value}/latest/{selector}", h.messages(key, false))
		r.HandleFunc(prefix+"/{value}/oldest/{selector}", h.messages(key, true))
	}
	srv.Handler = r
	return srv.ListenAndServe()
}

func (httpHandler) forward(url string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		url = strings.TrimRight(r.URL.Path+url, "/")
		// TODO: Is this the right code for POST/DELETE requests?
		http.Redirect(w, r, url, 307)
	}
}

func (h *httpHandler) help(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not supported", 405)
		return
	}
	tmpl := newTemplate()
	if err := tmpl.render(w, h.helpTmpl); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
	}
}

func (h *httpHandler) messages(key string, oldest bool) func(w http.ResponseWriter, r *http.Request) {
	return h.handler(func(h *httpHandler, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		cr := newCacheRequest()
		cr.oldest = oldest
		cr.header = key
		cr.value = vars["value"]
		if !h.indexer.keys.has(cr.header) {
			return errNotFound
		}
		if vars["selector"] == "" {
			return errNotFound // XXX: bad request
		}
		if err := selector(vars["selector"]).parse(cr); err != nil {
			return errNotFound // XXX: bad request
		}
		w.Header().Set("Content-Type", "text/plain")
		return h.writeMessages(cr, w)
	})
}

func (h *httpHandler) list(k string) func(w http.ResponseWriter, r *http.Request) {
	return h.handler(func(h *httpHandler, w http.ResponseWriter, r *http.Request) error {
		cr, err := makeCacheListRequest(k)
		if err != nil {
			return err
		}
		h.cache.listCh <- cr
		data := <-cr.data
		if data == nil || len(data) == 0 {
			return errNotFound
		}
		tmpl := newTemplate()
		tw := newListTemplate(cr.header, data)

		w.Header().Set("Content-Type", "text/html")
		return tmpl.render(w, tw)
	})
}

func (h *httpHandler) handler(fn func(h *httpHandler, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch err := fn(h, w, r); err {
		case nil:
			// XXX: Log successful requests?
			return
		case errNotFound:
			log.Print(r.URL.Path, ": ", err)
			http.NotFound(w, r)
		default:
			log.Print(r.URL.Path, ": ", err)
			http.Error(w, err.Error(), 500)
		}
	}
}

func (h *httpHandler) writeMessages(cr *cacheRequest, w io.Writer) error {
	cr.match = h.indexer.keys.keyType(cr.header)

	h.cache.requestCh <- cr
	data := <-cr.data

	if data == nil || len(data) == 0 {
		return errNotFound
	}

	data.writeTo(w, h.config)
	return nil
}
