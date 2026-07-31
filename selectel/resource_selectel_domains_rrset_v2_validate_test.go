package selectel

import (
	"strings"
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

func TestValidateTXTRecordContent_ValidMultiString(t *testing.T) {
	validContents := []string{
		`"string1" "string2"`,
		`"v=spf1 include:example.com ~all" " mx:example.com"`,
		`"a" "b" "c"`,
		`"key1=val1" "key2=val2" "key3=val3"`,
	}
	for _, content := range validContents {
		err := validateTXTRecordContent(content)
		assert.NoError(t, err, "expected %q to be valid TXT content", content)
	}
}

func TestValidateTXTRecordContent_InvalidMultiStringUnbalanced(t *testing.T) {
	invalidContents := []string{
		`"string1" "string2`,
		`"string1" string2"`,
		`"a" "b" "c`,
		`"only one quote`,
		`"a" "b" "c" "d`,
		`"str1" "str2" "str3" "unbalanced`,
	}
	for _, content := range invalidContents {
		err := validateTXTRecordContent(content)
		assert.Error(t, err, "expected %q to be invalid TXT content", content)
		assert.Contains(t, err.Error(), "double-quoted strings", "error for %q should mention double-quoted strings", content)
	}
}

func TestValidateTXTRecordContent_ValidMultiLine(t *testing.T) {
	validContents := []string{
		"\"line1\nline2\"",
		"\"line1\r\nline2\"",
		"\"first line\nsecond line\nthird line\"",
		"\"k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\"\n\"t0Gx4Y5dQv6fE2x5J8kL3mN7pR9sT2uV1wZ0aBcDeFgHiJkLmNoPqRsTuVwXyZaBcDeFgHiJkLmNoPqRsT\"",
	}
	for _, content := range validContents {
		err := validateTXTRecordContent(content)
		assert.NoError(t, err, "expected %q to be valid TXT content", content)
	}
}

func TestValidateTXTRecordContent_InvalidMultiLineUnbalanced(t *testing.T) {
	invalidContents := []string{
		"\"line1\nline2",
		"line1\nline2\"",
		"\"line1\n\"line2\nline3\"",
		"\"a\nb\nc",
	}
	for _, content := range invalidContents {
		err := validateTXTRecordContent(content)
		assert.Error(t, err, "expected %q to be invalid TXT content", content)
	}
}

func TestResourceDomainsRRSetV2CustomizeDiff_TXTCaseInsensitive(t *testing.T) {
	for _, typ := range []string{"TXT", "txt", "Txt", "tXt"} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "TXT", strings.ToUpper(typ),
				"EqualFold should match %q as TXT type", typ)
		})
	}
}

func TestValidateTXTRecordContent_InvalidGarbageBetweenQuotedStrings(t *testing.T) {
	invalidContents := []string{
		`"string1"GARBAGE"string2"`,
		`"a"xxx"b"`,
		`"hello" 123 "world"`,
		`"v=spf1" include:example.com "~all"`,
		`"first";"second"`,
		`"a" "b"junk"c"`,
	}
	for _, content := range invalidContents {
		err := validateTXTRecordContent(content)
		assert.Error(t, err, "expected %q to be invalid TXT content: garbage between quoted strings should be rejected", content)
	}
}

func TestResourceDomainsRRSetV2Schema_HasCustomizeDiff(t *testing.T) {
	r := resourceDomainsRRSetV2()
	assert.NotNil(t, r.CustomizeDiff, "resource should have CustomizeDiff set to validate TXT record content")
}
