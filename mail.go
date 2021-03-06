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

func (i indexKey) all() []string {
	keys := make([]string, len(i))
	for k := range i {
		keys = append(keys, k)
	}
	return keys
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

func (m *mailIndexer) parse(filename string) (*mail.Message, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return mail.ReadMessage(reader)
}

func (m *mailIndexer) cacheEntries(file mailFile, msg *mail.Message) []cacheEntry {
	entries := make([]cacheEntry, 0)
	headers := ciHeader(msg.Header)

	for key, kt := range m.keys {
		headerKey, val := headers.get(key)

		switch kt {
		case keyTypeAny:
			entries = append(entries, cacheEntry{
				name: "", key: "", value: file,
			})
		case keyTypeAddr:
			addresses, err := headers.AddressList(headerKey)
			if err != nil {
				log.Print(file, ": error parsing header ", headerKey, ": ", err)
				continue
			}

			for _, a := range addresses {
				entries = append(entries, cacheEntry{
					name:  key,
					key:   strings.ToLower(a.Address),
					value: file,
				})
			}
		default:
			if val == nil && key == "" {
				entries = append(entries, cacheEntry{
					name: "", key: "", value: file,
				})
				continue
			}

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
