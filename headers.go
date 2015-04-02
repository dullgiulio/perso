package main

import (
	"net/mail"
	"strings"
	"time"
)

type ciHeader mail.Header

// Return real header name and value (as slice)
func (h ciHeader) get(ho string) (string, []string) {
	for header, v := range h {
		hl := strings.ToLower(header)
		if hl == ho {
			return header, v
		}
	}

	return "", nil
}

func (h ciHeader) AddressList(key string) ([]*mail.Address, error) {
	return mail.Header(h).AddressList(key)
}

func (h ciHeader) Date() (time.Time, error) {
	return mail.Header(h).Date()
}
