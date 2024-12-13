package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL      string
	Duration time.Duration
	Res      *http.Response
	Error    error
}

func (r *Result) String() string {
	var sb strings.Builder

	sb.WriteString("URL: ")
	sb.WriteString(r.URL)
	sb.WriteString("\n")

	sb.WriteString("Duration: ")
	sb.WriteString(r.Duration.String())
	sb.WriteString("\n")

	if r.Error != nil {
		sb.WriteString("Error: ")
		sb.WriteString(r.Error.Error())
		return sb.String()
	}

	sb.WriteString("Status: ")
	sb.WriteString(r.Res.Status)

	return sb.String()
}

func fetchURL(url string) *Result {
	now := time.Now()
	res, err := http.Get(url)

	r := &Result{
		URL:      url,
		Duration: time.Since(now),
	}

	if err != nil {
		r.Error = err
		return r
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("responce status is not %q. got=%d", http.StatusText(http.StatusOK), res.StatusCode)
		r.Error = err

		return r
	}

	r.Res = res

	return r
}

func main() {
	urls := os.Args[1:]
	resultsch := make(chan Result, len(urls))

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)

		go func(u string) {
			defer wg.Done()

			result := fetchURL(u)

			resultsch <- *result
		}(url)
	}

	wg.Wait()
	close(resultsch)

	for r := range resultsch {
		if r.Error != nil {
			slog.Error("fetching err", "err", r.Error)
			fmt.Println(r.String())
		} else {
			slog.Info("fetching result")
			fmt.Println(r.String())
		}
	}
}
