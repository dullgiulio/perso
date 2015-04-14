package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
)

type templateWriter interface {
	writeTitle(w io.Writer) error
	writeContent(w io.Writer) error
}

type template struct {
	buf bytes.Buffer
}

func newTemplate() *template {
	return &template{}
}

var beforeTitle string = `<!DOCTYPE html>
<html>
<head>
    <title>`
var afterTitle string = `</title>
</head>
<body>`
var afterBody string = `<body>
</html>`

func (t *template) write(s []byte) error {
	_, err := t.buf.Write(s)
	return err
}

func (t *template) render(w io.Writer, tw templateWriter) error {
	if err := t.write([]byte(beforeTitle)); err != nil {
		return err
	}
	if err := tw.writeTitle(&t.buf); err != nil {
		return err
	}
	if err := t.write([]byte(afterTitle)); err != nil {
		return err
	}
	if err := tw.writeContent(&t.buf); err != nil {
		return err
	}

	_, err := io.Copy(w, &t.buf)
	return err
}

type listTemplate struct {
	header string
	values []string
}

func newListTemplate(header string, values []string) *listTemplate {
	return &listTemplate{header: header, values: values}
}

func (l *listTemplate) writeTitle(w io.Writer) error {
	// XXX: Escape header for HTML
	if _, err := fmt.Fprintf(w, "Perso - List for %s", l.header); err != nil {
		return err
	}
	return nil
}

func (l *listTemplate) writeContent(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "<ul>"); err != nil {
		return err
	}

	for _, val := range l.values {
		if _, err := fmt.Fprintf(w, `<li><a href="/%s/%s/latest/0">%s</a></li>`,
			url.QueryEscape(l.header), url.QueryEscape(val), val); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, "</ul>"); err != nil {
		return err
	}
	return nil
}
