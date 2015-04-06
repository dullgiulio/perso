package main

import (
	"errors"
	"path/filepath"
	"strings"
)

type mailID struct {
	mailbox string
	file    string
}

var errInvalidPath = errors.New("Invalid Path")

func makeMailID(filename string) (mailID, error) {
	parts := strings.Split(filepath.ToSlash(filename), "/")

	if len(parts) < 2 {
		return mailID{}, errInvalidPath
	}

	mdPart := parts[len(parts)-2]
	if mdPart != "cur" && mdPart != "new" {
		return mailID{}, errInvalidPath
	}

	return mailID{
		mailbox: strings.Join(parts[0:len(parts)-2], "/") + "/",
		file:    strings.Join(parts[len(parts)-2:], "/"),
	}, nil
}

func (m mailID) String() string {
	return m.mailbox + m.file
}

type mailIDSlice []mailID

func newMailIDSlice() mailIDSlice {
	return make([]mailID, 0)
}

func slicePresent(m mailID, elements mailIDSlice) bool {
	for _, e := range elements {
		if e == m {
			return true
		}
	}
	return false
}

func sliceDiff(a, b mailIDSlice) mailIDSlice {
	r := make([]mailID, 0)

	for _, e := range a {
		if !slicePresent(e, b) {
			r = append(r, e)
		}
	}

	return r
}

func (p mailIDSlice) Len() int {
	return len(p)
}

func (p mailIDSlice) Less(i, j int) bool {
	if p[i].mailbox == p[j].mailbox {
		return p[i].file < p[j].file
	}

	return p[i].mailbox < p[j].mailbox
}

func (p mailIDSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
