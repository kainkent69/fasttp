// Package status provides typed HTTP status code constants for fasttp tests,
// so assertions read as status.Ok / status.NotFound instead of bare numbers.
package status

import "net/http"

const (
	Ok        = 200
	Created   = 201
	Accepted  = 202
	NoContent = 204
	Moved     = 301
	Found     = 302

	BadRequest       = 400
	Unauthorized     = 401
	Forbidden        = 403
	NotFound         = 404
	MethodNotAllowed = 405
	Conflict         = 409
	Unprocessable    = 422
	TooMany          = 429

	Internal           = 500
	ServiceUnavailable = 503
)

// Text returns the textual representation of a status code
// (wraps http.StatusText).
func Text(code int) string {
	return http.StatusText(code)
}
