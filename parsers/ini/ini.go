package ini

import (
	"fmt"
	"gopkg.in/ini.v1"
	"strings"
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
				out[key.Name()] = key.Value()
			}
		} else {
			mpS := make(map[string]interface{}, 0)
			for _, key := range section.Keys() {
				mpS[key.Name()] = key.Value()
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
		default:
			s += fmt.Sprintf("%s = %s\n", k, v)
		}
	}

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
