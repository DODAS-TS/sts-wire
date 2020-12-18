package validator

import (
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestValidLocalDir(t *testing.T) {
	localDir := "./out_folder"
	if valid, err := LocalPath(localDir); !valid || err != nil {
		t.Logf(`localDir %s is %t != %t, error: %s`, localDir, valid, true, err)
		t.FailNow()
	}
}

func TestValidTempDir(t *testing.T) {
	localDir := "/tmp"
	if valid, err := LocalPath(localDir); !valid || err != nil {
		t.Logf(`localDir %s is %t != %t, error: %s`, localDir, valid, true, err)
		t.FailNow()
	}
}

func TestRootValidLocalDir(t *testing.T) {
	localDir := "./ ; rm -f /"
	if valid, err := LocalPath(localDir); !valid || err != nil {
		t.Fatalf(`localDir %s is %t != %t, error: %s`, localDir, valid, false, err)
	}
}
