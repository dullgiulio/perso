package main

import (
	"fmt"
)

func reverseURLs(keys indexKey) []string {
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

// TODO: Use "template" to generate a simple HTML help with reverse URLs
