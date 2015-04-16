// +build linux

package main

import (
	"golang.org/x/exp/inotify"
	"log"
	"os"
)

type event struct {
	e *inotify.Event
}

func (e *event) handle(c *crawler) {
	ev := e.e
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

	if ev.Mask&(inotify.IN_DELETE|inotify.IN_MOVE|inotify.IN_MOVED_FROM|inotify.IN_CLOSE_WRITE) != 0 {
		files := make([]mailFile, 1)
		files[0] = file

		c.cache.removeCh <- files
		c.remove(files)
	}

	if ev.Mask&(inotify.IN_MOVED_TO|inotify.IN_CLOSE_WRITE) != 0 {
		c.markAdded(file, info)
	}
}

type notify struct {
	events  chan *event
	watcher *inotify.Watcher
}

func newNotify(dir string) (*notify, error) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Watch(dir); err != nil {
		return nil, err
	}

	ino := &notify{
		events:  make(chan *event),
		watcher: watcher,
	}

	go func() {
		for e := range watcher.Event {
			ino.events <- &event{e: e}
		}
	}()

	return ino, nil
}

func (i *notify) eventsChannel() chan *event {
	return i.events
}

func (i *notify) errorsChannel() chan error {
	return i.watcher.Error
}
