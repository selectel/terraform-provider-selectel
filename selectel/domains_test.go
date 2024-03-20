package selectel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainsV1ParseDomainRecordIDsPair(t *testing.T) {
	tableTest := []struct {
		input            string
		expectedDomainID int
		expectedRecordID int
		err              error
	}{
		{
			input:            "123/321",
			expectedDomainID: 123,
			expectedRecordID: 321,
		},
		{
			input:            "321",
			expectedDomainID: -1,
			expectedRecordID: -1,
			err:              errParseDomainsDomainRecordV1IDsPair("321"),
		},
		{
			input:            "",
			expectedDomainID: -1,
			expectedRecordID: -1,
			err:              errParseDomainsDomainRecordV1IDsPair(""),
		},
		{
			input:            "invalid/123",
			expectedDomainID: -1,
			expectedRecordID: -1,
			err:              errParseDomainsDomainV1ID("invalid"),
		},
		{
			input:            "123/invalid",
			expectedDomainID: -1,
			expectedRecordID: -1,
			err:              errParseDomainsRecordV1ID("invalid"),
		},
	}

	for _, test := range tableTest {
		gotDomainID, gotRecordID, err := domainsV1ParseDomainRecordIDsPair(test.input)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expectedDomainID, gotDomainID)
		assert.Equal(t, test.expectedRecordID, gotRecordID)
	}
}

func TestGetIntPtrOrNil(_ *testing.T) {
	tableTest := []struct {
		input    interface{}
		expected *int
	}{
		{
			input:    123,
			expected: intPtr(123),
		},
		{
			input:    nil,
			expected: nil,
		},
	}

	for _, test := range tableTest {
		getIntPtrOrNil(test.input)
	}
}
