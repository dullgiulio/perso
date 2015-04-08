package main

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidURL = errors.New("Invalid URL")

func makeCacheRequest(url string) (*cacheRequest, error) {
	cr := newCacheRequest()

	url = strings.Trim(url, "/")
	parts := strings.Split(url, "/")

	haveFilter := true

	if parts[0] == "latest" || parts[0] == "oldest" {
		haveFilter = false
	}

	if haveFilter && len(parts) < 2 {
		return nil, errInvalidURL
	}

	if haveFilter {
		cr.header = parts[0]
		cr.value = parts[1]
		parts = parts[2:]
	}

	if len(parts) == 0 {
		return cr, nil
	}

	if parts[0] == "oldest" {
		cr.oldest = true
	}

	if len(parts) < 2 {
		cr.limit = 1
		return cr, nil
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
			cr.limit = int(index) - cr.index
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

	if limit, err := strconv.ParseInt(parts[1], 10, 32); err != nil {
		return nil, err
	} else {
		cr.limit = int(limit)
		return cr, nil
	}

	return nil, errInvalidURL
}
