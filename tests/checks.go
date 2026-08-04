package tests

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

// checkData reads the parsed response body from Body.parsed via Parsed().
func checkData(fn func(data interface{})) CheckFunc {
	return func(tr *T) {
		fn(tr.Parsed())
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
