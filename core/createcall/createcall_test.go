package createcall

import (
	"testing"
)

func CheckDNDStatus(t *testing.T) {

	res, err := CheckDNDStatus("+918133910729")
	if err != nil {
		t.Error("Status couldn't be recievved")
	}
	if !res {
		t.Errorf("Status should have been true")
	}

	res, err := CheckDNDStatus("8133910729")
	if err != nil {
		t.Error("Status couldn't be recievved")
	}
	if !res {
		t.Errorf("Status should have been true")
	}

	res, err = CheckDNDStatus("7002966108")
	if err != nil {
		t.Error("Status couldn't be recievved")
	}
	if res {
		t.Errorf("Status should have been false")
	}
}
