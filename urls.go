package main

import (
	"errors"
	"strconv"
	"strings"
)

type errorRedirect string

func (e errorRedirect) Error() string {
	return string(e)
}

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

func makeCacheRequest(url string) (*cacheRequest, error) {
	cr := newCacheRequest()

	url = strings.Trim(url, "/")
	parts := strings.Split(url, "/")

	haveFilter := true

	if parts[0] == "latest" || parts[0] == "oldest" {
		haveFilter = false
	}

	if haveFilter && len(parts) < 2 {
		return nil, errorRedirect("/latest/0")
	}

	if haveFilter {
		cr.header = parts[0]
		cr.value = parts[1]
		parts = parts[2:]
	}

	if len(parts) == 0 {
		return nil, errorRedirect("/" + url + "/latest/0")
	}

	if parts[0] == "oldest" {
		cr.oldest = true
	}

	if len(parts) < 2 || parts[1] == "" {
		return nil, errorRedirect("/" + url + "/0")
	}

	if strings.ContainsRune(parts[1], '-') {
		rangeParts := strings.SplitN(parts[1], "-", 2)

		if len(rangeParts) < 2 {
			return nil, errInvalidURL
		}

		if index, err := strconv.ParseInt(rangeParts[0], 10, 32); err != nil {
			return nil, err
		} else {
			cr.index = int(index)
		}
		if cr.index < 0 {
			return nil, errInvalidURL
		}

		if index, err := strconv.ParseInt(rangeParts[1], 10, 32); err != nil {
			return nil, err
		} else {
			if int(index) < cr.index {
				return nil, errInvalidURL
			}
			// Include last message in range
			cr.limit = int(index) - cr.index + 1
			if cr.limit < 0 {
				return nil, errInvalidURL
			}
		}
		return cr, nil
	}

	if strings.ContainsRune(parts[1], ',') {
		rangeParts := strings.SplitN(parts[1], ",", 2)

		if len(rangeParts) < 2 {
			return nil, errInvalidURL
		}

		if index, err := strconv.ParseInt(rangeParts[0], 10, 32); err != nil {
			return nil, err
		} else {
			cr.index = int(index)
		}
		if cr.index < 0 {
			return nil, errInvalidURL
		}

		if limit, err := strconv.ParseInt(rangeParts[1], 10, 32); err != nil {
			return nil, err
		} else {
			cr.limit = int(limit)
		}
		if cr.limit < 0 {
			return nil, errInvalidURL
		}
		return cr, nil
	}

	if index, err := strconv.ParseInt(parts[1], 10, 32); err != nil {
		return nil, err
	} else {
		cr.index = int(index)
		return cr, nil
	}

	return nil, errInvalidURL
}
