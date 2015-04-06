package main

import (
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type cacheString map[string][]mailID
type cacheTime map[time.Time][]mailID

type cacheFileStatus uint

const (
	cacheFileStatusDeleted cacheFileStatus = iota
	cacheFileStatusAdded
	cacheFileStatusUpdated
	cacheFileStatusUnchanged
)

type cacheFile struct {
	status cacheFileStatus
	info   os.FileInfo
}

type caches struct {
	root         string
	files        map[mailID]*cacheFile
	str          map[string]cacheString
	time         map[string]cacheTime
	cancelCh     chan struct{}
	mailCh       chan cacheMail
	requestCh    chan cacheRequest
	walkInterval time.Duration
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

func newCaches(root string) *caches {
	return &caches{
		root:         root,
		files:        make(map[mailID]*cacheFile),
		str:          make(map[string]cacheString),
		time:         make(map[string]cacheTime),
		cancelCh:     make(chan struct{}),
		requestCh:    make(chan cacheRequest),
		walkInterval: time.Second,
	}
}

func (c *caches) initCachesString(name string) {
	c.str[name] = make(map[string][]mailID)
}

func (c *caches) initCachesTime(name string) {
	c.time[name] = make(map[time.Time][]mailID)
}

func (c *caches) indexMail(id mailID, headers ciHeader) {
	for name := range c.str {
		headerKey, val := headers.get(name)
		if val == nil {
			continue
		}

		switch name {
		case "to", "from":
			if addresses, err := headers.AddressList(headerKey); err == nil {
				for _, a := range addresses {
					c.addString(name, strings.ToLower(a.Address), id)
				}
			} else {
				log.Print(err)
			}
		default:
			for _, v := range val {
				c.addString(name, v, id)
			}
		}
	}

	for name := range c.time {
		// XXX: Only support the "date" header for now.
		if name != "date" {
			continue
		}

		if date, err := headers.Date(); err == nil {
			c.addTime(name, date, id)
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

func (c *caches) loadMail(id mailID) (*mail.Message, error) {
	mailfile := string(id)
	reader, err := os.Open(mailfile)
	if err != nil {
		return nil, err
	}

	return mail.ReadMessage(reader)
}

func (c *caches) updateFile(file mailID, info os.FileInfo) {
	c.files[file].status = cacheFileStatusUpdated
	c.files[file].info = info
}

func (c *caches) addFile(file mailID, info os.FileInfo) {
	c.files[file] = &cacheFile{
		status: cacheFileStatusAdded,
		info:   info,
	}

	msg, err := c.loadMail(file)
	if err != nil {
		log.Print(err)
	}

	// Index this entry
	c.indexMail(file, ciHeader(msg.Header))
}

func (c *caches) unchangedFile(file mailID, info os.FileInfo) {
	c.files[file].status = cacheFileStatusUnchanged
	c.files[file].info = info
}

func (c *caches) addOrUpdateFile(file mailID, finfo os.FileInfo) {
	if entry, ok := c.files[file]; ok {
		if entry.info.Size() != finfo.Size() ||
			entry.info.ModTime() != finfo.ModTime() {
			c.updateFile(file, finfo)
		} else {
			c.unchangedFile(file, finfo)
		}
	} else {
		c.addFile(file, finfo)
	}
}

func (c *caches) markForRemoval() {
	for key, entry := range c.files {
		entry.status = cacheFileStatusDeleted
		c.files[key] = entry
	}
}

func (c *caches) sweepByStatus(status cacheFileStatus) ([]mailID, []os.FileInfo) {
	removedIDs := make([]mailID, 0)
	removedFinfo := make([]os.FileInfo, 0)

	for key, entry := range c.files {
		if entry.status == status {
			delete(c.files, key)

			removedIDs = append(removedIDs, key)
			removedFinfo = append(removedFinfo, entry.info)
		}
	}

	c.sweep(removedIDs)
	return removedIDs, removedFinfo
}

func (c *caches) scan() {
	// Initially, set all files as to be removed
	c.markForRemoval()

	filepath.Walk(c.root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && err == nil {
			c.addOrUpdateFile(mailID(path), f)
		}
		return err
	})

	c.sweepByStatus(cacheFileStatusDeleted)
	updatedIDs, updatedFinfo := c.sweepByStatus(cacheFileStatusUpdated)

	for i := 0; i < len(updatedIDs); i++ {
		c.addFile(updatedIDs[i], updatedFinfo[i])
	}
}

func (c *caches) cancel() {
	c.cancelCh <- struct{}{}
}

func (c *caches) request(r cacheRequest) {
	c.requestCh <- r
}

func (c *caches) run(availableCh chan<- bool) {
	c.scan()
	availableCh <- true

	for {
		select {
		case <-c.cancelCh:
			return
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
		case <-time.After(c.walkInterval):
			availableCh <- false
			c.scan()
			availableCh <- true
		}
	}
}
