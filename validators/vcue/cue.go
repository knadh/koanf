package vcue

import (
	"errors"
	"io/ioutil"

	"cuelang.org/go/cuego"
)

type Strategy int

const (
	BlockAll Strategy = iota
	AllowValid
	Fill
)

type VCUE struct {
	Mode Strategy
	Scheme map[string]interface{}
}

func Validator(path string, mode Strategy) *VCUE {
	var mp = make(map[string]interface{})
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	cuego.Constrain(mp, string(data))

	return &VCUE{Mode: mode, Scheme: mp}
}

func (v *VCUE) Validate(mp map[string]interface{}) error {
	if v.Mode == BlockAll {
		return cuego.Validate(mp)
	}
	if v.Mode == AllowValid {
		return errors.New("AllowValid strategy is not implemented.")
	}
	if v.Mode == Fill {
		return errors.New("Fill strategy is not implemented.")
	}

	return nil
}
