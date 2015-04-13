package main

import (
	"bytes"
	"fmt"
)

var page string = `<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
</head>
<body>
%s
</ul>
<body>
</html>`

func template(title, content string) []byte {
	var b bytes.Buffer

	fmt.Fprintf(&b, page, title, content)

	return b.Bytes()
}
