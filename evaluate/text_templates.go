package evaluate

import (
	"bytes"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"log"
	"text/template"
)

type evaluator struct {
	values          map[string]interface{}
	evaluatedValues map[string]interface{}
}

func (e *evaluator) evaluate(key string) interface{} {
	return e.evaluateInner(nil, key)
}

func (e *evaluator) evaluateInner(inEvaluation []string, key string) interface{} {
	if result, ok := e.evaluatedValues[key]; ok {
		return result
	}

	for _, inEvaluationKey := range inEvaluation {
		if inEvaluationKey == key {
			log.Printf("cycle detected: %#v\n", inEvaluationKey)
			return "<<< CYCLE-ERROR >>>"
		}
	}

	value := e.values[key]

	if valueString, isString := value.(string); isString {
		inEvaluation = append(inEvaluation, key)
		t, err := template.New(key).
			Funcs(template.FuncMap{
				"koanf": func(key string) interface{} {
					return e.evaluateInner(inEvaluation, key)
				},
			}).
			Parse(valueString)

		if err != nil {
			log.Printf("failed to evaluate %s: %s\n", valueString, err.Error())
			return valueString
		}

		var resultBuffer bytes.Buffer
		err = t.Execute(&resultBuffer, nil)

		if err != nil {
			log.Printf("failed to evaluate %s: %s\n", valueString, err.Error())
			return valueString
		}

		result := resultBuffer.String()

		e.evaluatedValues[key] = result
		return result
	} else {
		return value
	}
}

func TextTemplates(k *koanf.Koanf) *koanf.Koanf {
	evaluator := evaluator{
		values:          k.All(),
		evaluatedValues: map[string]interface{}{},
	}

	newValues := map[string]interface{}{}

	for key := range k.All() {
		newValues[key] = evaluator.evaluate(key)
	}

	kNew := koanf.New(k.Delim())
	err := kNew.Load(confmap.Provider(newValues, k.Delim()), nil)
	if err != nil {
		panic(err)
	}

	return kNew
}
