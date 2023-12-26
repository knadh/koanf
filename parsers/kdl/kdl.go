// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	kdl "github.com/sblinch/kdl-go"
)

// KDL implements a KDL parser.
type KDL struct{}

// Parser returns a KDL Parser.
func Parser() *KDL {
	return &KDL{}
}

// Unmarshal parses the given KDL bytes.
func (p *KDL) Unmarshal(b []byte) (map[string]interface{}, error) {
	var input interface{}
	if err := kdl.Unmarshal(b, &input); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, nil
	}

	inputType := reflect.TypeOf(input)

	switch {
	case inputType == reflect.TypeOf(map[string]interface{}{}):
		return input.(map[string]interface{}), nil

	default:
		return nil, fmt.Errorf("unimplemented input type: %v", inputType)
	}

	return nil, fmt.Errorf("unimplemented")
}

func transformFirstLine(firstLine string) (string, error) {
	pairRegex := regexp.MustCompile(`(\w+)=("[^"]*"|\S+)`)
	matches := pairRegex.FindAllStringSubmatch(firstLine, -1)

	var transformedPairs []string
	for _, match := range matches {
		if len(match) < 3 {
			return "", fmt.Errorf("invalid pair format")
		}
		transformedPairs = append(transformedPairs, match[1]+" "+match[2])
	}

	return strings.Join(transformedPairs, "\n"), nil
}

func transformWrappedInput(input string) (string, error) {
	if len(input) < 6 {
		return "", fmt.Errorf("input too short to trim")
	}
	trimmedInput := input[3 : len(input)-3]

	firstLineRegex := regexp.MustCompile(`^(.*?)\n`)
	firstLineMatch := firstLineRegex.FindStringSubmatch(trimmedInput)
	if len(firstLineMatch) < 2 {
		return "", fmt.Errorf("no matching first line found")
	}

	transformedFirstLine, err := transformFirstLine(firstLineMatch[1])
	if err != nil {
		return "", err
	}

	result := firstLineRegex.ReplaceAllString(trimmedInput, transformedFirstLine+"\n")

	lines := strings.Split(result, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.TrimPrefix(lines[i], "\t")
	}

	return strings.Join(lines, "\n"), nil
}

// Marshal marshals the given config map to KDL bytes.
func (p *KDL) Marshal(o map[string]interface{}) ([]byte, error) {
	wrapper := map[string]interface{}{
		"": o,
	}
	result, err := kdl.Marshal(wrapper)

	transformedResult, err := transformWrappedInput(string(result))
	if err != nil {
		return nil, err
	}

	return []byte(transformedResult), nil
}
