// Package ini implements a koanf.Parser that parses INI bytes as conf maps.
package ini

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type INI struct{}

var (
	assignRegex		= regexp.MustCompile(`^([^=]+)=(.*)$`)
	sectionRegex	= regexp.MustCompile(`^\[(.*)\]$`)
)

func Parser() *INI {
	return &INI{}
}

func (p *INI) Unmarshal(b []byte) (map[string]interface{}, error) {
	out := make(map[string]interface{}, 0)

	var mpS map[string]interface{} = nil
	var buf []byte = b

	var section string = ""
	var index int
	for done := false; !done; {
		index = bytes.Index(buf, []byte("\n"))
		if index == -1 {
			done = true
		}
		index++
		line := buf[:index]
		buf = buf[index:]

		line = []byte(strings.TrimSpace(string(line)))

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Skip comments
		if line[0] == ';' || line[0] == '#' {
			continue
		}

		if groups := assignRegex.FindStringSubmatch(string(line)); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)

			// first section: without header
			if strings.Compare("", section) == 0 {
				out[key] = val
			} else {
				mpS[key] = val
			}
		} else if groups := sectionRegex.FindStringSubmatch(string(line)); groups != nil {

			// first section without header is written as key: value.
			// Other sections: the header is used as a key;
			// map[string]string for a KV set of every section.
			if strings.Compare("", section) != 0 {
				out[section] = mpS
			}

			mpNewSection := make(map[string]interface{})
			mpS = mpNewSection
			section = strings.TrimSpace(groups[1])
		} else {
			return nil, errors.New("Syntax error")
		}
	}

	out[section] = mpS

	return out, nil
}

func (p *INI) Marshal(o map[string]interface{}) ([]byte, error) {
	var s string = ""

	// empty section
	for k, v := range o {
		switch v.(type) {
		case map[string]interface{}:
			continue
		default:
			s += fmt.Sprintf("%s = %s\n", k, v)
		}
	}

	// sections
	for k, v := range o {
		switch v.(type) {
		case map[string]interface{}:
			s += fmt.Sprintf("[%s]\n", k)
			for kData, vData := range v.(map[string]interface{}) {
				s += fmt.Sprintf("%s = %s\n", kData, vData)
			}
		}
	}

	return []byte(s), nil
}
