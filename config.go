package main

import (
	"errors"
	"flag"
	"strings"
	"time"
)

var errInvalidFlag error = errors.New("Invalid flag")

type duration time.Duration

func (d *duration) String() string {
	return time.Duration(*d).String()
}

func (d *duration) Set(s string) error {
	dur, err := time.ParseDuration(s)
	*d = duration(dur)
	return err
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(n string) error {
	if n == "" {
		return errInvalidFlag
	}

	*s = append(*s, n)
	return nil
}

type config struct {
	keys     indexKey
	listen   string
	root     string
	agent    string
	interval duration
}

func newConfig() *config {
	keys := makeIndexKeys()
	keys.add("", keyTypeNormal)
	keys.add("from", keyTypeAddr)
	keys.add("to", keyTypeAddr)

	return &config{
		keys:     keys,
		root:     ".",
		interval: duration(2 * time.Second),
	}
}

func (c *config) parseFlags() {
	headers := stringSlice(make([]string, 0))
	addrs := stringSlice(make([]string, 0))
	parts := stringSlice(make([]string, 0))
	flag.Var(&headers, "H", "Header to index as-is")
	flag.Var(&addrs, "A", "Header containing addresses (defaults to 'From' and 'To')")
	flag.Var(&parts, "P", "Header that can be matched by a substring")
	flag.Var(&c.interval, "i", "Interval between runs of the crawler")
	flag.StringVar(&c.listen, "s", "0.0.0.0:8888", "Where to listen from (default: 0.0.0.0:8888)")
	flag.StringVar(&c.agent, "a", "MAILER-DAEMON-PERSO", "What to write after 'From ' in mbox format")
	flag.Parse()

	for i := range headers {
		c.keys.add(headers[i], keyTypeNormal)
	}
	for i := range addrs {
		c.keys.add(addrs[i], keyTypeAddr)
	}
	for i := range parts {
		c.keys.add(parts[i], keyTypePart)
	}

	c.root = flag.Arg(0)
}
