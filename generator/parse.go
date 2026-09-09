package main

import (
	"net/netip"
	"net/url"
	"strings"
)

func parseLines(lines []string, format Format, behavior Behavior) []element {
	var elems []element

	for _, line := range lines {
		line, _, _ = strings.Cut(line, " ")
		line = strings.ToLower(line)

		if strings.HasPrefix(line, "#") {
			continue
		}

		switch format {
		case fmtRaw:
		case fmtURL:
			u, err := url.Parse(line)
			if err != nil || u.Hostname() == "" {
				continue
			}
			line = u.Hostname()
		}

		isSuffix := false

		addr, err := netip.ParseAddr(line)
		isIP := err == nil
		if isIP {
			if addr.Zone() != "" {
				continue
			}
		}

		switch behavior {
		case bhvDomain:
			if isIP {
				continue
			}
			line = strings.TrimRight(line, ".")
			isSuffix = strings.Count(line, ".") <= 1

			invalid := strings.ContainsFunc(line, func(s rune) bool {
				switch {
				case 'a' <= s && s <= 'z':
				case '0' <= s && s <= '9':
				case s == '-':
				case s == '.':
				default:
					return true
				}
				return false
			})
			if invalid {
				continue
			}
		case bhvIPCIDR:
			if !isIP { // TODO: Resolve
				continue
			}
			addr = addr.Unmap()
			line = addr.String()

			if addr.Is4() {
				line += "/32"
			} else {
				line += "/128"
			}
		}

		if line == "" {
			continue
		}

		elems = append(elems, element{value: line, isSuffix: isSuffix})
	}

	return elems
}
