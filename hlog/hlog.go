// Package hlog provides a set of http.Handler helpers for zerolog.
package hlog

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FromRequest gets the logger in the request's context.
// This is a shortcut for log.Ctx(r.Context())
func FromRequest(r *http.Request) zerolog.Logger {
	return log.Ctx(r.Context())
}

// NewHandler injects log into requests context.
func NewHandler(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(log.WithContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// URLHandler adds the requested URL as a field to the context's logger
// using fieldKey as field key.
func URLHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			log = log.With().Str(fieldKey, r.URL.String()).Logger()
			r = r.WithContext(log.WithContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// MethodHandler adds the request method as a field to the context's logger
// using fieldKey as field key.
func MethodHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			log = log.With().Str(fieldKey, r.Method).Logger()
			r = r.WithContext(log.WithContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// RequestHandler adds the request method and URL as a field to the context's logger
// using fieldKey as field key.
func RequestHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())
			log = log.With().Str(fieldKey, r.Method+" "+r.URL.String()).Logger()
			r = r.WithContext(log.WithContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// RemoteAddrHandler adds the request's remote address as a field to the context's logger
// using fieldKey as field key.
func RemoteAddrHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				log := zerolog.Ctx(r.Context())
				log = log.With().Str(fieldKey, host).Logger()
				r = r.WithContext(log.WithContext(r.Context()))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserAgentHandler adds the request's user-agent as a field to the context's logger
// using fieldKey as field key.
func UserAgentHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ua := r.Header.Get("User-Agent"); ua != "" {
				log := zerolog.Ctx(r.Context())
				log = log.With().Str(fieldKey, ua).Logger()
				r = r.WithContext(log.WithContext(r.Context()))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RefererHandler adds the request's referer as a field to the context's logger
// using fieldKey as field key.
func RefererHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ref := r.Header.Get("Referer"); ref != "" {
				log := zerolog.Ctx(r.Context())
				log = log.With().Str(fieldKey, ref).Logger()
				r = r.WithContext(log.WithContext(r.Context()))
			}
			next.ServeHTTP(w, r)
		})
	}
}

type idKey struct{}

// IDFromRequest returns the unique id accociated to the request if any.
func IDFromRequest(r *http.Request) (id xid.ID, ok bool) {
	if r == nil {
		return
	}
	id, ok = r.Context().Value(idKey{}).(xid.ID)
	return
}

// RequestIDHandler returns a handler setting a unique id to the request which can
// be gathered using IDFromRequest(req). This generated id is added as a field to the
// logger using the passed fieldKey as field name. The id is also added as a response
// header if the headerName is not empty.
//
// The generated id is a URL safe base64 encoded mongo object-id-like unique id.
// Mongo unique id generation algorithm has been selected as a trade-off between
// size and ease of use: UUID is less space efficient and snowflake requires machine
// configuration.
func RequestIDHandler(fieldKey, headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := IDFromRequest(r)
			if !ok {
				id = xid.New()
				ctx := context.WithValue(r.Context(), idKey{}, id)
				r = r.WithContext(ctx)
			}
			if fieldKey != "" {
				log := zerolog.Ctx(r.Context())
				log = log.With().Str(fieldKey, id.String()).Logger()
				r = r.WithContext(log.WithContext(r.Context()))
			}
			if headerName != "" {
				w.Header().Set(headerName, id.String())
			}
			next.ServeHTTP(w, r)
		})
	}
}
