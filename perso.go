package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileStatus int

const (
	fileStatusNone fileStatus = iota
	fileStatusRem
	fileStatusAdd
)

type files struct {
	f map[string]fileStatus
	l *sync.Mutex
}

func newFiles() *files {
	return &files{
		f: make(map[string]fileStatus),
		l: &sync.Mutex{},
	}
}

func (f *files) _add(file string) {
	f.f[file] = fileStatusAdd
	fmt.Printf("+ %s\n", file)
}

func (f *files) _unchanged(file string) {
	f.f[file] = fileStatusNone
}

// TODO: Store and compare os.FileInfo data (file size, etc)
func (f *files) addOrUnchanged(file string, finfo os.FileInfo) {
	f.l.Lock()
	defer f.l.Unlock()

	if _, ok := f.f[file]; ok {
		f._unchanged(file)
	} else {
		f._add(file)
	}
}

func (f *files) markForRemoval() {
	f.l.Lock()
	defer f.l.Unlock()

	for k, _ := range f.f {
		f.f[k] = fileStatusRem
	}
}

func (f *files) sweepRemoved() {
	f.l.Lock()
	defer f.l.Unlock()

	for k, s := range f.f {
		if s == fileStatusRem {
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
			m.files.addOrUnchanged(path, f)
		}
		return err
	})
}

func (m *mailDir) updateLoop() {
	for {
		// TODO: Errors on error channel.
		m.update()

		<-time.After(m.walkInterval)
	}
}

func (m *mailDir) run() {
	go m.updateLoop()

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

	md := newMailDir(root, time.Second)
	go md.run()

	<-time.After(3 * time.Second)
	md.cancel()
}
