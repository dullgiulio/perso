package main

import (
	"log"
	"net/mail"
	"sort"
	"strings"
	"time"
)

type cacheString map[string][]mailID
type cacheTime map[time.Time][]mailID

type caches struct {
	str       map[string]cacheString
	time      map[string]cacheTime
	cancelCh  chan struct{}
	mailCh    chan cacheMail
	requestCh chan cacheRequest
	sweepCh   chan []mailID
}

type cacheRequest struct {
	name     string
	time     time.Time
	str      string
	submatch bool
	lower    bool
	data     chan []mailID
}

type cacheMail struct {
	id      mailID
	headers mail.Header
}

func newCacheRequest() *cacheRequest {
	return &cacheRequest{
		data: make(chan []mailID),
	}
}

func makeCaches() *caches {
	return &caches{
		str:       make(map[string]cacheString),
		time:      make(map[string]cacheTime),
		cancelCh:  make(chan struct{}),
		mailCh:    make(chan cacheMail),
		requestCh: make(chan cacheRequest),
		sweepCh:   make(chan []mailID),
	}
}

func (c *caches) initCachesString(name string) {
	c.str[name] = make(map[string][]mailID)
}

func (c *caches) initCachesTime(name string) {
	c.time[name] = make(map[time.Time][]mailID)
}

func (m *cacheMail) getHeader(h string) (string, []string) {
	for header, v := range m.headers {
		hl := strings.ToLower(header)
		if hl == h {
			return header, v
		}
	}

	return "", nil
}

func (c *caches) indexMail(m cacheMail) {
	for name := range c.str {
		headerKey, val := m.getHeader(name)
		if val == nil {
			continue
		}

		switch name {
		case "to", "from":
			if addresses, err := m.headers.AddressList(headerKey); err == nil {
				for _, a := range addresses {
					c.addString(name, strings.ToLower(a.Address), m.id)
				}
			} else {
				log.Print(err)
			}
		default:
			for _, v := range val {
				c.addString(name, v, m.id)
			}
		}
	}

	for name := range c.time {
		// XXX: Only support the "date" header for now.
		if name != "date" {
			continue
		}

		if date, err := m.headers.Date(); err == nil {
			c.addTime(name, date, m.id)
		} else {
			log.Print(err)
		}
	}
}

func (c *caches) addString(name string, key string, value mailID) {
	if _, found := c.str[name][key]; !found {
		c.str[name][key] = make([]mailID, 0)
	}

	c.str[name][key] = append(c.str[name][key], value)
}

func (c *caches) getString(name string, key string) []mailID {
	return c.str[name][key]
}

func (c *caches) addTime(name string, key time.Time, value mailID) {
	if _, found := c.time[name][key]; !found {
		c.time[name][key] = make([]mailID, 0)
	}

	c.time[name][key] = append(c.time[name][key], value)
}

func (c *caches) getKeysString(name string) []string {
	keys := make([]string, 0)

	for k := range c.str[name] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (c *caches) getKeysTime(name string) []time.Time {
	keys := make([]time.Time, 0)

	for k := range c.time[name] {
		keys = append(keys, k)
	}

	sort.Sort(timeSlice(keys))

	return keys
}

func slicePresent(m mailID, elements []mailID) bool {
	for _, e := range elements {
		if e == m {
			return true
		}
	}
	return false
}

func sliceDiff(a, b []mailID) []mailID {
	r := make([]mailID, 0)

	for _, e := range a {
		if !slicePresent(e, b) {
			r = append(r, e)
		}
	}

	return r
}

func (c *caches) sweepCacheStr(name string, removedIDs []mailID) {
	for k := range c.str[name] {
		sort.Sort(mailIDSlice(c.str[name][k]))
		c.str[name][k] = sliceDiff(c.str[name][k], removedIDs)
	}
}

func (c *caches) sweepCacheTime(name string, removedIDs []mailID) {
	for k := range c.str[name] {
		sort.Sort(mailIDSlice(c.str[name][k]))
		c.str[name][k] = sliceDiff(c.str[name][k], removedIDs)
	}
}

func (c *caches) sweep(removedIDs []mailID) {
	for k := range c.str {
		c.sweepCacheStr(k, removedIDs)
	}

	for k := range c.time {
		c.sweepCacheTime(k, removedIDs)
	}
}

func (c *caches) cancel() {
	c.cancelCh <- struct{}{}
}

func (c *caches) request(r cacheRequest) {
	c.requestCh <- r
}

func (c *caches) index(id mailID, headers mail.Header) {
	c.mailCh <- cacheMail{id: id, headers: headers}
}

func (c *caches) run() {
	for {
		select {
		case <-c.cancelCh:
			return
		case m := <-c.mailCh:
			// Index this mail
			c.indexMail(m)
		case r := <-c.requestCh:
			// Send back []mailID
			if r.str != "" {
				if cache, found := c.str[r.name]; found {
					ids := cache[r.str]
					result := make([]mailID, len(ids))
					copy(result, ids)
					r.data <- result
					continue
				}
			}

			r.data <- nil
		case mailIDs := <-c.sweepCh:
			c.sweep(mailIDs)
		}
	}
}
