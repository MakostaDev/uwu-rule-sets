package main

import (
	"cmp"
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"go4.org/netipx"
)

func compactRuleSets(ruleSets payloads) error {
	for bhv, byName := range ruleSets {
		for name, elems := range byName {
			switch bhv {
			case bhvDomain:
				byName[name] = compactDomain(elems)
			case bhvIPCIDR:
				compacted, err := compactIPCIDR(elems)
				if err != nil {
					return fmt.Errorf("compact ipcidr %s: %w", name, err)
				}
				byName[name] = compacted
			default:
				return fmt.Errorf("unknown behavior: %s", bhv)
			}
		}
	}
	return nil
}

type domainKey struct {
	labels []string
	el     element
}

func (k domainKey) hasLabelPrefix(parent []string) bool {
	return len(parent) <= len(k.labels) && slices.Equal(k.labels[:len(parent)], parent)
}

func (k domainKey) rank() int {
	if k.el.isSuffix {
		return 0
	}
	return 1
}

func compactDomain(elems []element) []element {
	keys := make([]domainKey, len(elems))
	for i, e := range elems {
		reversedLabels := strings.Split(e.value, ".")
		slices.Reverse(reversedLabels)
		keys[i] = domainKey{labels: reversedLabels, el: e}
	}

	slices.SortFunc(keys, func(a, b domainKey) int {
		return cmp.Or(
			slices.Compare(a.labels, b.labels),
			cmp.Compare(a.rank(), b.rank()),
		)
	})
	keys = slices.CompactFunc(keys, func(a, b domainKey) bool {
		return slices.Equal(a.labels, b.labels)
	})

	out := make([]element, 0, len(keys))
	var parent []string

	for _, k := range keys {
		if parent != nil && k.hasLabelPrefix(parent) {
			continue
		}

		parent = nil
		out = append(out, k.el)
		if k.el.isSuffix {
			parent = k.labels
		}
	}

	return out
}

func compactIPCIDR(elems []element) ([]element, error) {
	var b netipx.IPSetBuilder
	for _, e := range elems {
		p, err := netip.ParsePrefix(e.value)
		if err != nil {
			return nil, fmt.Errorf("parse prefix %q: %w", e.value, err)
		}
		b.AddPrefix(p)
	}

	set, err := b.IPSet()
	if err != nil {
		return nil, fmt.Errorf("build ip set: %w", err)
	}

	prefixes := set.Prefixes()
	out := make([]element, 0, len(prefixes))
	for _, p := range prefixes {
		out = append(out, element{value: p.String()})
	}
	return out, nil
}
