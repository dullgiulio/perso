package main

import (
	"log"
	"net/mail"
	"os"
	"strings"
)

type keyType int

const (
	keyTypeNormal keyType = iota
	keyTypeAddr
	keyTypePart
	keyTypeAny
)

type indexKey map[string]keyType

func makeIndexKeys() indexKey {
	return make(map[string]keyType)
}

func (i indexKey) add(key string, kt keyType) {
	key = strings.ToLower(key)
	i[key] = kt
}

func (i indexKey) has(key string) bool {
	_, found := i[key]
	return found
}

func (i indexKey) keyType(key string) keyType {
	k, _ := i[key]
	return k
}

type mailIndexer struct {
	keys indexKey
}

func newMailIndexer(keys indexKey) *mailIndexer {
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

	for key, kt := range m.keys {
		headerKey, val := headers.get(key)
		if val == nil {
			continue
		}

		switch kt {
		case keyTypeAny:
			entries = append(entries, cacheEntry{
				name: "", key: "", value: file,
			})
		case keyTypeAddr:
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
