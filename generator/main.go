package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configFile := flag.String("c", "", "set configuration file")
	outputDir := flag.String("o", "", "set output directory")
	flag.Parse()

	if *outputDir == "" {
		return errors.New("output directory is required; use -o ./out")
	}

	rawConfig, err := readConfig(*configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	c, err := parseConfig(rawConfig)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	ruleSets, err := loadRuleSets(c.RuleSets)
	if err != nil {
		return fmt.Errorf("load rule-sets: %w", err)
	}

	if err = compactRuleSets(ruleSets); err != nil {
		return fmt.Errorf("compact rule-sets: %w", err)
	}

	if err = saveRuleSets(ruleSets, *outputDir); err != nil {
		return fmt.Errorf("save rule-sets: %w", err)
	}

	return nil
}
