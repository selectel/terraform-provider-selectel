package selvpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	config := &Config{
		Token: "secret",
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
