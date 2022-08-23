// Package vcue implements a koanf.Validator cue.
package vcue

import (
	"errors"
	"io/ioutil"

	"cuelang.org/go/cuego"
)

type Strategy int

// Validation strategies. "BlockAll" blocks all the data,
// if some data is invalid. "AllowValid" blocks invalid
// data, valid data remains. "Fill" fills the data.
const (
	BlockAll Strategy = iota
	AllowValid
	Fill
)

// VCUE implements a CUE validator.
type VCUE struct {
	Mode Strategy
	Scheme map[string]interface{}
}

// Validator returns a CUE Validator.
func Validator(path string, mode Strategy) *VCUE {
	var mp = make(map[string]interface{})
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	cuego.Constrain(mp, string(data))

	return &VCUE{Mode: mode, Scheme: mp}
}

// Validate validates the given data map.
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
