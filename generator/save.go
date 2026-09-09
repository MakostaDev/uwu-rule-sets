package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type target string

const (
	tgtRay  target = "ray"
	tgtMeta target = "meta"
	tgtSing target = "sing"
)

var targets = []target{tgtRay, tgtMeta, tgtSing}

func saveRuleSets(ruleSets payloads, outputDir string) error {
	for _, t := range targets {
		for bhv, byName := range ruleSets {
			dir := filepath.Join(outputDir, string(t), string(bhv))
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}

			for name, elems := range byName {
				if len(elems) == 0 {
					return fmt.Errorf("rule-set %s/%s is empty", bhv, name)
				}
				if err := saveRuleSet(dir, t, bhv, name, elems); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func saveRuleSet(dir string, t target, bhv Behavior, name string, elems []element) error {
	var path string
	var data []byte
	var err error
	switch t {
	case tgtRay:
		path = filepath.Join(dir, name)
		data = generateRayData(elems, bhv)
	case tgtMeta:
		path = filepath.Join(dir, name+".list")
		data = generateMetaData(elems, bhv)
	case tgtSing:
		path = filepath.Join(dir, name+".json")
		data, err = generateSingData(elems, bhv)
		if err != nil {
			return fmt.Errorf("generate sing data: %w", err)
		}
	default:
		return fmt.Errorf("unknown target: %s", t)
	}

	if err = os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func generateRayData(elems []element, bhv Behavior) []byte {
	rayLine := func(e element) string {
		if bhv == bhvDomain && !e.isSuffix {
			return "full:" + e.value
		}
		return e.value
	}
	return elementsToBytes(elems, rayLine)
}

func generateMetaData(elems []element, bhv Behavior) []byte {
	metaLine := func(e element) string {
		if bhv == bhvDomain && e.isSuffix {
			return "+." + e.value
		}
		return e.value
	}
	return elementsToBytes(elems, metaLine)
}

func elementsToBytes(elems []element, line func(element) string) []byte {
	var b bytes.Buffer
	for _, e := range elems {
		b.WriteString(line(e))
		b.WriteByte('\n')
	}
	return b.Bytes()
}

type singRuleSet struct {
	Version int        `json:"version"`
	Rules   []singRule `json:"rules"`
}

type singRule struct {
	Domain       []string `json:"domain,omitempty"`
	DomainSuffix []string `json:"domain_suffix,omitempty"`
	IPCIDR       []string `json:"ip_cidr,omitempty"`
}

func generateSingData(elems []element, bhv Behavior) ([]byte, error) {
	var ver int
	var rule singRule

	switch bhv {
	case bhvDomain:
		for _, e := range elems {
			if e.isSuffix {
				rule.DomainSuffix = append(rule.DomainSuffix, e.value)
			} else {
				rule.Domain = append(rule.Domain, e.value)
			}
		}
		ver = 2
	case bhvIPCIDR:
		for _, e := range elems {
			rule.IPCIDR = append(rule.IPCIDR, e.value)
		}
		ver = 1
	default:
		return nil, fmt.Errorf("unknown behavior: %s", bhv)
	}

	srs := singRuleSet{Version: ver, Rules: []singRule{rule}}

	data, err := json.MarshalIndent(srs, "", "  ")
	if err != nil {
		return nil, err
	}

	return data, nil
}
