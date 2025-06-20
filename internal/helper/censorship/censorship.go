package censorship

import (
	"bytes"
	"regexp"
	"strings"
)

func PersonNamePartialCensor(name string) string {
	split := strings.Split(name, " ")
	buffer := &strings.Builder{}

	for i, part := range split {
		// To avoid counting by bytes, we convert
		// the text into slice of runes, so that
		// we count the characters by UTF16, not UTF8,
		// this is to account for non UTF8 character
		runes := []rune(part)
		for i, r := range runes {
			if i > 0 && i < len(runes)-1 {
				if len(runes) >= 5 && i == len(runes)/2 {
					buffer.WriteRune(r)
					continue
				}

				buffer.WriteByte('*')
				continue
			}

			buffer.WriteRune(r)
		}

		if i != len(split)-1 {
			buffer.WriteByte(' ')
		}
	}

	return buffer.String()
}

var (
	phoneRegex               = regexp.MustCompile(`^((\+)?62|0)?8(?P<censor>[0-9]+)(?P<remaining>[0-9]{2})$`)
	censorRegexGroupIndex    = phoneRegex.SubexpIndex("censor")
	remainingRegexGroupIndex = phoneRegex.SubexpIndex("remaining")
)

func PhoneNumPartialCensor(phone string) string {
	buffer := &bytes.Buffer{}
	matches := phoneRegex.FindStringSubmatch(phone)
	if matches == nil {
		return ""
	}

	buffer.WriteString("+628")

	censor := matches[censorRegexGroupIndex]
	for range len(censor) {
		buffer.WriteByte('*')
	}

	buffer.WriteString(matches[remainingRegexGroupIndex])

	return buffer.String()
}
