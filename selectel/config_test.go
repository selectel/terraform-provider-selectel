package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testsNoError := []Config{
		{Token: "secret"},
		{Token: "secret", Region: "ru-3"},
		{Token: "secret", User: "user"},
		{Token: "secret", Region: "ru-3", User: "user"},
		{Token: "secret", User: "user", Password: "password", DomainName: "domain"},
		{Token: "secret", Region: "ru-3", User: "user", Password: "password", DomainName: "domain"},
		{User: "user", Password: "password", DomainName: "domain"},
	}

	for _, tc := range testsNoError {
		assert.NoError(t, tc.Validate())
	}
}

func TestValidateNoTokenOrIncompleteCredentials(t *testing.T) {
	testsError := []*Config{
		{},
		{User: "user"},
	}

	for _, tc := range testsError {
		assert.EqualError(t, tc.Validate(), "token or credentials with domain name must be specified")
	}
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

func TestUseSelectelToken(t *testing.T) {
	type test struct {
		config   *Config
		expected bool
	}

	tests := []test{
		{&Config{}, true},
		{&Config{Token: "secret"}, true},
		{&Config{User: "user"}, true},
		{&Config{Token: "secret", User: "user"}, true},
		{&Config{User: "user", Password: "password", DomainName: "domain"}, false},
		{&Config{Token: "secret", User: "user", Password: "password", DomainName: "domain"}, false},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, tc.config.useSelectelToken())
	}
}
