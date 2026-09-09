package main

type payloads map[Behavior]map[string][]element

type element struct {
	value    string
	isSuffix bool
}
