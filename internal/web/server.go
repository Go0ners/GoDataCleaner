// Package web provides the HTTP server and WebUI for GoDataCleaner.
package web

import (
	"fmt"
	"log"
	"net/http"

	"godatacleaner/internal/storage"
)

// Server handles HTTP requests for the WebUI and REST API.
type Server struct {
	storage *storage.Storage
	host    string
	port    int
}

// NewServer creates a new web server.
func NewServer(storage *storage.Storage, host string, port int) *Server {
	return &Server{
		storage: storage,
		host:    host,
		port:    port,
	}
}

// Start starts the HTTP server with configured routes.
// It sets up the HTTP router with routes for the WebUI and REST API.
func (s *Server) Start() error {
	// Create a new ServeMux for routing
	mux := http.NewServeMux()

	// Configure routes for WebUI
	mux.HandleFunc("GET /", s.handleIndex)

	// Configure routes for Torrent API
	mux.HandleFunc("GET /api/torrent/files", s.handleTorrentFiles)
	mux.HandleFunc("GET /api/torrent/stats", s.handleTorrentStats)
	mux.HandleFunc("GET /api/torrent/folders", s.handleTorrentFolders)

	// Configure routes for Local API
	mux.HandleFunc("GET /api/local/files", s.handleLocalFiles)
	mux.HandleFunc("GET /api/local/stats", s.handleLocalStats)
	mux.HandleFunc("GET /api/local/folders", s.handleLocalFolders)

	// Configure routes for Orphans API
	mux.HandleFunc("GET /api/orphans/files", s.handleOrphanFiles)
	mux.HandleFunc("GET /api/orphans/stats", s.handleOrphanStats)
	mux.HandleFunc("GET /api/orphans/export", s.handleOrphanExport)

	// Configure routes for Unknown extensions API
	mux.HandleFunc("GET /api/unknown/extensions", s.handleUnknownExtensions)

	// Build the server address
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Log server startup
	log.Printf("Starting web server on http://%s", addr)

	// Start the HTTP server
	return http.ListenAndServe(addr, mux)
}
