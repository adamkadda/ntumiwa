package session

import "net/http"

type sessionWriter struct {
	http.ResponseWriter
	request   *http.Request
	manager   *SessionManager
	cookieSet bool
}

func (w *sessionWriter) WriteHeader(statusCode int) {
	w.writeCookieOnce()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *sessionWriter) Write(b []byte) (int, error) {
	w.writeCookieOnce()
	return w.ResponseWriter.Write(b)
}

func (w *sessionWriter) writeCookieOnce() {
	if w.cookieSet {
		return
	}

	w.manager.WriteCookie(w.ResponseWriter, w.request)

	w.cookieSet = true
}
