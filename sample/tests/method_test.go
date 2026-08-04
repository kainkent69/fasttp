package tests

import (
	"net/http"
	"testing"

	fasttp "github.com/kainkent69/fasttp/tests"
)

// ---------- All HTTP methods on /methods ----------

func TestMethodGET(t *testing.T) {
	fasttp.NewFuncTest(t, "GET", "/methods", nil, func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodGet,
			Status: http.StatusOK,
			JSON:   true,
		}
	})
}

func TestMethodPOST(t *testing.T) {
	fasttp.NewFuncTest(t, "POST", "/methods", []byte(`{}`), func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPost,
			Status:  http.StatusCreated,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestMethodPUT(t *testing.T) {
	fasttp.NewFuncTest(t, "PUT", "/methods", []byte(`{}`), func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPut,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestMethodPATCH(t *testing.T) {
	fasttp.NewFuncTest(t, "PATCH", "/methods", []byte(`{}`), func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method:  http.MethodPatch,
			Status:  http.StatusOK,
			JSON:    true,
			Headers: fasttp.HeaderMap{"Content-Type": "application/json"},
		}
	})
}

func TestMethodDELETE(t *testing.T) {
	fasttp.NewFuncTest(t, "DELETE", "/methods", nil, func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodDelete,
			Status: http.StatusOK,
			JSON:   true,
		}
	})
}

func TestMethodHEAD(t *testing.T) {
	fasttp.NewFuncTest(t, "HEAD", "/methods", nil, func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodHead,
			Status: http.StatusOK,
		}
	})
}

func TestMethodOPTIONS(t *testing.T) {
	fasttp.NewFuncTest(t, "OPTIONS", "/methods", nil, func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodOptions,
			Status: http.StatusNoContent,
		}
	})
}

// ---------- CORS OPTIONS ----------

func TestCORSPreflight(t *testing.T) {
	fasttp.NewFuncTest(t, "cors preflight", "/items", nil, func(tr *fasttp.TestRunner) fasttp.TestInfo {
		return fasttp.TestInfo{
			Method: http.MethodOptions,
			Status: http.StatusNoContent,
		}
	})
}
