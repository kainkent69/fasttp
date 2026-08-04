package fasttp

import (
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/stretchr/testify/require"
)

// ---------- format validation checks (run during Run's parse phase,
// validating against Recorder.Body) ----------

func checkValidJSON() CheckFunc {
	return func(tr *T) {
		var dummy json.RawMessage
		err := json.Unmarshal(tr.Recorder.Body.Bytes(), &dummy)
		require.NoError(tr.t, err, "response body is not valid JSON: %s", tr.Recorder.Body.String())
	}
}

func checkValidYAML() CheckFunc {
	return func(tr *T) {
		// Stub: gopkg.in/yaml.v3 not yet in go.mod.
		// When added: yaml.Unmarshal(tr.Recorder.Body.Bytes(), &dummy)
		require.NotEmpty(tr.t, tr.Recorder.Body.Bytes(), "YAML validation stub: body is empty")
	}
}

func checkValidXML() CheckFunc {
	return func(tr *T) {
		decoder := xml.NewDecoder(tr.Recorder.Body)
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			require.NoError(tr.t, err, "response body is not valid XML: %s", tr.Recorder.Body.String())
		}
	}
}

func checkValidBuffer() CheckFunc {
	return func(tr *T) {
		require.NotNil(tr.t, tr.Recorder.Body.Bytes(), "response body buffer is nil")
	}
}

// ---------- status check ----------

func checkStatus(expected int) CheckFunc {
	return func(tr *T) {
		require.Equal(tr.t, expected, tr.Recorder.Code,
			"expected status %d, got %d", expected, tr.Recorder.Code)
	}
}

// ---------- data check ----------

// checkData calls the callback with the parsed response body. The callback
// receives the actual data and returns the expected data. After the callback
// returns, the framework asserts expected == actual. No manual assert needed.
//
// For maps/slices: the callback receives a shallow copy, modifies it in-place,
// and returns it. For primitives/bytes: receives the actual, returns expected.
//
// Return nil to skip auto-assert (for manual assertion callbacks).
func checkData(fn func(data interface{}) interface{}) CheckFunc {
	return func(tr *T) {
		actual := tr.Parsed()
		expected := fn(copyForEdit(actual))
		if expected != nil {
			require.Equal(tr.t, expected, actual)
		}
	}
}

// copyForEdit returns a shallow copy suitable for in-place modification:
// maps and slices are cloned; primitives (including []byte) pass through as-is
// since the callback returns a replacement value for them.
func copyForEdit(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		cp := make(map[string]interface{}, len(val))
		for k, vv := range val {
			cp[k] = vv
		}
		return cp
	case []interface{}:
		cp := make([]interface{}, len(val))
		copy(cp, val)
		return cp
	default:
		return v
	}
}

// ---------- response header checks ----------

func checkResponseHeader(key, expected string) CheckFunc {
	return func(tr *T) {
		got := tr.Recorder.Header().Get(key)
		require.Equal(tr.t, expected, got,
			"response header %q: expected %q, got %q", key, expected, got)
	}
}

func checkResponseHeaderExists(key string) CheckFunc {
	return func(tr *T) {
		got := tr.Recorder.Header().Get(key)
		require.NotEmpty(tr.t, got,
			"response header %q: expected to be present and non-empty", key)
	}
}
