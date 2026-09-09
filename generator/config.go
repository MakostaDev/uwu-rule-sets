package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"slices"

	"go.yaml.in/yaml/v3"
)

type Format string

const (
	fmtRaw Format = "raw"
	fmtURL Format = "url"
)

type Behavior string

const (
	bhvDomain Behavior = "domain"
	bhvIPCIDR Behavior = "ipcidr"
)

type Config struct {
	RuleSets []RuleSet `yaml:"rule-sets"`
}

type RuleSet struct {
	Name     string   `yaml:"name"`
	Format   Format   `yaml:"format"`
	Behavior Behavior `yaml:"behavior"`
	URLs     []string `yaml:"urls"`
}

func (rs RuleSet) validate() error {
	if rs.Name == "" {
		return errors.New("name is empty")
	}

	switch rs.Format {
	case fmtRaw, fmtURL:
	default:
		return fmt.Errorf("unknown format: %s", rs.Format)
	}

	switch rs.Behavior {
	case bhvDomain, bhvIPCIDR:
	default:
		return fmt.Errorf("unknown behavior: %s", rs.Behavior)
	}

	if len(rs.URLs) == 0 {
		return errors.New("urls is empty")
	}

	for _, s := range rs.URLs {
		u, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("parse url %q: %w", s, err)
		}
		if u.Scheme != "https" {
			return fmt.Errorf("url %q has invalid scheme %q; want https", s, u.Scheme)
		}
	}

	return nil
}

func parseConfig(raw []byte) (Config, error) {
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("parsing yaml: %w", err)
	}

	for i, rs := range c.RuleSets {
		if err := rs.validate(); err != nil {
			return Config{}, fmt.Errorf("rule-sets[%d] %q: %w", i, rs.Name, err)
		}
		slices.Sort(c.RuleSets[i].URLs)
		c.RuleSets[i].URLs = slices.Compact(c.RuleSets[i].URLs)
	}

	return c, nil
}

func readConfig(configFile string) ([]byte, error) {
	if configFile != "" {
		return os.ReadFile(configFile)
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat stdin: %w", err)
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return nil, errors.New("no config file; use -c rule-sets.yaml or pipe via stdin")
	}

	return io.ReadAll(os.Stdin)
}
