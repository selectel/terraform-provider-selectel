package selectel

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testErrString = "got 503"

func TestErrParsingPrefixLength(t *testing.T) {
	id := "5"
	object := "subnet"
	err := errors.New(testErrString)

	expected := "[DEBUG] can't parse prefix length from subnet '5' CIDR: got 503"

	actual := errParsingPrefixLength(object, id, err)

	assert.Equal(t, expected, actual)
}

func TestErrSettingComplexAttr(t *testing.T) {
	attr := "servers"
	err := errors.New(testErrString)

	expected := "[DEBUG] error setting servers: got 503"

	actual := errSettingComplexAttr(attr, err)

	assert.Equal(t, expected, actual)
}

func TestErrParseID(t *testing.T) {
	object := "floating IP"
	id := "c0d10656-a4ae-468e-92db-44b2032e256b"

	expected := errors.New("unable to parse floating IP ID: 'c0d10656-a4ae-468e-92db-44b2032e256b'")

	actual := errParseID(object, id)

	assert.Equal(t, expected, actual)
}

func TestErrParseProjectV2Quotas(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing quotas: got 503")

	actual := errParseProjectV2Quotas(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseCrossRegionSubnetV2Regions(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing regions: got 503")

	actual := errParseCrossRegionSubnetV2Regions(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseCrossRegionSubnetV2ProjectID(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing project ID: got 503")

	actual := errParseCrossRegionSubnetV2ProjectID(err)

	assert.Equal(t, expected, actual)
}

func TestErrSearchingProjectRole(t *testing.T) {
	projectID := "uuid"
	err := errors.New(testErrString)

	expected := errors.New("can't find role for project 'uuid': got 503")

	actual := errSearchingProjectRole(projectID, err)

	assert.Equal(t, expected, actual)
}

func TestErrSearchingKeypair(t *testing.T) {
	keypairName := "key1"
	err := errors.New(testErrString)

	expected := errors.New("can't find keypair 'key1': got 503")

	actual := errSearchingKeypair(keypairName, err)

	assert.Equal(t, expected, actual)
}

func TestErrCreatingObject(t *testing.T) {
	object := "some stuff"
	err := errors.New(testErrString)

	expected := errors.New("error creating some stuff: got 503")

	actual := errCreatingObject(object, err)

	assert.Equal(t, expected, actual)
}

func TestErrUpdatingObject(t *testing.T) {
	object := "license"
	licenseID := "aaa"
	err := errors.New(testErrString)

	expected := errors.New("error updating license 'aaa': got 503")

	actual := errUpdatingObject(object, licenseID, err)

	assert.Equal(t, expected, actual)
}

func TestErrGettingObject(t *testing.T) {
	object := "project"
	projectID := "project_1"
	err := errors.New(testErrString)

	expected := errors.New("error getting project 'project_1': got 503")

	actual := errGettingObject(object, projectID, err)

	assert.Equal(t, expected, actual)
}

func TestErrDeletingObject(t *testing.T) {
	object := "user"
	projectID := "some_user"
	err := errors.New(testErrString)

	expected := errors.New("error deleting user 'some_user': got 503")

	actual := errDeletingObject(object, projectID, err)

	assert.Equal(t, expected, actual)
}

func TestErrResourceDeprecated(t *testing.T) {
	resource := "some_vpc_object"

	expected := errors.New("some_vpc_object resource has been deprecated")

	actual := errResourceDeprecated(resource)

	assert.Equal(t, expected, actual)
}

func TestErrParseDomainsDomainV1ID(t *testing.T) {
	domainID := "badid"

	expected := fmt.Errorf("got error parsing domain ID: %s", domainID)

	actual := errParseDomainsDomainV1ID(domainID)

	assert.Equal(t, expected, actual)
}

func TestErrParseDomainsRecordV1ID(t *testing.T) {
	recordID := "badid"

	expected := fmt.Errorf("got error parsing record ID: %s", recordID)

	actual := errParseDomainsRecordV1ID(recordID)

	assert.Equal(t, expected, actual)
}

func TestErrParseDomainsDomainRecordV1IDsPair(t *testing.T) {
	idsPair := "badid/badid"

	expected := fmt.Errorf("got error parsing domain/record IDs pair: %s", idsPair)

	actual := errParseDomainsDomainRecordV1IDsPair(idsPair)

	assert.Equal(t, expected, actual)
}

func TestErrGettingObjects(t *testing.T) {
	object := "datastore-types"
	err := errors.New(testErrString)

	expected := errors.New("error getting datastore-types: got 503")

	actual := errGettingObjects(object, err)

	assert.Equal(t, expected, actual)
}

func TestErrParseDatastoreV1Flavor(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing flavor: got 503")

	actual := errParseDatastoreV1Flavor(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseDatastoreV1Pooler(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing pooler opts: got 503")

	actual := errParseDatastoreV1Pooler(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseDatastoreV1Firewall(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing firewall opts: got 503")

	actual := errParseDatastoreV1Firewall(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseDatastoreV1Resize(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing resize opts: got 503")

	actual := errParseDatastoreV1Resize(err)

	assert.Equal(t, expected, actual)
}

func TestErrParseDatastoreV1Restore(t *testing.T) {
	err := errors.New(testErrString)

	expected := errors.New("got error parsing restore opts: got 503")

	actual := errParseDatastoreV1Restore(err)

	assert.Equal(t, expected, actual)
}
