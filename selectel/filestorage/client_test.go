package filestorage

import (
	"errors"
	"testing"

	sdk "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/assert"
)

func TestAPIError(t *testing.T) {
	t.Run("microversion names the pool and the version", func(t *testing.T) {
		err := &sdk.Error{Kind: sdk.KindMicroversion, StatusCode: 406, Message: "Version 2.75 is not supported by the API."}

		diagnostic := apiError("Create file storage", "ru-9", err)

		assert.Equal(t, "Create file storage", diagnostic.Summary())
		assert.Contains(t, diagnostic.Detail(), "pool ru-9")
		assert.Contains(t, diagnostic.Detail(), "microversion 2.75")
		assert.Contains(t, diagnostic.Detail(), err.Error())
	})

	t.Run("other errors keep the API answer", func(t *testing.T) {
		err := errors.New("filestorage: forbidden: status 403")

		diagnostic := apiError("Read file storage", "ru-9", err)

		assert.Equal(t, err.Error(), diagnostic.Detail())
	})
}
