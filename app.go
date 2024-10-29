package main

import (
	"fmt"
	"sync"
	"time"
)

type ProcessStatus struct {
	startedProcessing bool
	finishedProcessed bool
	depth             int
}

type URLStatus map[string]ProcessStatus

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// This implementation doesn't do either:
	maxDepth := 4
	workerCount := 2
	var urlsStatus URLStatus = make(URLStatus)
	urlsStatus[url] = ProcessStatus{}
	var mu sync.Mutex
	done := make(chan int, workerCount)
	fetchWorker := func() {
	restart:
		mu.Lock()
		for url, status := range urlsStatus {
			if !status.startedProcessing && status.depth < maxDepth {
				status.startedProcessing = true
				mu.Unlock()
				body, urls, err := fetcher.Fetch(url)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("fetched: %s with content %q\n", url, body)
				mu.Lock()
				for _, discoveredUrl := range urls {
					_, ok := urlsStatus[discoveredUrl]
					if !ok {
						urlsStatus[discoveredUrl] = ProcessStatus{depth: status.depth + 1}
					}

				}
				mu.Unlock()
				goto restart
			}
		}
		done <- 0
	}
	for i := 0; i < workerCount; i++ {
		go fetchWorker()
	}
	workersDone := 0
	for range done {
		workersDone += 1
		if workersDone == workerCount {
			close(done)
		}
	}
}

func main() {
	Crawl("https://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	time.Sleep(500 * time.Millisecond)
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
