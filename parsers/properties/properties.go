package properties

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/magiconair/properties"
)

type JPRTS struct{}

func Parser() *JPRTS {
	return &JPRTS{}
}

func (p *JPRTS) Unmarshal(b []byte) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	ps := properties.MustLoadString(string(b))
	for _, key := range ps.Keys() {
		out[key] = convertVal(ps.MustGet(key))
	}

	return out, nil
}

func (p *JPRTS) Marshal(o map[string]interface{}) ([]byte, error) {
	keys := make([]string, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var s string = ""

	for _, k := range keys {
		switch o[k].(type) {
		case bool:
			s += fmt.Sprintf("%s = %t\n", k, o[k])
		case int:
			s += fmt.Sprintf("%s = %d\n", k, o[k])
		case float64:
			s += fmt.Sprintf("%s = %f\n", k, o[k])
		case time.Time:
			s += fmt.Sprintf("%s = %s\n", k, o[k].(time.Time).String())
		default:
			s += fmt.Sprintf("%s = %s\n", k, o[k])
		}
	}

	return []byte(s), nil
}

func convertVal(v string) interface{} {
	b, err := strconv.ParseBool(v)
	if err == nil {
		return b
	}

	n, err := strconv.Atoi(v)
	if err == nil {
		return n
	}

	x, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return x
	}

	t, err := time.Parse(time.RFC3339, v)
	if err == nil {
		return t
	}

	return v
}
