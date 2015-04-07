package main

import (
	"net/mail"
	"sort"
	"time"
)

type cacheString map[string]mailFiles

type caches struct {
	str       map[string]cacheString
	cancelCh  chan struct{}
	mailCh    chan cacheMail
	requestCh chan cacheRequest
	addCh     chan cacheEntry
	removeCh  chan mailFiles
}

type cacheRequest struct {
	name     string
	time     time.Time
	str      string
	submatch bool
	lower    bool
	data     chan mailFiles
}

type cacheEntry struct {
	name  string
	key   string
	value mailFile
}

type cacheMail struct {
	id      mailFile
	headers mail.Header
}

func newCacheRequest() *cacheRequest {
	return &cacheRequest{
		data: make(chan mailFiles),
	}
}

func newCaches(root string) *caches {
	return &caches{
		str:       make(map[string]cacheString),
		cancelCh:  make(chan struct{}),
		requestCh: make(chan cacheRequest),
		addCh:     make(chan cacheEntry),
		removeCh:  make(chan mailFiles),
	}
}

func (c *caches) initCachesString(name string) {
	c.str[name] = make(map[string]mailFiles)
}

func (c *caches) add(entry cacheEntry) {
	name, key, value := entry.name, entry.key, entry.value

	if _, found := c.str[name][key]; !found {
		c.str[name][key] = newMailFiles()
	}

	c.str[name][key] = append(c.str[name][key], value)
}

func (c *caches) getString(name string, key string) mailFiles {
	return c.str[name][key]
}

func (c *caches) getKeysString(name string) []string {
	keys := make([]string, 0)

	for k := range c.str[name] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (c *caches) remove(files mailFiles) {
	for name := range c.str {
		for k := range c.str[name] {
			sort.Sort(c.str[name][k])
			c.str[name][k] = sliceDiff(c.str[name][k], files)
		}
	}
}

func (c *caches) cancel() {
	c.cancelCh <- struct{}{}
}

func (c *caches) request(r cacheRequest) {
	c.requestCh <- r
}

func (c *caches) run() {
	for {
		select {
		case <-c.cancelCh:
			return
		case r := <-c.requestCh:
			if r.str == "" {
				r.data <- nil
				continue
			}

			if cache, found := c.str[r.name]; found {
				ids := cache[r.str]
				result := mailFiles(make([]mailFile, len(ids)))
				copy(result, ids)
				r.data <- result
				continue
			}

			r.data <- nil
		case entry := <-c.addCh:
			c.add(entry)
		case files := <-c.removeCh:
			c.remove(files)
		}
	}
}
