// +build linux

package main

import (
	fsnotify "gopkg.in/fsnotify.v1"
	"log"
	"os"
)

type event fsnotify.Event

func (ev event) handle(c *crawler) {
	if ev.Name == "" {
		return
	}

	info, err := os.Stat(ev.Name)
	if err != nil || info.IsDir() {
		return
	}

	file, err := makeMailFile(ev.Name)
	if err != nil {
		log.Print(ev.Name, ": error parsing file ", err)
		return
	}

	if ev.Op&fsnotify.Remove == fsnotify.Remove ||
		ev.Op&fsnotify.Rename == fsnotify.Rename {
		files := make([]mailFile, 1)
		files[0] = file

		c.cache.removeCh <- files
		c.remove(files)
	}

	if ev.Op&fsnotify.Rename == fsnotify.Rename ||
		ev.Op&fsnotify.Write == fsnotify.Write ||
		ev.Op&fsnotify.Create == fsnotify.Create {
		c.markAdded(file, info)
	}
}

type notify struct {
	events  chan *event
	watcher *fsnotify.Watcher
}

func newNotify(dir string) (*notify, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(dir); err != nil {
		return nil, err
	}

	n := &notify{
		events:  make(chan *event),
		watcher: watcher,
	}

	// Listen to events forever, no need to close this watcher.
	go func() {
		for ev := range n.watcher.Events {
			ev := event(ev)
			n.events <- &ev
		}
	}()

	return n, nil
}

func (i *notify) eventsChannel() chan *event {
	return i.events
}

func (i *notify) errorsChannel() chan error {
	return i.watcher.Errors
}
