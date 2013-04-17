httptail
========

HTTPtail is a "tail" style command line utility for continuously streaming the contents of a file on an HTTP server.  The main difference is "tail" is line-oriented whereas HTTPtail is byte-oriented.  Line orientation is a possible future feature.

How to use
---
    git clone https://github.com/tedsparc/httptail.git
    cd httptail
    GOPATH=`pwd` go install httptail
    bin/httptail -c 10240 -f http://example.com/foo.txt

Invocation examples
---
    httptail http://example.com/foo.txt
        (same as "curl -r -1024 http://example.com/foo.txt" or, locally, "tail -c 1024 foo.txt")

    httptail -c 20480 http://example.com/foo.txt
        (same as "curl  -r -20480 http://example.com/foo.txt")

    httptail -c 20480 -f http://example.com/foo.txt

Differences from tail
---
1. HTTPtail supports the -c and -f flags from tail but does not support -b, -F, -n, or -r (see "man tail")
1. HTTPtail does not support tailing multiple URLs simultaneously

How does httptail work?
---

1. Read command line arguments for URL (required), -f (follow), and -c (byte count; optional, default 1024 bytes)
1. GET the URL with a "Range:" header specifying the last --bytes bytes: "bytes=-1024"
1. The HTTP response MUST be a code 206 (Partial Content); a 200 response implies the server ignored the Range header
1. Print the output to stdout
1. If -f not specified, exit
1. Sleep 1 second
1. Repeat fetch/print, with a "Range:" header starting at the last byte from the "Content-Range" response: "bytes=100000
