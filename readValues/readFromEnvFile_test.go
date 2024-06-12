package readValues_test

import (
	"icapeg/readValues"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadIntFromEnv(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	result := readValues.ReadIntFromEnv("TEST_INT")
	assert.Equal(t, 42, result)
}

func TestReadStringFromEnv(t *testing.T) {
	os.Setenv("TEST_STRING", "hello")
	result := readValues.ReadStringFromEnv("TEST_STRING")
	assert.Equal(t, "hello", result)
}

func TestReadBoolFromEnv(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	result := readValues.ReadBoolFromEnv("TEST_BOOL")
	assert.Equal(t, true, result)
}

func TestReadDurationFromEnv(t *testing.T) {
	os.Setenv("TEST_DURATION", "1h")
	result := readValues.ReadDurationFromEnv("TEST_DURATION")
	assert.Equal(t, time.Hour, result)
}

func TestReadSliceFromEnv(t *testing.T) {
	os.Setenv("TEST_SLICE", `["a","b","c"]`)
	result := readValues.ReadSliceFromEnv("TEST_SLICE")
	assert.Equal(t, []string{"a", "b", "c"}, result)
}
