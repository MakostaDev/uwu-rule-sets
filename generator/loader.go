package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func loadRuleSets(ruleSets []RuleSet) (payloads, error) {
	rsets := make(payloads)
	for _, rs := range ruleSets {
		lines, err := fetchFiles(rs.URLs)
		if err != nil {
			return nil, fmt.Errorf("fetch %s (%s): %w", rs.Name, rs.Behavior, err)
		}

		elems := parseLines(lines, rs.Format, rs.Behavior)

		if rsets[rs.Behavior] == nil {
			rsets[rs.Behavior] = make(map[string][]element)
		}
		rsets[rs.Behavior][rs.Name] = append(rsets[rs.Behavior][rs.Name], elems...)
	}

	return rsets, nil
}

func fetchFiles(urls []string) ([]string, error) {
	var lines []string

	for _, u := range urls {
		fetched, err := fetchFile(u)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", u, err)
		}

		lines = append(lines, fetched...)
	}

	return lines, nil
}

func fetchFile(url string) ([]string, error) {
	r, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", r.Status)
	}

	var lines []string
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	return lines, nil
}
