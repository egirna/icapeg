package utils

import (
	"strings"
	"testing"
)

func TestPrepareLogMsg(t *testing.T) {
	testXICAPMetadata := "example-metadata"
	testMsg := "example-message"

	expectedJSON := `{"X-ICAP-Metadata":"example-metadata","log":"example-message"}`
	expectedResult := strings.ReplaceAll(expectedJSON, `\`, "")

	result := PrepareLogMsg(testXICAPMetadata, testMsg)

	if result != expectedResult {
		t.Errorf("Expected: %s, but got: %s", expectedResult, result)
	}

	// Additional tests can be added here to cover more cases
}
