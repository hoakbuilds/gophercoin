package gcd

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

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
func NewRouter(routes Routes) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

// Logger is a wrapper for the http handler.
// It gets passed the handler and returns the same handler
// with added logging and timing functionalities.
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"[GCDAPI] %s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// BuildAndServeAPI is the function used to serve the API endpoints
func (s *Server) BuildAndServeAPI() {

	log.Println("[GCDAPI] Building API endpoints.")

	var routes = Routes{
		Route{
			"Index",
			"GET",
			"/",
			s.Index,
		},
		Route{
			"NewAddress",
			"GET",
			"/new_address",
			s.NewAddress,
		},
		Route{
			"CreateWallet",
			"POST",
			"/create_wallet",
			s.CreateWallet,
		},
		Route{
			"CreateBlockchain",
			"POST",
			"/create_blockchain",
			s.CreateBlockchain,
		},
		Route{
			"GenerateBlocks",
			"POST",
			"/generate_blocks/{Amount}",
			s.GenerateBlocks,
		},
		Route{
			"GetBalance",
			"GET",
			"/get_balance/{Address}",
			s.GetBalance,
		},
		Route{
			"ListAddresses",
			"GET",
			"/list_addresses",
			s.ListAddresses,
		},
		Route{
			"ListMempool",
			"GET",
			"/list_mempool",
			s.ListMempool,
		},
		Route{
			"ListBlocks",
			"GET",
			"/list_blocks",
			s.ListBlocks,
		},
		Route{
			"NodeInfo",
			"GET",
			"/node_info",
			s.NodeInfo,
		},
		Route{
			"SubmitTx",
			"POST",
			"/submit_tx/{From}/{To}/{Amount}",
			s.SubmitTx,
		},
		Route{
			"AddNode",
			"POST",
			"/add_node/{Address}",
			s.AddNode,
		},
	}
	s.Router = NewRouter(routes)

	log.Printf("[GCDAPI] Listening and Serving API. Port: %s", s.cfg.restPort)

	log.Fatal(http.ListenAndServe(":"+s.cfg.restPort,
		handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedOrigins([]string{"*"}),
		)(s.Router),
	),
	)
}
