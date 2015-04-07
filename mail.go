package main

import (
	"log"
	"net/mail"
	"os"
	"strings"
)

type mailIndexer struct {
	keys []string
}

func newMailIndexer(keys []string) *mailIndexer {
	return &mailIndexer{
		keys: keys,
	}
}

func (m *mailIndexer) parse(file mailFile) (*mail.Message, error) {
	filename := file.String()
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return mail.ReadMessage(reader)
}

func (m *mailIndexer) cacheEntries(file mailFile, msg *mail.Message) []cacheEntry {
	entries := make([]cacheEntry, 0)
	headers := ciHeader(msg.Header)

	for _, key := range m.keys {
		headerKey, val := headers.get(key)
		if val == nil {
			continue
		}

		switch key {
		case "to", "from":
			if addresses, err := headers.AddressList(headerKey); err == nil {
				for _, a := range addresses {
					entries = append(entries, cacheEntry{
						name:  key,
						key:   strings.ToLower(a.Address),
						value: file,
					})
				}
			} else {
				log.Print(file, ": error parsing header ", headerKey, ": ", err)
			}
		default:
			for _, v := range val {
				entries = append(entries, cacheEntry{
					name:  key,
					key:   v,
					value: file,
				})
			}
		}
	}

	return entries
}
