package main

import (
	"net/mail"
	"sort"
)

type cacheString map[string]mailFiles

type caches struct {
	data      map[string]cacheString
	cancelCh  chan struct{}
	mailCh    chan cacheMail
	requestCh chan cacheRequest
	addCh     chan cacheEntry
	removeCh  chan mailFiles
}

type cacheRequest struct {
	header string
	value  string
	index  int
	limit  int
	oldest bool
	data   chan mailFiles
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
		data:      make(map[string]cacheString),
		cancelCh:  make(chan struct{}),
		requestCh: make(chan cacheRequest),
		addCh:     make(chan cacheEntry),
		removeCh:  make(chan mailFiles),
	}
}

func (c *caches) initCachesString(name string) {
	c.data[name] = make(map[string]mailFiles)
}

func (c *caches) add(entry cacheEntry) {
	name, key, value := entry.name, entry.key, entry.value

	if _, found := c.data[name][key]; !found {
		c.data[name][key] = newMailFiles()
	}

	c.data[name][key] = append(c.data[name][key], value)
}

func (c *caches) remove(files mailFiles) {
	for name := range c.data {
		for k := range c.data[name] {
			sort.Sort(c.data[name][k])
			c.data[name][k] = sliceDiff(c.data[name][k], files)
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
			if cache, found := c.data[r.header]; found {
				ids := cache[r.value]
				if r.limit == 0 {
					r.limit = len(ids)
				}

				// Copy only elements between r.index and r.index + r.limit
				result := mailFiles(make([]mailFile, r.limit))
				for i, j := 0, r.index; i < r.limit; i, j = i+1, j+1 {
					result[i] = ids[j]
				}

				// Invert result. This assumes the result is ordered by date
				// TODO: Not sure this is already ordered by date
				if r.oldest {
					for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
						result[i], result[j] = result[j], result[i]
					}
				}

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
