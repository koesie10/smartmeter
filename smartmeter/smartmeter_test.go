package smartmeter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/koesie10/smartmeter/smartmeter"
)

func TestDSMR22(t *testing.T) {
	testFile(t, "dsmr22.txt")
}

func TestDSMR40(t *testing.T) {
	testFile(t, "dsmr40.txt")
}

func TestESMR50(t *testing.T) {
	testFile(t, "esmr50.txt")
}

func testFile(t *testing.T, file string) {
	f, err := os.Open(filepath.Join("test", file))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sm, err := smartmeter.New(f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sm.Read()
	if err != nil {
		t.Fatal(err)
	}
}
