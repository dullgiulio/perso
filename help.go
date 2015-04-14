package main

import (
	"bytes"
	"fmt"
	"io"
)

var helpPage1 string = `<h1>Perso - Maildir to REST daemon</h1>
<ul>
	<li>Available URLs:`
var helpPage2 string = `</li>
	<li>N can be: (1) a number (ex: "1", "2", "135"), (2) a range (eg: "1-5", "8-9"), (3) a number with limit (eg "1,2": from the first, two elements; "6,3": from the sixth, three elements)
	</li>
</ul>
<body>
</html>`

type help struct {
	data []byte
	keys indexKey
}

func newHelp(keys indexKey) *help {
	return &help{
		keys: keys,
	}
}

func (h *help) reverseURLs() []string {
	urls := make([]string, 0)

	urls = append(urls, "/help")

	for k, t := range h.keys {
		if k == "" {
			continue
		}

		urls = append(urls, "/"+k)

		var url string

		switch t {
		case keyTypeAny:
			continue
		case keyTypeAddr:
			url = fmt.Sprintf("/%s/EMAIL-ADDRESS", k)
		case keyTypePart:
			url = fmt.Sprintf("/%s/PARTIAL-HEADER-VALUE", k)
		default:
			if k != "" {
				url = fmt.Sprintf("/%s/FULL-HEADER-VALUE", k)
			}
		}

		if url != "" {
			urls = append(urls, url)
		}
		urls = append(urls, url+"/latest/N")
		urls = append(urls, url+"/oldest/N")
	}

	return urls
}

func (h *help) renderURLs(w io.Writer) error {
	if _, err := fmt.Fprint(w, "<ul>"); err != nil {
		return err
	}

	for _, url := range h.reverseURLs() {
		if _, err := fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, url, url); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "</ul>"); err != nil {
		return err
	}
	return nil
}

func (h *help) writeTitle(w io.Writer) error {
	if _, err := fmt.Fprint(w, "Perso - Help"); err != nil {
		return err
	}
	return nil
}

func (h *help) renderHelp(w io.Writer) error {
	if _, err := fmt.Fprint(w, helpPage1); err != nil {
		return err
	}
	if err := h.renderURLs(w); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, helpPage2); err != nil {
		return err
	}
	return nil
}

func (h *help) writeContent(w io.Writer) error {
	if h.data == nil {
		var b bytes.Buffer
		h.renderHelp(&b)
		h.data = b.Bytes()
	}

	r := bytes.NewReader(h.data)
	_, err := io.Copy(w, r)
	return err
}
