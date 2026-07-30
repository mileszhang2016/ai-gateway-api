package xreq

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeValidator struct {
	validateErr error
}

func (f *fakeValidator) Validate() error {
	return f.validateErr
}

func TestValidateDataWithValidator(t *testing.T) {
	// Plain struct without Validator interface passes through struct validation.
	plain := &struct {
		Name string `validate:"required"`
	}{Name: "ok"}
	assert.NoError(t, validateData(plain, nil))

	// Validator returning an error is wrapped as a PARAM error.
	v := &fakeValidator{validateErr: errors.New("custom error")}
	err := validateData(v, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PARAM")
	assert.Contains(t, err.Error(), "custom error")

	// Validator returning nil succeeds.
	v.validateErr = nil
	assert.NoError(t, validateData(v, nil))
}
