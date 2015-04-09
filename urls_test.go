package main

import (
	"testing"
)

func TestUrlLatest(t *testing.T) {
	cr, err := makeCacheRequest("/latest/2")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 0 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 2 {
		t.Error("Index is not set correctly")
	}
	if cr.oldest == true {
		t.Error("Oldest should not be set")
	}
}

func TestUrlOldest(t *testing.T) {
	cr, err := makeCacheRequest("/oldest/17")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 0 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 17 {
		t.Error("Index is not set correctly")
	}
	if cr.oldest != true {
		t.Error("Oldest should be set")
	}
}

func TestUrlLimitRange(t *testing.T) {
	cr, err := makeCacheRequest("/oldest/17-20")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 4 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 17 {
		t.Error("Index is not set correctly")
	}
	if cr.oldest != true {
		t.Error("Oldest should be set")
	}
}

func TestUrlLimit(t *testing.T) {
	cr, err := makeCacheRequest("/oldest/17,3")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 3 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 17 {
		t.Error("Index is not set correctly")
	}
	if cr.oldest != true {
		t.Error("Oldest should be set")
	}
}

func TestUrlFrom(t *testing.T) {
	cr, err := makeCacheRequest("/from/dullgiulio@gmail.com/oldest/17")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 0 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 17 {
		t.Error("Index is not set correctly")
	}
	if cr.oldest != true {
		t.Error("Oldest should be set")
	}
	if cr.header != "from" {
		t.Error("Header is not set correctly")
	}
	if cr.value != "dullgiulio@gmail.com" {
		t.Error("Value is not set correctly")
	}
}

func TestRedirects(t *testing.T) {
	urls := map[string]string{
		"/":                                 "/latest/0",
		"/oldest/":                          "/oldest/0",
		"/latest/":                          "/latest/0",
		"/from/dullgiulio@gmail.com/":       "/from/dullgiulio@gmail.com/latest/0",
		"/from/dullgiulio@gmail.com/latest": "/from/dullgiulio@gmail.com/latest/0",
		"/from/dullgiulio@gmail.com/oldest": "/from/dullgiulio@gmail.com/oldest/0",
	}

	for url, expected := range urls {
		var ok bool
		_, err := makeCacheRequest(url)
		if err == nil {
			t.Error("Expected error, not null")
			continue
		}
		if err, ok = err.(errorRedirect); !ok {
			t.Error("Expected a redirect error for ", url)
			continue
		}
		if err.Error() != expected {
			t.Error("Got ", err, " expected ", expected)
		}
	}
}

func TestUrlInvalid(t *testing.T) {
	urls := []string{
		"/oldest/ok",
		"/latest/11-",
		"/latest/13-11",
		"/oldest/,11",
		"/latest/12,",
		"/oldest/,0",
		"/oldest/2,-1",
	}

	for i := range urls {
		_, err := makeCacheRequest(urls[i])
		if err == nil {
			t.Error("Expected error")
		}
	}
}
