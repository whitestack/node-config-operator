package modules

import (
	"reflect"
	"testing"
)

func TestGetAptErrors(t *testing.T) {
	stderr := []byte(`Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
Package vim is a virtual package provided by:
  vim-nox 2:9.1.0016-1ubuntu7.1 (= 2:9.1.0016-1ubuntu7.1)
  vim-motif 2:9.1.0016-1ubuntu7.1 (= 2:9.1.0016-1ubuntu7.1)
  vim-gtk3 2:9.1.0016-1ubuntu7.1 (= 2:9.1.0016-1ubuntu7.1)
You should explicitly select one to install.

E: Version '1124.12' for 'vim' was not found
E: Unable to locate package asdfa`)

	expected := [][]byte{
		[]byte("E: Version '1124.12' for 'vim' was not found"),
		[]byte("E: Unable to locate package asdfa"),
	}

	actual, err := getAptErrors(stderr)
	if err != nil {
		t.Errorf("got error: %s", err)
		return
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected: %s, got: %s", expected, actual)
		return
	}
}
