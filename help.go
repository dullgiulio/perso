package main

import (
	"bytes"
	"fmt"
	"io"
)

type help struct {
	reader *bytes.Reader
}

func newHelp(keys indexKey) *help {
	h := &help{}
	h.renderHelp(keys)
	return h
}

func (h *help) write(w io.Writer) error {
	_, err := io.Copy(w, h.reader)
	return err
}

func (h *help) reverseURLs(keys indexKey) []string {
	urls := make([]string, 0)

	for k, t := range keys {
		var url string

		switch t {
		case keyTypeAny:
			continue
		case keyTypeAddr:
			url = fmt.Sprintf("/%s/EMAIL-ADDRESS", k)
		case keyTypePart:
			url = fmt.Sprintf("/%s/PARTIAL-HEADER-VALUE", k)
		default:
			url = fmt.Sprintf("/%s/FULL-HEADER-VALUE", k)
		}

		urls = append(urls, url)
		urls = append(urls, url+"/latest/N")
		urls = append(urls, url+"/oldest/N")
	}

	return urls
}

func (h *help) renderURLs(keys indexKey) []byte {
	var u bytes.Buffer

	fmt.Fprint(&u, "<ul>")

	for _, url := range h.reverseURLs(keys) {
		fmt.Fprintf(&u, `<li><a href="%s">%s</a></li>`, url, url)
	}

	fmt.Fprint(&u, "</ul>")
	return u.Bytes()
}

func (h *help) renderHelp(keys indexKey) {
	var b bytes.Buffer

	page := `<!DOCTYPE html>
<html>
<head>
	<title>Perso - Help</title>
</head>
<body>
<h1>Perso - Maildir to REST daemon</h1>
<ul>
	<li>Available URLs:
	%s
	</li>
	<li>N can be: (1) a number (ex: "1", "2", "135"), (2) a range (eg: "1-5", "8-9"), (3) a number with limit (eg "1,2": from the first, two elements; "6,3": from the sixth, three elements)
	</li>
</ul>
<body>
</html>`

	fmt.Fprintf(&b, page, h.renderURLs(keys))
	h.reader = bytes.NewReader(b.Bytes())
}
