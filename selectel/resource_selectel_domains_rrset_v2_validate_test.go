package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTXTRecordContent_ValidQuoted(t *testing.T) {
	validContents := []string{
		`"hello"`,
		`"some txt value"`,
		`"v=spf1 include:example.com ~all"`,
		`""`,
		`"a"`,
	}
	for _, content := range validContents {
		err := validateTXTRecordContent(content)
		assert.NoError(t, err, "expected %q to be valid TXT content", content)
	}
}

func TestValidateTXTRecordContent_InvalidUnquoted(t *testing.T) {
	invalidContents := []string{
		"hello",
		"v=spf1 include:example.com ~all",
		`"missing closing`,
		`missing opening"`,
		`"`,
	}
	for _, content := range invalidContents {
		err := validateTXTRecordContent(content)
		assert.Error(t, err, "expected %q to be invalid TXT content", content)
	}
}

func TestResourceDomainsRRSetV2Schema_HasCustomizeDiff(t *testing.T) {
	r := resourceDomainsRRSetV2()
	assert.NotNil(t, r.CustomizeDiff, "resource should have CustomizeDiff set to validate TXT record content")
}
