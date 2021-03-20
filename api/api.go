package api

import (
	log "github.com/murlokito/gophercoin/log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// APIServer is a structure that defines the API Server.
// This is defined in order for it to be easily created with a structure
// that defines the API Routes
type APIServer struct {
	config Config
	router *mux.Router
	logger log.Logger
	wg     *sync.WaitGroup
}

// Route is a structure that defines the endpoints of the API.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a type that encapsulates a list of Route elements
type Routes []Route

// NewRouter creates a new router based on Gorilla's mux router
// It also wraps the handlers with logging functionality.
func NewRouter(logger log.Logger, routes Routes) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = LogMiddleware(logger, handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

// LogMiddleware is a wrapper for the http handler.
// It gets passed the handler and returns the same handler
// with added logging and timing functionalities.
func LogMiddleware(logger log.Logger, inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)
		details := []log.Detail{
			log.NewDetail("method", r.Method),
			log.NewDetail("route", r.RequestURI),
			log.NewDetail("route_name", name),
			log.NewDetail("duration", time.Since(start)),
		}
		logger.WithDetails(details...).Info("Request handled")
	})
}

// BuildAndServeAPI is the function used to serve the API endpoints
func (s APIServer) BuildAndServeAPI() {
	s.logger.Info("Building API endpoints.")

	s.router = NewRouter(s.logger, s.config.Routes)

	s.logger.Info("Listening and Serving API. Port: %s", s.config.Port)

	err := http.ListenAndServe(":"+s.config.Port,
		handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedOrigins([]string{"*"}),
		)(s.router),
	)

	s.logger.WithError(err)
}

// NewAPIServer creates and runs a new API Server
func NewAPIServer(wg *sync.WaitGroup, config Config) *APIServer {
	server := &APIServer{
		config: config,
		router: nil,
		logger: log.NewLogger(config.LogLevel),
		wg:     wg,
	}

	go server.BuildAndServeAPI()
	wg.Add(1)

	return server
}
