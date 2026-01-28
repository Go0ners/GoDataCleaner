// Package models defines the data structures used throughout GoDataCleaner.
package models

// Torrent represents a torrent from qBittorrent.
type Torrent struct {
	Hash     string
	Name     string
	Size     int64
	SavePath string
}

// TorrentFile represents a file within a torrent.
type TorrentFile struct {
	TorrentHash string
	TorrentName string
	FileName    string
	FilePath    string
	Size        int64
}

// LocalFile represents a file found on the local filesystem.
type LocalFile struct {
	FilePath string
	FileName string
	Size     int64
	Category string // "4k", "movies", "shows"
}

// OrphanFile represents a local file that is not present in the torrent database.
type OrphanFile struct {
	FilePath string
	FileName string
	Size     int64
	Category string
}

// Stats represents global statistics for torrents.
type Stats struct {
	TotalFiles    int64
	TotalTorrents int64
	TotalSize     int64
}

// FolderStats represents statistics for a specific folder.
type FolderStats struct {
	Folder    string `json:"folder"`
	FileCount int64  `json:"file_count"`
	TotalSize int64  `json:"total_size"`
}

// CategoryStats represents statistics for a specific category.
type CategoryStats struct {
	Category  string `json:"category"`
	FileCount int64  `json:"file_count"`
	TotalSize int64  `json:"total_size"`
}

// QueryOptions defines parameters for paginated queries.
type QueryOptions struct {
	Page     int
	PerPage  int
	Sort     string
	Order    string // "asc" ou "desc"
	Search   string
	Category string
}

// PaginatedResponse represents a paginated API response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// TorrentStatsResponse represents the API response for torrent statistics.
type TorrentStatsResponse struct {
	TotalFiles    int64 `json:"total_files"`
	TotalTorrents int64 `json:"total_torrents"`
	TotalSize     int64 `json:"total_size"`
}

// FolderStatsResponse represents the API response for folder statistics.
type FolderStatsResponse struct {
	Folders []FolderStats `json:"folders"`
}

// CategoryStatsResponse represents the API response for category statistics.
type CategoryStatsResponse struct {
	Categories []CategoryStats `json:"categories"`
}
