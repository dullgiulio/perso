package main

import (
	"errors"
	"path/filepath"
	"strings"
)

type mailFile struct {
	mailbox string
	file    string
}

var errInvalidPath = errors.New("Invalid Path")

func makeMailFile(filename string) (mailFile, error) {
	parts := strings.Split(filepath.ToSlash(filename), "/")

	if len(parts) < 2 {
		return mailFile{}, errInvalidPath
	}

	mdPart := parts[len(parts)-2]
	if mdPart != "cur" && mdPart != "new" {
		return mailFile{}, errInvalidPath
	}

	return mailFile{
		mailbox: strings.Join(parts[0:len(parts)-2], "/") + "/",
		file:    strings.Join(parts[len(parts)-2:], "/"),
	}, nil
}

func (m mailFile) String() string {
	return m.mailbox + m.file
}

type mailFiles []mailFile

func newMailFiles() mailFiles {
	return make([]mailFile, 0)
}

func slicePresent(m mailFile, elements mailFiles) bool {
	for _, e := range elements {
		if e == m {
			return true
		}
	}
	return false
}

func sliceDiff(a, b mailFiles) mailFiles {
	r := make([]mailFile, 0)

	for _, e := range a {
		if !slicePresent(e, b) {
			r = append(r, e)
		}
	}

	return r
}

func (p mailFiles) Len() int {
	return len(p)
}

func (p mailFiles) Less(i, j int) bool {
	if p[i].mailbox == p[j].mailbox {
		return p[i].file < p[j].file
	}

	return p[i].mailbox < p[j].mailbox
}

func (p mailFiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
