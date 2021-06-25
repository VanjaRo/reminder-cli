package middleware

import "net/http"

type Middleware struct {
	functions []func(h http.Handler) http.Handler
}

func NewMiddleware(ms ...func(h http.Handler) http.Handler) *Middleware {
	return &Middleware{
		functions: ms,
	}
}

func (m *Middleware) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}
	for i := range m.functions {
		// eache middleware returns http.Handler
		h = m.functions[len(m.functions)-1-i](h)
	}
	return h
}
