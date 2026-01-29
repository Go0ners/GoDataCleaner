// Package web provides HTTP handlers for the REST API.
package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"godatacleaner/internal/models"
)

// parseQueryOptions extracts pagination parameters from the request.
func parseQueryOptions(r *http.Request) models.QueryOptions {
	opts := models.QueryOptions{
		Page:    1,
		PerPage: 100,
		Order:   "asc",
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			opts.Page = v
		}
	}
	if p := r.URL.Query().Get("per_page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 && v <= 1000 {
			opts.PerPage = v
		}
	}
	if s := r.URL.Query().Get("sort"); s != "" {
		opts.Sort = s
	}
	if o := r.URL.Query().Get("order"); o == "asc" || o == "desc" {
		opts.Order = o
	}
	if s := r.URL.Query().Get("search"); s != "" {
		opts.Search = s
	}
	if c := r.URL.Query().Get("category"); c != "" {
		opts.Category = c
	}
	if u := r.URL.Query().Get("unique"); u == "true" {
		opts.Unique = true
	}
	return opts
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func totalPages(total int64, perPage int) int {
	return int((total + int64(perPage) - 1) / int64(perPage))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w)
}

func (s *Server) handleTorrentFiles(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	files, total, err := s.storage.GetTorrentFiles(context.Background(), opts)
	if err != nil {
		writeError(w, 500, "Failed to get torrent files")
		return
	}
	if files == nil {
		files = []models.TorrentFile{}
	}
	writeJSON(w, 200, models.PaginatedResponse{
		Data: files, Total: total, Page: opts.Page, PerPage: opts.PerPage, TotalPages: totalPages(total, opts.PerPage),
	})
}

func (s *Server) handleTorrentStats(w http.ResponseWriter, r *http.Request) {
	unique := r.URL.Query().Get("unique") == "true"
	stats, err := s.storage.GetTorrentStats(context.Background(), unique)
	if err != nil {
		writeError(w, 500, "Failed to get torrent stats")
		return
	}
	writeJSON(w, 200, models.TorrentStatsResponse{
		TotalFiles: stats.TotalFiles, TotalTorrents: stats.TotalTorrents, TotalSize: stats.TotalSize,
	})
}

func (s *Server) handleTorrentFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := s.storage.GetFolderStats(context.Background(), "torrent_files")
	if err != nil {
		writeError(w, 500, "Failed to get folder stats")
		return
	}
	if folders == nil {
		folders = []models.FolderStats{}
	}
	writeJSON(w, 200, models.FolderStatsResponse{Folders: folders})
}

func (s *Server) handleLocalFiles(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	files, total, err := s.storage.GetLocalFiles(context.Background(), opts)
	if err != nil {
		writeError(w, 500, "Failed to get local files")
		return
	}
	if files == nil {
		files = []models.LocalFile{}
	}
	writeJSON(w, 200, models.PaginatedResponse{
		Data: files, Total: total, Page: opts.Page, PerPage: opts.PerPage, TotalPages: totalPages(total, opts.PerPage),
	})
}

func (s *Server) handleLocalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.storage.GetLocalStats(context.Background())
	if err != nil {
		writeError(w, 500, "Failed to get local stats")
		return
	}
	if stats == nil {
		stats = []models.CategoryStats{}
	}
	writeJSON(w, 200, models.CategoryStatsResponse{Categories: stats})
}

func (s *Server) handleLocalFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := s.storage.GetFolderStats(context.Background(), "local_files")
	if err != nil {
		writeError(w, 500, "Failed to get folder stats")
		return
	}
	if folders == nil {
		folders = []models.FolderStats{}
	}
	writeJSON(w, 200, models.FolderStatsResponse{Folders: folders})
}

func (s *Server) handleOrphanFiles(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	files, total, err := s.storage.GetOrphanFiles(context.Background(), opts)
	if err != nil {
		writeError(w, 500, "Failed to get orphan files")
		return
	}
	if files == nil {
		files = []models.OrphanFile{}
	}
	writeJSON(w, 200, models.PaginatedResponse{
		Data: files, Total: total, Page: opts.Page, PerPage: opts.PerPage, TotalPages: totalPages(total, opts.PerPage),
	})
}

func (s *Server) handleOrphanStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.storage.GetOrphanStats(context.Background())
	if err != nil {
		writeError(w, 500, "Failed to get orphan stats")
		return
	}
	if stats == nil {
		stats = []models.CategoryStats{}
	}
	writeJSON(w, 200, models.CategoryStatsResponse{Categories: stats})
}

func (s *Server) handleUnknownExtensions(w http.ResponseWriter, r *http.Request) {
	stats, err := s.storage.GetUnknownExtensionStats(context.Background())
	if err != nil {
		writeError(w, 500, "Failed to get extension stats")
		return
	}
	if stats == nil {
		stats = []models.ExtensionStats{}
	}
	writeJSON(w, 200, models.ExtensionStatsResponse{Extensions: stats})
}

func (s *Server) handleOrphanExport(w http.ResponseWriter, r *http.Request) {
	// Get all orphan files (no pagination for export)
	opts := models.QueryOptions{Page: 1, PerPage: 1000000}
	files, _, err := s.storage.GetOrphanFiles(context.Background(), opts)
	if err != nil {
		writeError(w, 500, "Failed to get orphan files")
		return
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=orphans.csv")
	w.WriteHeader(200)

	// Write CSV content (just file paths)
	for _, f := range files {
		w.Write([]byte(f.FilePath + "\n"))
	}
}
