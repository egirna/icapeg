package readValues

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ReadValuesInt is used to get the int value of from toml or from env vars
//if it found the value of var in the toml value starts with "$_", it calls ReadIntFromEnv which
//retrieves th e value from env vars of the machine
func ReadValuesInt(varName string) int {

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var result int
	tempName := viper.GetString(varName)
	if strings.Index(tempName, "$_") == 0 {
		result = ReadIntFromEnv(tempName[2:len(tempName)])
	} else {
		if !viper.IsSet(varName) {
			fmt.Println(varName + " doesn't exist in config.go file")
			os.Exit(1)
		}
		result = viper.GetInt(varName)
	}
	return result
}

// ReadValuesString is used to get the string value of from toml or from env vars
//if it found the value of var in the toml value starts with "$_", it calls ReadIntFromEnv which
//retrieves th e value from env vars of the machine
func ReadValuesString(varName string) string {

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var result string
	tempName := viper.GetString(varName)
	if strings.Index(tempName, "$_") == 0 {
		result = ReadStringFromEnv(tempName[2:len(tempName)])
	} else {
		if !viper.IsSet(varName) {
			fmt.Println(varName + " doesn't exist in config.go file")
			os.Exit(1)
		}
		result = viper.GetString(varName)
	}
	return result
}

// ReadValuesBool is used to get the bool value of from toml or from env vars
//if it found the value of var in the toml value starts with "$_", it calls ReadIntFromEnv which
//retrieves th e value from env vars of the machine
func ReadValuesBool(varName string) bool {

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var result bool
	tempName := viper.GetString(varName)
	if strings.Index(tempName, "$_") == 0 {
		result = ReadBoolFromEnv(tempName[2:len(tempName)])
	} else {
		if !viper.IsSet(varName) {
			fmt.Println(varName + " doesn't exist in config.go file")
			os.Exit(1)
		}
		result = viper.GetBool(varName)
	}
	return result
}

// ReadValuesDuration is used to get the time.duration value of from toml or from env vars
//if it found the value of var in the toml value starts with "$_", it calls ReadIntFromEnv which
//retrieves th e value from env vars of the machine
func ReadValuesDuration(varName string) time.Duration {

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var result time.Duration
	tempName := viper.GetString(varName)
	if strings.Index(tempName, "$_") == 0 {
		result = ReadDurationFromEnv(tempName[2:len(tempName)])
	} else {
		if !viper.IsSet(varName) {
			fmt.Println(varName + " doesn't exist in config.go file")
			os.Exit(1)
		}
		result = viper.GetDuration(varName)
	}
	return result
}

// ReadValuesSlice is used to get the string slice value of from toml or from env vars
//if it found the value of var in the toml value starts with "$_", it calls ReadIntFromEnv which
//retrieves th e value from env vars of the machine
func ReadValuesSlice(varName string) []string {

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}
	var result []string
	tempName := viper.GetString(varName)
	if strings.Index(tempName, "$_") == 0 {
		result = ReadSliceFromEnv(tempName[2:len(tempName)])
	} else {
		if !viper.IsSet(varName) {
			fmt.Println(varName + " doesn't exist in config.go file")
			os.Exit(1)
		}
		result = viper.GetStringSlice(varName)
	}
	return result
}

// IsSecExists is used to check if a section exists in config.go file or not
func IsSecExists(varName string) bool {
	return viper.IsSet(varName)
}
