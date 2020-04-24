package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	config := &Config{
		Token:  "secret",
		Region: "ru-3",
	}

	err := config.Validate()

	assert.NoError(t, err)
}

func TestValidateNoToken(t *testing.T) {
	config := &Config{}

	expected := "token must be specified"

	actual := config.Validate()

	assert.EqualError(t, actual, expected)
}

func TestValidateErrRegion(t *testing.T) {
	config := &Config{
		Token:  "secret",
		Region: "unknown region",
	}

	expected := "region is invalid: unknown region"

	actual := config.Validate()

	assert.EqualError(t, actual, expected)
}
