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
}

type crawler struct {
	cache    *caches
	root     string
	files    map[mailFile]*fileMeta
	interval time.Duration
	indexer  *mailIndexer
}

func newCrawler(indexer *mailIndexer, cache *caches, root string) *crawler {
	return &crawler{
		cache:   cache,
		root:    root,
		files:   make(map[mailFile]*fileMeta),
		indexer: indexer,
	}
}

func (c *crawler) markUpdated(file mailFile, info os.FileInfo) {
	c.files[file].status = fileStatusUpdated
	c.files[file].info = info
}

func (c *crawler) markAdded(file mailFile, info os.FileInfo) {
	msg, err := c.indexer.parse(file)
	if msg == nil || err != nil {
		log.Print(file, ": error parsing ", err)
		return
	}
	// Non fatal errors
	if err != nil {
		log.Print(file, ": error parsing ", err)
	}

	if date, err := msg.Header.Date(); err == nil {
		file.date = date
	}

	c.files[file] = &fileMeta{
		status: fileStatusAdded,
		info:   info,
	}

	// Index this entry
	entries := c.indexer.cacheEntries(file, msg)
	for i := range entries {
		c.cache.addCh <- entries[i]
	}
}

func (c *crawler) markUnchanged(file mailFile) {
	c.files[file].status = fileStatusUnchanged
}

func (c *crawler) markAllDeleted() {
	for key, entry := range c.files {
		entry.status = fileStatusDeleted
		c.files[key] = entry
	}
}

func (c *crawler) markFile(file mailFile, finfo os.FileInfo) {
	entry, found := c.files[file]
	if !found {
		c.markAdded(file, finfo)
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

	for key, entry := range c.files {
		if entry.status == status {
			files = append(files, key)
			infos = append(infos, entry.info)
		}
	}

	return files, infos
}

func (c *crawler) remove(files mailFiles) {
	for _, f := range files {
		delete(c.files, f)
	}
}

func (c *crawler) walk() {
	filepath.Walk(c.root, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() || err != nil {
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
