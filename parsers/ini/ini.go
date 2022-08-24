package ini

import (
	"fmt"
	"gopkg.in/ini.v1"
	"strings"
	"time"
)

type INI struct {}

func Parser() *INI {
	return &INI{}
}

func (p *INI) Unmarshal(b []byte) (map[string]interface{}, error) {
	out := make(map[string]interface{}, 0)

	iniFile, err := ini.Load(b)
	if err != nil {
		fmt.Println("ini.Load error")
		return nil, err
	}

	for _, section := range iniFile.Sections() {
		if strings.Compare(section.Name(), "DEFAULT") == 0 {
			for _, key := range section.Keys() {
				out[key.Name()] = convertKey(key)
			}
		} else {
			mpS := make(map[string]interface{}, 0)
			for _, key := range section.Keys() {
				mpS[key.Name()] = convertKey(key)
			}
			out[section.Name()] = mpS
		}
	}

	return out, nil
}

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

func convertKey(key *ini.Key) interface{} {
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
