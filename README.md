# perso - Maildir to REST

## About

Perso is a small, self-contained REST server for a Maildir mailbox.

It includes configurable header indexing and crawling.

Perso is useful when used on testing or staging sites to have a quick
means to access a mailbox remotely, for example by a integration testing
framework.

Perso can make you forget about POP/IMAP servers for simple use cases.

## Getting Started

Download the binary (see downloads below) or build it with Go:

```sh
$ go get -u github.com/dullgiulio/perso
$ cd $GOPATH/src/github.com/dullgiulio/perso
$ go install
```

After that, cd to a directory containing e-mail messages (a Maildir,
usually containing subdirectories 'cur', 'new' and 'tmp'). And, presto, you
can run "perso".

```sh
$ cd ~/Mail/ # For example
$ $GOPATH/bin/perso

```

Now perso is listening on port 8888. Just point your browser to
http://localhost:8888/ to proceed. You should be able to see your latest
email message.

A short help is also available: http://localhost:8888/help

Good, that was it. Read on if you want to know what you just did!

## How does it work?

Perso works in a very simple way: first it reads all email messages it finds,
then makes an index for each indexed header. By default, it indices 'To' and
'From'.

For each indexed header, you can go through the matching messages in
chronological order.

```
/from/hello@mysite.com/latest/0
```

The first part is the header we want to use to select (in lowercase, so that
'X-Mailer' becomes just 'x-mailer'). The second part is the value we want the
header to have (or contain, see below).

Finally, '/latest/N' will give the Nth message from the newest, while
'/older/N' will grab starting from the oldest you have.

You can specify your own headers to index when you start "presto", see the
invocation section below.

Instead of just a number, the last part of the URL can be a range (N-M) or
N,M meaning: from mail N, M messages. Ranges are inclusive (Mth mail is shown.)

The messages will be shown in 'mbox' format. In other words, full RFC 2822,
with a "From MAILER-DAEMON-PERSO..." line between each message. The content of the 'Form'
separator is configurable for easy splitting (mbox format sucks, sorry).

## Invocation

```sh
$ perso -help
Usage of perso:
  -A Header containing addresses (defaults to 'From' and 'To')
  -H Header to index as-is
  -P Header that can be matched by a substring
  -a What to write after 'From ' in mbox format
  -i Interval between runs of the crawler
  -s Where to listen from (default: 0.0.0.0:8888)
```

After all options, you can specify the directory containing your messages. If none is
specified, perso will index the current directory.

You can specify multiple headers to index (in addition to 'From' and 'To': they are
always indexed). For example, you want to make a 'permalink' to your messages:

```sh
$ perso -A Message-ID mail-directory/
```

If a header contains addresses, use '-A' instead of '-H'.

Finally, use '-F' if you want to be able to match the contents of a header with
"fuzzy search" (case-insensitive submatches).

```sh
$ perso -P User-Agent mail-directory/
```

The '-a' flag can be used to modify the 'mbox' separator line (see above).

To modify how often to check for changes inside the mail directory, use '-i':

```sh
$ perso -i 2m
```

Here for example we index the current directory and check for changes every two
minutes.

## Example setup with Postfix

In this example, we setup Postfix to always send a copy of each outgoing email to
a local mailbox. Keep in mind that this is a privacy violation (also, illegal in
many Countries) and should only be done on test machines.

To send everything to a local mailbox, add this to /etc/postfix/main.cf:

```
always_bcc = my-mailbox@localhost
```

You probably also want to catch everything and send it to some test account on
another server (main corporate mail server, for example):

Add this to main.cf:

```
recipient_canonical_maps = regexp:/etc/postfix/recipient_map.regexp
sender_canonical_maps = regexp:/etc/postfix/sender_map.regexp
local_header_rewrite_clients = static:all

virtual_mailbox_domains = test.mailbox
virtual_mailbox_base = /var/mail
virtual_mailbox_maps = regexp:/etc/postfix/maildir_map.regexp
virtual_uid_maps = static:111
virtual_gid_maps = static:222
```

Substitute 111 with the UID of the user you want to own the mail. Same goes for
the 222 as GID.

This will create a mailbox called "test.mailbox" inside /var/mail.

Let's see the contents of the mapping files:

```sh
$ cat /etc/postfix/maildir_map.regexp
/.*/    test.mailbox/
$ cat /etc/postfix/recipient_map.regexp
/(.*)@localhost/ ${1}@localhost
/^([^@]*)@(.*)$/!^(my\-name).*^ my-name@mail-server.com
$ cat /etc/postfix/sender_map.regexp
/^([^@]*)@(.*)$/!^(my\-name).*^ my-name@mail-server.com
```

After any changes, remember to restart postfix:

```sh
# service reload postfix
```

Then start perso (under a screen/tmux session is best):

```sh
$ perso -A Message-Id -i 1s /var/mail/test.mailbox/
```

And voila, it's running. New messages are detected every second.

## Example setup with Exim

Pull request welcome!

## Downloads

Available as Github releases: https://github.com/dullgiulio/perso/releases

## Bugs

Please report bugs or suggestions: https://github.com/dullgiulio/perso/issues

## References

1. Maildir: http://cr.yp.to/proto/maildir.html
2. mbox format: https://en.wikipedia.org/wiki/Mbox
