package api

import (
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicMiddleware creates a New Relic middleware for HTTP request tracking
func NewRelicMiddleware(app *newrelic.Application) Middleware {
	if app == nil {
		// Return no-op middleware if New Relic is not configured
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start New Relic transaction
			txn := app.StartTransaction(r.Method + " " + r.URL.Path)
			defer txn.End()

			// Add transaction to request context
			r = newrelic.RequestWithTransactionContext(r, txn)

			// Set transaction name based on route
			txn.SetName(r.Method + " " + r.URL.Path)

			// Wrap response writer to capture status code
			wrappedWriter := &newRelicResponseWriter{
				ResponseWriter: w,
				statusCode:    http.StatusOK,
				txn:           txn,
			}

			// Process request
			next.ServeHTTP(wrappedWriter, r)

			// Set status code for transaction
			txn.SetWebResponse(wrappedWriter).WriteHeader(wrappedWriter.statusCode)
		})
	}
}

// newRelicResponseWriter wraps http.ResponseWriter to capture status code
type newRelicResponseWriter struct {
	http.ResponseWriter
	statusCode int
	txn         *newrelic.Transaction
}

func (rw *newRelicResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StartTransaction starts a New Relic transaction for a given request
func StartTransaction(app *newrelic.Application, name string) *newrelic.Transaction {
	if app == nil {
		return nil
	}
	return app.StartTransaction(name)
}

// StartSegment starts a New Relic segment within a transaction
func StartSegment(txn *newrelic.Transaction, name string) *newrelic.Segment {
	if txn == nil {
		return nil
	}
	return newrelic.StartSegment(txn, name)
}

// EndSegment ends a New Relic segment
func EndSegment(seg *newrelic.Segment) {
	if seg != nil {
		seg.End()
	}
}

