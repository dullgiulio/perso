// +build !linux

package main

type event struct{}

func (e *event) handle(c *crawler) {
}

type notify struct{}

func newNotify(dir string) (*notify, error) {
	return nil, nil
}

func (i *notify) eventsChannel() chan *event {
	return nil
}

func (i *notify) errorsChannel() chan error {
	return nil
}
