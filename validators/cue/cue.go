package vcue

import (
	"io/ioutil"

	"cuelang.org/go/cuego"
)

type Validator interface {
	Validate (map[string]interface{}) error
}

type VCUE struct {
	Scheme map[string]interface{}
}

func Validator(path string) *VCUE {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	cuego.Constrain(Scheme, string(data))
}

func (v *VCUE) Validate(mp map[string]interface{}) error {
	return cuego.Validate(mp)
}
