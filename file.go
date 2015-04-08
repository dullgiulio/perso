package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type mailFile struct {
	mailbox string
	file    string
	date    time.Time
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
	return m.file
}

func (m mailFile) filename() string {
	return m.mailbox + m.file
}

func (m mailFile) writeTo(w io.Writer) error {
	recipient := "MAILER-DAEMON"

	r, err := os.Open(m.filename())
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := fmt.Fprintf(w, "From %s %s\n", recipient, m.date.Format(time.UnixDate)); err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}

type mailFiles []mailFile

func newMailFiles() mailFiles {
	return make([]mailFile, 0)
}

func (ms mailFiles) writeTo(w io.Writer) {
	for _, m := range ms {
		if err := m.writeTo(w); err != nil {
			log.Print("could not write ", m, ": ", err)
		}
	}
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
	if p[i].date != p[j].date {
		return p[i].date.Before(p[j].date)
	}

	if p[i].mailbox == p[j].mailbox {
		return p[i].file < p[j].file
	}

	return p[i].mailbox < p[j].mailbox
}

func (p mailFiles) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
