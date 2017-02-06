package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type fileStatus uint

const (
	fileStatusDeleted fileStatus = iota
	fileStatusAdded
	fileStatusUpdated
	fileStatusUnchanged
)

type fileMeta struct {
	status fileStatus
	info   os.FileInfo
	mfile  mailFile
}

type crawler struct {
	cache    *caches
	root     string
	files    map[string]*fileMeta
	interval time.Duration
	wakeup   chan struct{}
	indexer  *mailIndexer
}

func newCrawler(indexer *mailIndexer, cache *caches, root string) *crawler {
	return &crawler{
		cache:   cache,
		root:    root,
		files:   make(map[string]*fileMeta),
		wakeup:  make(chan struct{}),
		indexer: indexer,
	}
}

func (c *crawler) markUpdated(file string, info os.FileInfo) {
	c.files[file].status = fileStatusUpdated
	c.files[file].info = info
}

func (c *crawler) markAdded(mfile mailFile, info os.FileInfo) {
	file := mfile.filename()
	msg, err := c.indexer.parse(file)
	if msg == nil && err != nil {
		log.Print(file, ": error parsing ", err)
		return
	}
	// Non fatal errors
	if err != nil {
		log.Print(file, ": error parsing ", err)
	}

	if date, err := msg.Header.Date(); err == nil {
		mfile.date = date
	}

	c.files[file] = &fileMeta{
		status: fileStatusAdded,
		info:   info,
		mfile:  mfile,
	}

	// Index this entry
	entries := c.indexer.cacheEntries(mfile, msg)
	for i := range entries {
		c.cache.addCh <- entries[i]
	}
}

func (c *crawler) markUnchanged(file string) {
	c.files[file].status = fileStatusUnchanged
}

func (c *crawler) markDeleted(file string) {
	c.files[file].status = fileStatusDeleted
}

func (c *crawler) markAllDeleted() {
	for key, entry := range c.files {
		entry.status = fileStatusDeleted
		c.files[key] = entry
	}
}

func (c *crawler) markFile(mfile mailFile, finfo os.FileInfo) {
	file := mfile.filename()
	entry, found := c.files[file]
	if !found {
		c.markAdded(mfile, finfo)
		return
	}

	if entry.info.Size() != finfo.Size() ||
		entry.info.ModTime() != finfo.ModTime() {
		c.markUpdated(file, finfo)
		return
	}

	c.markUnchanged(file)
}

func (c *crawler) filesByStatus(status fileStatus) (mailFiles, []os.FileInfo) {
	files := newMailFiles()
	infos := make([]os.FileInfo, 0)

	for _, entry := range c.files {
		if entry.status == status {
			files = append(files, entry.mfile)
			infos = append(infos, entry.info)
		}
	}

	return files, infos
}

func (c *crawler) remove(files mailFiles) {
	for _, f := range files {
		delete(c.files, f.filename())
	}
}

func (c *crawler) walk() {
	filepath.Walk(c.root, func(path string, f os.FileInfo, err error) error {
		if err != nil || f.IsDir() {
			return err
		}

		file, err := makeMailFile(path)
		if err != nil {
			log.Print(path, ": error parsing file ", err)
			return nil
		}

		c.markFile(file, f)
		return err
	})
}

func (c *crawler) scan() {
	// Initially, set all files as to be removed
	c.markAllDeleted()

	c.walk()

	// Remove removed files
	filesDel, _ := c.filesByStatus(fileStatusDeleted)
	c.cache.removeCh <- filesDel
	c.remove(filesDel)

	// Remove and add again updated files
	filesUp, infosUp := c.filesByStatus(fileStatusUpdated)
	c.cache.removeCh <- filesUp

	for i := 0; i < len(filesUp); i++ {
		// XXX: Probably should only mark in markAdded,
		// then select by status and add in a new function.
		c.markAdded(filesUp[i], infosUp[i])
	}
}

func (c *crawler) rescan() {
	c.wakeup <- struct{}{}
}

func (c *crawler) run(events <-chan *event, errors <-chan error, tick <-chan time.Time) {
	for {
		select {
		case e := <-events:
			e.handle(c)
		case err := <-errors:
			log.Print(err)
		case <-c.wakeup:
			c.scan()
		case <-tick:
			c.scan()
		}
	}
}
