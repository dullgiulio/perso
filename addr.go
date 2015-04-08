package main

import (
	"net/mail"
	"strings"
)

type parserStatus int

const (
	parserStatusNone parserStatus = iota
	parserStatusAddr
	parserStatusName
)

func parseNonstandardAddress(s string) (*mail.Address, error) {
	var begin, end int
	var status parserStatus

	i := 0
	for ; i < len(s); i++ {
		if s[i] == '<' {
			status = parserStatusAddr
			break
		}
	}

	if status == parserStatusAddr {
		status = parserStatusNone
		begin = i + 1
		for ; i < len(s); i++ {
			if s[i] == ' ' || s[i] == '\t' {
				begin = i
				continue
			}
			if s[i] == '@' {
				status = parserStatusAddr
				continue
			}
			if s[i] == '>' {
				end = i
				break
			}
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

func parseAddressesList(s string) ([]*mail.Address, error) {
	addrs := make([]*mail.Address, 0)
	addresses := strings.Split(s, ", ")

	for a := range addresses {
		addr, err := mail.ParseAddress(addresses[a])

		if err != nil {
			addr, err = parseNonstandardAddress(addresses[a])
			if err != nil {
				continue
			}
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}
