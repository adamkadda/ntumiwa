package middleware

import (
	"net/http"

	"github.com/adamkadda/ntumiwa-site/internal/cookies"
	"github.com/adamkadda/ntumiwa-site/internal/session"
)

/*
	This type should be the only wrapper around the
	http.ResponseWriter type.

	Only the logging middleware directly accesses its data fields.
*/

type WrappedWriter struct {
	http.ResponseWriter
	request    *http.Request
	manager    *session.SessionManager
	statusCode int
	size       int
}

func NewWrappedWriter(
	w http.ResponseWriter,
	r *http.Request,
	m *session.SessionManager,
) *WrappedWriter {

	return &WrappedWriter{
		ResponseWriter: w,
		request:        r,
		manager:        m,
		statusCode:     http.StatusOK,
	}
}

func (w *WrappedWriter) WriteHeader(code int) {
	w.writeCookieOnce()

	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}

func (w *WrappedWriter) Write(b []byte) (int, error) {
	w.writeCookieOnce()

	bytesWritten, err := w.ResponseWriter.Write(b)
	w.size += bytesWritten
	return bytesWritten, err
}

/*
	Call stack when writing response body (or headers):

	WrappedWriter.Write(b)
	 └─> writeCookieOnce()
		  └─> SessionManager.WriteCookie(w.ResponseWriter, r)
			   └─> cookies.WriteSigned(w.ResponseWriter, cookie, secretKey)
					└─> cookies.Write(w.ResponseWriter, cookie)
						 └─> http.SetCookie(w.ResponseWriter, &cookie)
							  └─> w.ResponseWriter.Header().Add("Set-Cookie", ...)

	Always pass the underlying http.ResponseWriter (w.ResponseWriter),
	not the WrappedWriter itself.

	If we passed WrappedWriter, downstream functions might call our
	overridden Write or WriteHeader methods, leading to unintended
	recursion.

	Downstream code is designed to work with the original
	http.ResponseWriter, not a wrapper around it.
*/

func (w *WrappedWriter) writeCookieOnce() {
	ctx := w.request.Context()
	if cookies.CookieWritten(ctx) {
		return
	}

	w.manager.WriteCookie(w.ResponseWriter, w.request)

	w.request = w.request.WithContext(cookies.WithCookieWritten(ctx))
}
