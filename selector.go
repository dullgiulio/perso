package main

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidURL = errors.New("Invalid URL")

func makeCacheListRequest(url string) (*cacheListRequest, error) {
	url = strings.Trim(url, "/")

	if strings.ContainsRune(url, '/') {
		return nil, errInvalidURL
	}

	cr := newCacheListRequest()
	cr.header = url
	return cr, nil
}

type selector string

func (sel selector) parseRange(cr *cacheRequest) error {
	s := string(sel)
	rangeParts := strings.SplitN(s, "-", 2)

	if len(rangeParts) < 2 {
		return errInvalidURL
	}

	var (
		index int64
		err   error
	)
	index, err = strconv.ParseInt(rangeParts[0], 10, 32)
	if err != nil {
		return err
	}
	cr.index = int(index)
	if cr.index < 0 {
		return errInvalidURL
	}

	index, err = strconv.ParseInt(rangeParts[1], 10, 32)
	if err != nil {
		return err
	}

	if int(index) < cr.index {
		return errInvalidURL
	}
	// Include last message in range
	cr.limit = int(index) - cr.index + 1
	if cr.limit < 0 {
		return errInvalidURL
	}
	return nil
}

func (sel selector) parseIndex(cr *cacheRequest) error {
	s := string(sel)
	rangeParts := strings.SplitN(s, ",", 2)

	if len(rangeParts) < 2 {
		return errInvalidURL
	}

	var (
		index, limit int64
		err          error
	)
	index, err = strconv.ParseInt(rangeParts[0], 10, 32)
	if err != nil {
		return err
	}
	cr.index = int(index)
	if cr.index < 0 {
		return errInvalidURL
	}

	limit, err = strconv.ParseInt(rangeParts[1], 10, 32)
	if err != nil {
		return err
	}
	cr.limit = int(limit)

	if cr.limit < 0 {
		return errInvalidURL
	}
	return nil
}

func (sel selector) parseNumber(cr *cacheRequest) error {
	index, err := strconv.ParseInt(string(sel), 10, 32)
	if err != nil {
		return err
	}
	cr.index = int(index)
	if index < 0 {
		return errInvalidURL
	}
	return nil
}

func (sel selector) parse(cr *cacheRequest) error {
	s := string(sel)
	if strings.ContainsRune(s, '-') {
		return sel.parseRange(cr)
	}
	if strings.ContainsRune(s, ',') {
		return sel.parseIndex(cr)
	}
	return sel.parseNumber(cr)
}
