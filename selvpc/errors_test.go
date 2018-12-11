package selvpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const errString = "got 503"

func TestErrParseProjectV2Quotas(t *testing.T) {
	err := errors.New(errString)

	expected := errors.New("got error parsing quotas: got 503")

	actual := errParseProjectV2Quotas(err)

	assert.Equal(t, expected, actual)
}

func TestErrSearchingProjectRole(t *testing.T) {
	projectID := "uuid"
	err := errors.New(errString)

	expected := errors.New("can't find role for project 'uuid': got 503")

	actual := errSearchingProjectRole(projectID, err)

	assert.Equal(t, expected, actual)
}

func TestErrCreatingObject(t *testing.T) {
	object := "some stuff"
	err := errors.New(errString)

	expected := errors.New("[DEBUG] error creating some stuff: got 503")

	actual := errCreatingObject(object, err)

	assert.Equal(t, expected, actual)
}

func TestErrUpdatingObject(t *testing.T) {
	object := "license"
	licenseID := "aaa"
	err := errors.New(errString)

	expected := errors.New("[DEBUG] error updating license 'aaa': got 503")

	actual := errUpdatingObject(object, licenseID, err)

	assert.Equal(t, expected, actual)
}

func TestErrGettingObject(t *testing.T) {
	object := "project"
	projectID := "project_1"
	err := errors.New(errString)

	expected := errors.New("[DEBUG] error getting project 'project_1': got 503")

	actual := errGettingObject(object, projectID, err)

	assert.Equal(t, expected, actual)
}

func TestErrDeletingObject(t *testing.T) {
	object := "user"
	projectID := "some_user"
	err := errors.New(errString)

	expected := errors.New("[DEBUG] error deleting user 'some_user': got 503")

	actual := errDeletingObject(object, projectID, err)

	assert.Equal(t, expected, actual)
}
