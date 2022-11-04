package test

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webcenter-fr/opensearch-operator/pkg/helper"
	"sigs.k8s.io/yaml"
)

func EqualFromYamlFile(t *testing.T, expectedYamlFile string, actual any) {

	if expectedYamlFile == "" {
		panic("expectedYamlFile must be provided")
	}

	// Read file
	f, err := os.ReadFile(expectedYamlFile)
	if err != nil {
		panic(err)
	}

	
	var n any

	// Create new object base from actual
	if reflect.ValueOf(actual).Kind() == reflect.Ptr {
		n = reflect.New(reflect.TypeOf(actual).Elem()).Interface()
	} else {
		n = reflect.New(reflect.TypeOf(actual)).Interface()
	}

	if err = yaml.Unmarshal(f, n); err != nil {
		panic(err)
	}


	diff := helper.Diff(n, actual) 

	if diff != "" {
		assert.Fail(t, diff)
	}
	
}