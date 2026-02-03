package httptransport

import "github.com/gorilla/mux"

// NewRouter creates a new Gorilla Mux router with the allocation and order endpoints.
// Registers POST /allocations and POST /orders routes with their respective handlers.
// Returns the configured router.
func NewRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/allocations", handler.CreateAllocation).Methods("POST")
	router.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
	return router
}
