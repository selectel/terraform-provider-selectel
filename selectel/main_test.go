package selectel

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if err := testAccDeleteBlockStorageSharedProject(); err != nil {
		fmt.Fprintf(os.Stderr, "delete Block Storage acceptance project: %v\n", err)
		code = 1
	}
	os.Exit(code)
}
