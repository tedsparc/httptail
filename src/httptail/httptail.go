package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"crypto/tls"
	"os"
	"strings"
	"time"
)

//FUTURE: it would be awesome to rewrite this as a "Reader", so the data is treated just as a byte stream

var debug = flag.Bool("d", false, "Show debug info")
var follow = flag.Bool("f", false, "Enable tail -f style follow behavior")
var byte_count = flag.Int("c", 1024, "Byte count to retrieve initially")
var ssl_skip_verify = flag.Bool("s", false, "Skip SSL certificate verification")

func main() {
	// Get flags
	flag.Parse()

	// Get URL argument
	var url = flag.Arg(0)
	// remove this after dev
	if url == "" {
		//url = "http://tedb.us/foo.txt"
		log.Fatal("URL must be specified, e.g. httptail http://example.com/foo.txt")
	}
	if strings.Index(url, "http://") != 0 {
		url = "http://" + url
	}
	if *debug {
		log.Printf("url: %v\n", url)
	}

	var client = &http.Client{}
	if *ssl_skip_verify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
		}
		client = &http.Client{Transport: tr}
	}

	next_request_range := range_request(client, url, fmt.Sprintf("-%d", *byte_count))

	if *follow {
		for {
			time.Sleep(1 * time.Second)
			next_request_range = range_request(client, url, next_request_range)
		}
	}
}

// Convenience method to error out if err is set
func err_fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Perform an HTTP request against the URL, specifying a range; copy output to stdout
// Returns the byte range string to use for the next HTTP request's Range: header
func range_request(client *http.Client, url string, range_header string) string {
	req, err := http.NewRequest("GET", url, nil)
	err_fatal(err)
	req.Header.Set("Range", "bytes=" + range_header)
	req.Header.Set("User-Agent", "HTTPtail")

	if *debug {
		req_string, _ := httputil.DumpRequest(req, false)
		log.Print(string(req_string))
	}

	resp, err := client.Do(req)
	err_fatal(err)
	defer resp.Body.Close()

	if *debug {
		resp_string, _ := httputil.DumpResponse(resp, true)
		log.Print(string(resp_string))
	}

	if resp.StatusCode == http.StatusPartialContent {
		_, err = io.Copy(os.Stdout, resp.Body)
		err_fatal(err)

		last_byte_position, _ := parse_content_range(resp.Header.Get("Content-Range"))
		return fmt.Sprintf("%d-", last_byte_position+1)
	} else if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable && *byte_count == 0 {
		// with byte_count == 0, we'll keep requesting range "0-" repeatedly, which will never succeed
		// Make the next request with the file size as our offset
		_, length_bytes := parse_content_range(resp.Header.Get("Content-Range"))
		return fmt.Sprintf("%d-", length_bytes)
	} else if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		// FIXME: error if the size of the file on the server has shrunk since last time
		return range_header
	} else {
		log.Fatalf("Status code must be 206 Partial Content or 416 Requested Range Not Satisfiable, not %v", resp.StatusCode)
	}
	return "" // will never reach
}

// Parse the Content-Range response header to extract the last_byte_position
func parse_content_range(range_header string) (int, int) {
	//bytes 91029-91128/91129
	var first_byte_position int
	var last_byte_position int
	var length_bytes int
	_, err := fmt.Sscanf(range_header, "bytes %d-%d/%d", &first_byte_position, &last_byte_position, &length_bytes)
	if err != nil {
		_, err := fmt.Sscanf(range_header, "bytes */%d", &length_bytes)
		err_fatal(err)
	}
	return last_byte_position, length_bytes
}
