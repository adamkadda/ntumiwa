package logging

import (
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/cookies"
	"github.com/adamkadda/ntumiwa/internal/session"
)

/*
	This type should be the only wrapper around the
	http.ResponseWriter type.

	Only the logging middleware directly accesses its data fields.
*/

type wrappedWriter struct {
	http.ResponseWriter
	request    *http.Request
	manager    *session.SessionManager
	statusCode int
	size       int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.writeCookieOnce()

	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}

func (w *wrappedWriter) Write(b []byte) (int, error) {
	w.writeCookieOnce()

	bytesWritten, err := w.ResponseWriter.Write(b)
	w.size += bytesWritten
	return bytesWritten, err
}

/*
	Call stack when writing response body (or headers):

	wrappedWriter.Write(b)
	 └─> writeCookieOnce()
		  └─> SessionManager.WriteCookie(w.ResponseWriter, r)
			   └─> cookies.WriteSigned(w.ResponseWriter, cookie, secretKey)
					└─> cookies.Write(w.ResponseWriter, cookie)
						 └─> http.SetCookie(w.ResponseWriter, &cookie)
							  └─> w.ResponseWriter.Header().Add("Set-Cookie", ...)

	Always pass the underlying http.ResponseWriter (w.ResponseWriter),
	not the wrappedWriter itself.

	If we passed wrappedWriter, downstream functions might call our
	overridden Write or WriteHeader methods, leading to unintended
	recursion.

	Downstream code is designed to work with the original
	http.ResponseWriter, not a wrapper around it.
*/

func (w *wrappedWriter) writeCookieOnce() {
	ctx := w.request.Context()
	if cookies.CookieWritten(ctx) {
		return
	}

	w.manager.WriteCookie(w.ResponseWriter, w.request)

	w.request = w.request.WithContext(cookies.WithCookieWritten(ctx))
}
