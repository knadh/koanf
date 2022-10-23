// Package json implements a koanf.Parser that parses INI bytes as conf maps.
package ini

import (
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

// INI implements an INI parser.
type INI struct{}

// Parser returns a new instance of the INI parser.
func Parser() *INI {
	return &INI{}
}

// Unmarshal parses the given INI bytes.
func (p *INI) Unmarshal(b []byte) (map[string]interface{}, error) {
	out := make(map[string]interface{}, 0)

	iniFile, err := ini.Load(b)
	if err != nil {
		return nil, err
	}

	for _, section := range iniFile.Sections() {
		if section.Name() == "DEFAULT" {
			for _, key := range section.Keys() {
				out[key.Name()] = getVal(key)
			}
		} else {
			mp := make(map[string]interface{})
			for _, key := range section.Keys() {
				mp[key.Name()] = getVal(key)
			}

			out[section.Name()] = mp
		}
	}

	return out, nil
}

// Marshal marshals the given config map to INI bytes.
func (p *INI) Marshal(o map[string]interface{}) ([]byte, error) {
	var s string = ""

	for k, v := range o {
		switch v.(type) {
		case map[string]interface{}:
			continue
		case bool:
			s += fmt.Sprintf("%s = %t\n", k, v)
		case int:
			s += fmt.Sprintf("%s = %d\n", k, v)
		case float64:
			s += fmt.Sprintf("%s = %f\n", k, v)
		case time.Time:
			s += fmt.Sprintf("%s = %s\n", k, v.(time.Time).String())
		default:
			s += fmt.Sprintf("%s = %s\n", k, v)
		}
	}

	for k, v := range o {
		switch v.(type) {
		case map[string]interface{}:
			s += fmt.Sprintf("[%s]\n", k)
			for kData, vData := range v.(map[string]interface{}) {
				switch vData.(type) {
				case bool:
					s += fmt.Sprintf("%s = %t\n", kData, vData)
				case int:
					s += fmt.Sprintf("%s = %d\n", kData, vData)
				case float64:
					s += fmt.Sprintf("%s = %f\n", kData, vData)
				case time.Time:
					s += fmt.Sprintf("%s = %s\n", kData, vData.(time.Time).String())
				default:
					s += fmt.Sprintf("%s = %s\n", kData, vData)
				}
			}
		}
	}

	return []byte(s), nil
}

func getVal(key *ini.Key) interface{} {
	b, err := key.Bool()
	if err == nil {
		return b
	}

	n, err := key.Int()
	if err == nil {
		return n
	}

	x, err := key.Float64()
	if err == nil {
		return x
	}

	t, err := key.Time()
	if err == nil {
		return t
	}

	s := key.String()
	return s
}
