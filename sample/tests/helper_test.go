package tests

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kainkent69/fasttp"
	"github.com/kainkent69/fasttp/sample/app"
)

func init() {
	fasttp.Start(app.SetupRouter())
}

// jsonMarshal is a test helper that panics on marshal errors (should never fail
// with the static test data we pass).
func jsonMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// itoa is a shorthand for strconv.Itoa.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// resetStore clears the app's in-memory store for tests that need deterministic state.
func resetStore() {
	app.ResetStore()
}

// assertMapHas asserts a key exists in a parsed JSON map.
func assertMapHas(t *testing.T, data interface{}, key string) interface{} {
	t.Helper()
	m, ok := data.(map[string]interface{})
	assert.True(t, ok, "expected map, got %T", data)
	v, exists := m[key]
	assert.True(t, exists, "expected key %q in map", key)
	return v
}

// assertMapEqual asserts a key equals a value.
func assertMapEqual(t *testing.T, data interface{}, key string, expected interface{}) {
	t.Helper()
	actual := assertMapHas(t, data, key)
	assert.Equal(t, expected, actual)
}

// assertArrayLen asserts the length of a parsed JSON array.
func assertArrayLen(t *testing.T, data interface{}, expected int) []interface{} {
	t.Helper()
	arr, ok := data.([]interface{})
	assert.True(t, ok, "expected array, got %T", data)
	assert.Len(t, arr, expected)
	return arr
}
