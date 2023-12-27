// Package kdl implements a koanf.Parser that parses KDL bytes as conf maps.
package kdl

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

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
}

func endsWithNonWhitespace(str string, char rune) bool {
	for i := len(str) - 1; i >= 0; i-- {
		if unicode.IsSpace(rune(str[i])) {
			continue
		}
		return rune(str[i]) == char
	}
	return false
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

func transformWrappedInput(input []byte) ([]byte, error) {
	inputStr := strings.TrimPrefix(strings.TrimSpace(string(input)), `"" `)
	splitStr := strings.Split(inputStr, "\n")
	if !strings.HasPrefix(splitStr[len(splitStr)-1], "}") {
		return []byte(inputStr + "\n"), nil
	}
	splitStr[len(splitStr)-1] = strings.TrimFunc(strings.TrimSuffix(splitStr[len(splitStr)-1], "}"), unicode.IsSpace)

	transformedFirstLine, err := transformFirstLine(splitStr[0])
	if err != nil {
		return nil, err
	}
	splitStr[0] = transformedFirstLine

	for i := 1; i < len(splitStr); i++ {
		splitStr[i] = strings.TrimPrefix(splitStr[i], "\t")
	}

	return []byte(strings.Join(splitStr, "\n")), nil
}

// Marshal marshals the given config map to KDL bytes.
func (p *KDL) Marshal(o map[string]interface{}) ([]byte, error) {
	if len(o) == 0 {
		return []byte{}, nil
	}
	wrapper := map[string]interface{}{
		"": o,
	}
	result, err := kdl.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	transformedResult, err := transformWrappedInput(result)
	if err != nil {
		return nil, err
	}
	return []byte(transformedResult), nil
}
