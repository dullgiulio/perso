package main

import (
	"flag"
	"fmt"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type mailFile struct {
	ID  mailID
	msg *mail.Message
}

func newMailFile(id mailID) *mailFile {
	return &mailFile{
		ID: id,
	}
}

func (m *mailFile) load() error {
	mailfile := string(m.ID)
	reader, err := os.Open(mailfile)
	if err != nil {
		return err
	}

	m.msg, err = mail.ReadMessage(reader)
	return err
}

type fileEntry struct {
	found bool
	info  os.FileInfo
}

type files struct {
	f      map[mailID]fileEntry
	l      *sync.Mutex
	caches *caches
}

func newFiles() *files {
	caches := makeCaches()
	caches.initCachesString("from")
	caches.initCachesString("to")
	caches.initCachesTime("date")
	go caches.run()

	return &files{
		f:      make(map[mailID]fileEntry),
		l:      &sync.Mutex{},
		caches: caches,
	}
}

func (f *files) _update(file mailID, info os.FileInfo) {
	// TODO: Remove this entry from the indices

	f.f[file] = fileEntry{found: true, info: info}

	// TODO: Index again this entry
}

func (f *files) _add(file mailID, info os.FileInfo) {
	f.f[file] = fileEntry{found: true, info: info}

	m := newMailFile(file)
	if err := m.load(); err != nil {
		log.Print(err)
	}

	// Index this entry
	f.caches.index(file, m.msg.Header)
}

func (f *files) _unchanged(file mailID, info os.FileInfo) {
	f.f[file] = fileEntry{found: true, info: info}
}

func (f *files) addOrUpdate(file mailID, finfo os.FileInfo) {
	f.l.Lock()
	defer f.l.Unlock()

	if entry, ok := f.f[file]; ok {
		if entry.info.Size() != finfo.Size() ||
			entry.info.ModTime() != finfo.ModTime() {
			f._update(file, finfo)
		} else {
			f._unchanged(file, finfo)
		}
	} else {
		f._add(file, finfo)
	}
}

func (f *files) markForRemoval() {
	f.l.Lock()
	defer f.l.Unlock()

	for k, _ := range f.f {
		entry := f.f[k]
		entry.found = false
		f.f[k] = entry
	}
}

func (f *files) sweepRemoved() {
	f.l.Lock()
	defer f.l.Unlock()

	for k, entry := range f.f {
		if !entry.found {
			delete(f.f, k)
		}
	}
}

type mailDir struct {
	root         string
	walkInterval time.Duration
	files        *files
	cancelCh     chan struct{}
}

func newMailDir(root string, interval time.Duration) *mailDir {
	return &mailDir{
		root:         root,
		walkInterval: interval,
		files:        newFiles(),
		cancelCh:     make(chan struct{}),
	}
}

func (m *mailDir) cancel() {
	m.cancelCh <- struct{}{}
}

func (m *mailDir) update() {
	// Initially, set all files as to be removed
	m.files.markForRemoval()

	filepath.Walk(m.root, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && err == nil {
			m.files.addOrUpdate(mailID(path), f)
		}
		return err
	})
}

func (m *mailDir) updateLoop(tick chan<- struct{}) {
	for {
		// TODO: Errors on error channel.
		m.update()

		tick <- struct{}{}

		<-time.After(m.walkInterval)
	}
}

func (m *mailDir) run(tick chan<- struct{}) {
	go m.updateLoop(tick)

	for {
		select {
		case <-m.cancelCh:
			return
			// TODO: Handle client requests here
		}
	}
}

func main() {
	flag.Parse()
	root := flag.Arg(0)

	indexing := make(chan struct{})

	md := newMailDir(root, time.Second)
	go md.run(indexing)

	<-indexing

	caches := md.files.caches
	cacheReq := cacheRequest{
		name: "to",
		str:  "dullgiulio@gmail.com",
		data: make(chan []mailID),
	}
	caches.requestCh <- cacheReq
	data := <-cacheReq.data

	fmt.Printf("%v\n", data)

	md.cancel()
}
