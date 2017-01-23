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
	header := mail.Header(h)

	// First try with stdlib implementation
	addresses, err := header.AddressList(key)
	if err == nil {
		return addresses, nil
	}

	value, found := header[key]
	if !found {
		return nil, nil
	}

	// "Forced" parsing of non-standard e-mail addresses
	return h.parseNonstandardAddressesList(strings.Join(value, " "))
}

func (h ciHeader) Date() (time.Time, error) {
	return mail.Header(h).Date()
}

// XXX: See if this part is necessary. If not, throw away
type parserStatus int

const (
	parserStatusNone parserStatus = iota
	parserStatusAddr
	parserStatusName
)

func (h ciHeader) parseNonstandardAddress(s string) (*mail.Address, error) {
	var begin, end int
	var status parserStatus

	i := 0
	for ; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			continue
		}
		if s[i] == '<' {
			continue
		}
		break
	}

	begin = i
	for ; i < len(s); i++ {
		if s[i] == '>' {
			end = i
			break
		}
		if s[i] == ' ' || s[i] == '\t' {
			end = i
			break
		}
	}

	email := s[begin:end]
	begin, end = 0, 0

	for i++; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			break
		}
	}

	if i < len(s) {
		if s[i] == '(' {
			status = parserStatusName
		}

		if status == parserStatusName {
			begin = i + 1
			for ; i < len(s); i++ {
				if s[i] == ')' {
					end = i
					break
				}
			}
		}
	}
	name := s[begin:end]

	return &mail.Address{
		Name:    name,
		Address: email,
	}, nil
}

func (h ciHeader) parseNonstandardAddressesList(s string) ([]*mail.Address, error) {
	addrs := make([]*mail.Address, 0)
	addresses := strings.Split(s, ", ")

	for a := range addresses {
		addr, err := mail.ParseAddress(addresses[a])

		if err != nil {
			addr, err = h.parseNonstandardAddress(addresses[a])
			if err != nil {
				continue
			}
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}
