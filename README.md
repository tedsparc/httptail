HTTPtail
========

HTTPtail is a command-line utility in the style of "tail", for continuously streaming the contents of a file as it is appended to, with the file residing on an HTTP server.  Locally, a system administrator might do "tail -f /var/log/foo.log", and with HTTPtail he can do "httptail -f http://example.com/log/foo.log" (if the appropriate Web server is in place).  The intended use case is streaming the contents of log files as they are appended to.

Requirements
---
HTTPtail requires an HTTP server in use that complies with the "Range" and "Content-Range" behavior specified in the HTTP RFC.  This program is tested with Nginx serving up a static file, but should also work on Apache or other standards-compliant Web servers.  Servers that do not return correct HTTP status codes 206 and 416 will cause this program to fail.

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

    # Don't show any existing content; only fetch "new" bytes that are appended
    httptail -c 0 -f http://example.com/foo.txt

Differences from tail
---
1. The main difference is "tail" is line-oriented whereas HTTPtail is byte-oriented.  Line orientation is a possible future feature.
1. HTTPtail supports the -c and -f flags from tail but does not support -b, -F, -n, or -r (see "man tail")
1. HTTPtail does not support tailing multiple URLs simultaneously

How does httptail work?
---

1. Read command line arguments for URL (required), -f (follow), and -c (byte count; optional, default 1024 bytes)
1. GET the URL with a "Range:" header specifying the last -c bytes: "bytes=-1024"
1. The HTTP response MUST be a code 206 (Partial Content); a 200 response implies the server ignored the Range header
1. Print the output to stdout
1. If -f not specified, exit
1. Sleep 1 second
1. Repeat fetch/print, with a "Range:" header starting at the last byte from the "Content-Range" response, e.g.  "bytes=100000-"
