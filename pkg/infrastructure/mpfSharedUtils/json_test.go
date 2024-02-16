package mpfSharedUtils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadJson(t *testing.T) {
	// Create a temporary JSON file for testing
	tempFile, err := ioutil.TempFile("", "test.json")
	assert.Nil(t, err)
	defer os.Remove(tempFile.Name())

	// Define a sample JSON content
	jsonContent := `{"name": "John Doe", "age": 30}`

	// Write the JSON content to the temporary file
	err = ioutil.WriteFile(tempFile.Name(), []byte(jsonContent), 0644)
	assert.Nil(t, err)

	// Call the ReadJson function with the temporary file path
	result, err := ReadJson(tempFile.Name())

	// Assert that there is no error
	assert.Nil(t, err)

	// Assert that the result matches the expected JSON content
	expected := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonContent), &expected)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
