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

type cacheString map[string]mailFiles

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
	files        map[mailFile]*cacheFile
	str          map[string]cacheString
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
	data     chan []mailFile
}

type cacheMail struct {
	id      mailFile
	headers mail.Header
}

func newCacheRequest() *cacheRequest {
	return &cacheRequest{
		data: make(chan []mailFile),
	}
}

func newCaches(root string) *caches {
	return &caches{
		root:         root,
		files:        make(map[mailFile]*cacheFile),
		str:          make(map[string]cacheString),
		cancelCh:     make(chan struct{}),
		requestCh:    make(chan cacheRequest),
		walkInterval: time.Second,
	}
}

func (c *caches) initCachesString(name string) {
	c.str[name] = make(map[string]mailFiles)
}

func (c *caches) indexMail(id mailFile, headers ciHeader) {
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
}

func (c *caches) addString(name string, key string, value mailFile) {
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

func (c *caches) sweepCacheStr(name string, removedIDs mailFiles) {
	for k := range c.str[name] {
		sort.Sort(c.str[name][k])
		c.str[name][k] = sliceDiff(c.str[name][k], removedIDs)
	}
}

func (c *caches) sweepCacheTime(name string, removedIDs mailFiles) {
	for k := range c.str[name] {
		sort.Sort(c.str[name][k])
		c.str[name][k] = sliceDiff(c.str[name][k], removedIDs)
	}
}

func (c *caches) sweep(removedIDs mailFiles) {
	for k := range c.str {
		c.sweepCacheStr(k, removedIDs)
	}
}

func (c *caches) loadMail(id mailFile) (*mail.Message, error) {
	mailfile := id.String()
	reader, err := os.Open(mailfile)
	if err != nil {
		return nil, err
	}

	return mail.ReadMessage(reader)
}

func (c *caches) updateFile(file mailFile, info os.FileInfo) {
	c.files[file].status = cacheFileStatusUpdated
	c.files[file].info = info
}

func (c *caches) addFile(file mailFile, info os.FileInfo) {
	c.files[file] = &cacheFile{
		status: cacheFileStatusAdded,
		info:   info,
	}

	msg, err := c.loadMail(file)
	if err != nil {
		log.Print(err)
	}

	if date, err := msg.Header.Date(); err == nil {
		file.date = date
	}

	// Index this entry
	c.indexMail(file, ciHeader(msg.Header))
}

func (c *caches) unchangedFile(file mailFile, info os.FileInfo) {
	c.files[file].status = cacheFileStatusUnchanged
	c.files[file].info = info
}

func (c *caches) addOrUpdateFile(file mailFile, finfo os.FileInfo) {
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

func (c *caches) sweepByStatus(status cacheFileStatus) (mailFiles, []os.FileInfo) {
	removedIDs := newMailFiles()
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
			if file, err := makeMailFile(path); err == nil {
				c.addOrUpdateFile(file, f)
			} else {
				log.Print(err)
			}
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
			// Send back []mailFile
			if r.str != "" {
				if cache, found := c.str[r.name]; found {
					ids := cache[r.str]
					result := make([]mailFile, len(ids))
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
