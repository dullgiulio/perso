package main

import (
	"testing"
)

func TestUrlLatest(t *testing.T) {
	cr, err := makeCacheRequest("/latest/1")
	if err != nil {
		t.Error("No error expected, got: ", err)
	}
	if cr.limit != 1 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 0 {
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
	if cr.limit != 17 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 0 {
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
	if cr.limit != 17 {
		t.Error("Limit is not set correctly")
	}
	if cr.index != 0 {
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
