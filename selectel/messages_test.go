package selectel

import (
	"testing"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/keypairs"
	"github.com/stretchr/testify/assert"
)

const (
	testMsgObject = "keypair"
	testMsgID     = "abcaae6eee494b76af4b006b882c1926/key1"
)

var testMsgOptions = keypairs.KeypairOpts{
	Name:      "key1",
	PublicKey: "ssh-rsa AAABBBCCC user0@example.org",
	UserID:    "abcaae6eee494b76af4b006b882c1926",
	Regions:   []string{"ru-3"},
}

func TestMsgCreate(t *testing.T) {
	expected := "[DEBUG] Creating keypair with options: {Name:key1 PublicKey:ssh-rsa AAABBBCCC user0@example.org Regions:[ru-3] UserID:abcaae6eee494b76af4b006b882c1926}"

	actual := msgCreate(testMsgObject, testMsgOptions)

	assert.Equal(t, expected, actual)
}

func TestMsgGet(t *testing.T) {
	expected := "[DEBUG] Getting keypair 'abcaae6eee494b76af4b006b882c1926/key1'"

	actual := msgGet(testMsgObject, testMsgID)

	assert.Equal(t, expected, actual)
}

func TestMsgUpdate(t *testing.T) {
	expected := "[DEBUG] Updating keypair 'abcaae6eee494b76af4b006b882c1926/key1' with options: {Name:key1 PublicKey:ssh-rsa AAABBBCCC user0@example.org Regions:[ru-3] UserID:abcaae6eee494b76af4b006b882c1926}"

	actual := msgUpdate(testMsgObject, testMsgID, testMsgOptions)

	assert.Equal(t, expected, actual)
}

func TestMsgDelete(t *testing.T) {
	expected := "[DEBUG] Deleting keypair 'abcaae6eee494b76af4b006b882c1926/key1'"

	actual := msgDelete(testMsgObject, testMsgID)

	assert.Equal(t, expected, actual)
}
