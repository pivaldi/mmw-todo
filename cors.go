package todo

import (
	"net/http"

	"github.com/pivaldi/mmw/todo/internal/infra/config"

	// Import the Connect CORS helpers and the rs/cors package
	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

const maxAge = 7200 // Unit is second

// withCORS adds CORS support for Connect, gRPC, and gRPC-Web
func withCORS(conf *config.Config, h http.Handler) http.Handler {
	allowedOrigins := "*"
	if !conf.Environment.IsDev() {
		allowedOrigins = conf.Server.Host
	}

	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigins},
		// The official Connect helpers to inject all required gRPC/Connect headers
		// See the documentation: https://pkg.go.dev/connectrpc.com/cors#section-readme
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         maxAge, // Optional cache preflight requests
	})

	return middleware.Handler(h)
}
