package readValues_test

import (
	"icapeg/readValues"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupConfig() {
	configContent := `
[int_var]
int_var = "$_TEST_INT"

[string_var]
string_var = "$_TEST_STRING"

[bool_var]
bool_var = "$_TEST_BOOL"

[duration_var]
duration_var = "$_TEST_DURATION"

[slice_var]
slice_var = "$_TEST_SLICE"

[existing_section]
existing_section = "value"
`
	ioutil.WriteFile("config_test.toml", []byte(configContent), 0644)
	viper.SetConfigFile("config_test.toml")
	viper.ReadInConfig()
}

func teardownConfig() {
	os.Remove("config_test.toml")
}

func TestReadValuesInt(t *testing.T) {
	setupConfig()
	defer teardownConfig()
	os.Setenv("TEST_INT", "42")

	result := readValues.ReadValuesInt("int_var.int_var")
	assert.Equal(t, 42, result)
}

func TestReadValuesString(t *testing.T) {
	setupConfig()
	defer teardownConfig()
	os.Setenv("TEST_STRING", "hello")

	result := readValues.ReadValuesString("string_var.string_var")
	assert.Equal(t, "hello", result)
}

func TestReadValuesBool(t *testing.T) {
	setupConfig()
	defer teardownConfig()
	os.Setenv("TEST_BOOL", "true")

	result := readValues.ReadValuesBool("bool_var.bool_var")
	assert.Equal(t, true, result)
}

func TestReadValuesDuration(t *testing.T) {
	setupConfig()
	defer teardownConfig()
	os.Setenv("TEST_DURATION", "1h")

	result := readValues.ReadValuesDuration("duration_var.duration_var")
	assert.Equal(t, time.Hour, result)
}

func TestReadValuesSlice(t *testing.T) {
	setupConfig()
	defer teardownConfig()
	os.Setenv("TEST_SLICE", `["a","b","c"]`)

	result := readValues.ReadValuesSlice("slice_var.slice_var")
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestIsSecExists(t *testing.T) {
	setupConfig()
	defer teardownConfig()

	result := readValues.IsSecExists("existing_section.existing_section")
	assert.Equal(t, true, result)

	result = readValues.IsSecExists("non_existing_section")
	assert.Equal(t, false, result)
}
