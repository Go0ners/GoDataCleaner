// Package storage provides SQLite storage functionality for GoDataCleaner.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"godatacleaner/internal/models"
)

// Storage manages SQLite database operations.
type Storage struct {
	db        *sql.DB
	batchSize int
}

// NewStorage creates a new SQLite storage with WAL mode optimizations.
// DSN includes: WAL journal mode, 10000 page cache, 5000ms busy timeout, shared cache.
func NewStorage(path string, batchSize int) (*Storage, error) {
	// Build DSN with optimizations as per requirements 3.1, 3.6
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_cache_size=10000&_busy_timeout=5000&cache=shared", path)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set max open connections to 1 to avoid "database is locked" errors
	db.SetMaxOpenConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Storage{
		db:        db,
		batchSize: batchSize,
	}, nil
}

// Initialize creates the database tables and indexes.
// Creates torrent_files and local_files tables with appropriate indexes.
func (s *Storage) Initialize(ctx context.Context) error {
	// SQL statements for table and index creation as per requirements 3.2, 3.3
	statements := []string{
		// Table des fichiers torrents
		`CREATE TABLE IF NOT EXISTS torrent_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			torrent_hash TEXT NOT NULL,
			torrent_name TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			relative_path TEXT NOT NULL,
			size INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Index sur torrent_hash
		`CREATE INDEX IF NOT EXISTS idx_torrent_hash ON torrent_files(torrent_hash)`,
		// Index sur file_path
		`CREATE INDEX IF NOT EXISTS idx_torrent_file_path ON torrent_files(file_path)`,
		// Index sur file_name
		`CREATE INDEX IF NOT EXISTS idx_torrent_file_name ON torrent_files(file_name)`,
		// Index sur relative_path pour les JOINs orphelins
		`CREATE INDEX IF NOT EXISTS idx_torrent_relative_path ON torrent_files(relative_path)`,

		// Table des fichiers locaux
		`CREATE TABLE IF NOT EXISTS local_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL,
			relative_path TEXT NOT NULL,
			size INTEGER NOT NULL,
			category TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Index sur file_path
		`CREATE INDEX IF NOT EXISTS idx_local_file_path ON local_files(file_path)`,
		// Index sur category
		`CREATE INDEX IF NOT EXISTS idx_local_category ON local_files(category)`,
		// Index sur file_name
		`CREATE INDEX IF NOT EXISTS idx_local_file_name ON local_files(file_name)`,
		// Index sur relative_path pour les JOINs orphelins
		`CREATE INDEX IF NOT EXISTS idx_local_relative_path ON local_files(relative_path)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return nil
}

// extractRelativePath extracts the relative path from a full path.
// It looks for /movies/, /shows/, or /4k/ and returns the path from that point.
// If none found, returns the original path.
func extractRelativePath(fullPath string) string {
	markers := []string{"/movies/", "/shows/", "/4k/"}
	for _, marker := range markers {
		if idx := strings.Index(fullPath, marker); idx != -1 {
			return fullPath[idx:]
		}
	}
	return fullPath
}

// normalizeLocalPath removes the /mnt prefix from local paths to match torrent paths.
func normalizeLocalPath(path string) string {
	if strings.HasPrefix(path, "/mnt") {
		return path[4:] // Remove "/mnt"
	}
	return path
}

// InsertTorrentFiles inserts torrent files in batches using prepared statements.
func (s *Storage) InsertTorrentFiles(ctx context.Context, files []models.TorrentFile) error {
	// Handle empty slice gracefully
	if len(files) == 0 {
		return nil
	}

	// Start a transaction for atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO torrent_files (torrent_hash, torrent_name, file_name, file_path, relative_path, size)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert files in batches
	for i := 0; i < len(files); i += s.batchSize {
		end := i + s.batchSize
		if end > len(files) {
			end = len(files)
		}

		// Insert each file in the current batch
		for _, file := range files[i:end] {
			relativePath := extractRelativePath(file.FilePath)
			_, err := stmt.ExecContext(ctx, file.TorrentHash, file.TorrentName, file.FileName, file.FilePath, relativePath, file.Size)
			if err != nil {
				return fmt.Errorf("failed to insert torrent file: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertLocalFiles inserts local files in batches using prepared statements.
func (s *Storage) InsertLocalFiles(ctx context.Context, files []models.LocalFile) error {
	// Handle empty slice gracefully
	if len(files) == 0 {
		return nil
	}

	// Start a transaction for atomicity
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement with INSERT OR REPLACE for UNIQUE constraint on file_path
	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO local_files (file_path, file_name, relative_path, size, category)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert files in batches
	for i := 0; i < len(files); i += s.batchSize {
		end := i + s.batchSize
		if end > len(files) {
			end = len(files)
		}

		// Insert each file in the current batch
		for _, file := range files[i:end] {
			// Normalize path by removing /mnt prefix
			normalizedPath := normalizeLocalPath(file.FilePath)
			relativePath := extractRelativePath(normalizedPath)
			_, err := stmt.ExecContext(ctx, normalizedPath, file.FileName, relativePath, file.Size, file.Category)
			if err != nil {
				return fmt.Errorf("failed to insert local file: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ClearTorrentFiles removes all torrent files from the database.
func (s *Storage) ClearTorrentFiles(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM torrent_files")
	if err != nil {
		return fmt.Errorf("failed to clear torrent_files: %w", err)
	}
	return nil
}

// ClearLocalFiles removes all local files from the database.
func (s *Storage) ClearLocalFiles(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM local_files")
	if err != nil {
		return fmt.Errorf("failed to clear local_files: %w", err)
	}
	return nil
}

// allowedTorrentColumns defines the whitelist of columns allowed for sorting in torrent_files queries.
// This prevents SQL injection via the Sort field.
var allowedTorrentColumns = map[string]string{
	"torrent_hash": "torrent_hash",
	"torrent_name": "torrent_name",
	"file_name":    "file_name",
	"file_path":    "file_path",
	"size":         "size",
}

// allowedLocalColumns defines the whitelist of columns allowed for sorting in local_files queries.
var allowedLocalColumns = map[string]string{
	"file_path": "file_path",
	"file_name": "file_name",
	"size":      "size",
	"category":  "category",
}

// allowedOrphanColumns defines the whitelist of columns allowed for sorting in orphan queries.
var allowedOrphanColumns = map[string]string{
	"file_path": "l.file_path",
	"file_name": "l.file_name",
	"size":      "l.size",
	"category":  "l.category",
}

// normalizeQueryOptions sets default values for pagination options.
// Default Page to 1 if not set, default PerPage to 100 if not set.
func normalizeQueryOptions(opts models.QueryOptions) models.QueryOptions {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PerPage < 1 {
		opts.PerPage = 100
	}
	if opts.PerPage > 1000 {
		opts.PerPage = 1000
	}
	// Normalize order to lowercase
	if opts.Order != "asc" && opts.Order != "desc" {
		opts.Order = "asc"
	}
	return opts
}

// GetTorrentFiles retrieves torrent files with pagination, sorting, and search.
func (s *Storage) GetTorrentFiles(ctx context.Context, opts models.QueryOptions) ([]models.TorrentFile, int64, error) {
	opts = normalizeQueryOptions(opts)

	// Build WHERE clause for search
	var whereClause string
	var args []interface{}
	if opts.Search != "" {
		whereClause = "WHERE file_name LIKE ? OR file_path LIKE ?"
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM torrent_files " + whereClause
	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count torrent files: %w", err)
	}

	// Build ORDER BY clause with whitelist validation
	orderClause := "ORDER BY id ASC"
	if opts.Sort != "" {
		if col, ok := allowedTorrentColumns[opts.Sort]; ok {
			orderClause = fmt.Sprintf("ORDER BY %s %s", col, opts.Order)
		}
	}

	// Calculate offset for pagination
	offset := (opts.Page - 1) * opts.PerPage

	// Build and execute the main query
	query := fmt.Sprintf(
		"SELECT torrent_hash, torrent_name, file_name, file_path, size FROM torrent_files %s %s LIMIT ? OFFSET ?",
		whereClause, orderClause,
	)
	args = append(args, opts.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query torrent files: %w", err)
	}
	defer rows.Close()

	var files []models.TorrentFile
	for rows.Next() {
		var f models.TorrentFile
		if err := rows.Scan(&f.TorrentHash, &f.TorrentName, &f.FileName, &f.FilePath, &f.Size); err != nil {
			return nil, 0, fmt.Errorf("failed to scan torrent file: %w", err)
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating torrent files: %w", err)
	}

	return files, total, nil
}

// GetLocalFiles retrieves local files with pagination, sorting, search, and category filtering.
func (s *Storage) GetLocalFiles(ctx context.Context, opts models.QueryOptions) ([]models.LocalFile, int64, error) {
	opts = normalizeQueryOptions(opts)

	// Build WHERE clause for search and category filtering
	var conditions []string
	var args []interface{}

	if opts.Search != "" {
		conditions = append(conditions, "(file_name LIKE ? OR file_path LIKE ?)")
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if opts.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, opts.Category)
	}

	var whereClause string
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM local_files " + whereClause
	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count local files: %w", err)
	}

	// Build ORDER BY clause with whitelist validation
	orderClause := "ORDER BY id ASC"
	if opts.Sort != "" {
		if col, ok := allowedLocalColumns[opts.Sort]; ok {
			orderClause = fmt.Sprintf("ORDER BY %s %s", col, opts.Order)
		}
	}

	// Calculate offset for pagination
	offset := (opts.Page - 1) * opts.PerPage

	// Build and execute the main query
	query := fmt.Sprintf(
		"SELECT file_path, file_name, size, category FROM local_files %s %s LIMIT ? OFFSET ?",
		whereClause, orderClause,
	)
	args = append(args, opts.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query local files: %w", err)
	}
	defer rows.Close()

	var files []models.LocalFile
	for rows.Next() {
		var f models.LocalFile
		if err := rows.Scan(&f.FilePath, &f.FileName, &f.Size, &f.Category); err != nil {
			return nil, 0, fmt.Errorf("failed to scan local file: %w", err)
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating local files: %w", err)
	}

	return files, total, nil
}

// GetOrphanFiles retrieves orphan files (local files not present in torrent_files) with pagination.
// Comparison is done on relative_path column which is pre-computed and indexed.
func (s *Storage) GetOrphanFiles(ctx context.Context, opts models.QueryOptions) ([]models.OrphanFile, int64, error) {
	opts = normalizeQueryOptions(opts)

	// Build WHERE clause for search and category filtering
	// Base condition: no matching torrent file (orphan detection via LEFT JOIN on relative_path)
	conditions := []string{"t.relative_path IS NULL"}
	var args []interface{}

	if opts.Search != "" {
		conditions = append(conditions, "(l.file_name LIKE ? OR l.file_path LIKE ?)")
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if opts.Category != "" {
		conditions = append(conditions, "l.category = ?")
		args = append(args, opts.Category)
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// Count total matching orphan records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM local_files l
		LEFT JOIN torrent_files t ON l.relative_path = t.relative_path
		%s`, whereClause)

	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orphan files: %w", err)
	}

	// Build ORDER BY clause with whitelist validation
	// Default to size DESC as per design.md orphan query
	orderClause := "ORDER BY l.size DESC"
	if opts.Sort != "" {
		if col, ok := allowedOrphanColumns[opts.Sort]; ok {
			orderClause = fmt.Sprintf("ORDER BY %s %s", col, opts.Order)
		}
	}

	// Calculate offset for pagination
	offset := (opts.Page - 1) * opts.PerPage

	// Build and execute the main query using LEFT JOIN on relative_path
	query := fmt.Sprintf(`
		SELECT l.file_path, l.file_name, l.size, l.category
		FROM local_files l
		LEFT JOIN torrent_files t ON l.relative_path = t.relative_path
		%s
		%s
		LIMIT ? OFFSET ?`, whereClause, orderClause)

	args = append(args, opts.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orphan files: %w", err)
	}
	defer rows.Close()

	var files []models.OrphanFile
	for rows.Next() {
		var f models.OrphanFile
		if err := rows.Scan(&f.FilePath, &f.FileName, &f.Size, &f.Category); err != nil {
			return nil, 0, fmt.Errorf("failed to scan orphan file: %w", err)
		}
		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating orphan files: %w", err)
	}

	return files, total, nil
}

// GetTorrentStats returns global torrent statistics.
// Returns COUNT files, COUNT DISTINCT torrent_hash, SUM size.
func (s *Storage) GetTorrentStats(ctx context.Context) (*models.Stats, error) {
	query := `
		SELECT 
			COUNT(*) as total_files,
			COUNT(DISTINCT torrent_hash) as total_torrents,
			COALESCE(SUM(size), 0) as total_size
		FROM torrent_files
	`

	var stats models.Stats
	err := s.db.QueryRowContext(ctx, query).Scan(&stats.TotalFiles, &stats.TotalTorrents, &stats.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get torrent stats: %w", err)
	}

	return &stats, nil
}

// GetLocalStats returns local file statistics by category.
// Groups by category and returns COUNT files, SUM size per category.
func (s *Storage) GetLocalStats(ctx context.Context) ([]models.CategoryStats, error) {
	query := `
		SELECT 
			category,
			COUNT(*) as file_count,
			COALESCE(SUM(size), 0) as total_size
		FROM local_files
		GROUP BY category
		ORDER BY category ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query local stats: %w", err)
	}
	defer rows.Close()

	var stats []models.CategoryStats
	for rows.Next() {
		var cs models.CategoryStats
		if err := rows.Scan(&cs.Category, &cs.FileCount, &cs.TotalSize); err != nil {
			return nil, fmt.Errorf("failed to scan local stats: %w", err)
		}
		stats = append(stats, cs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating local stats: %w", err)
	}

	return stats, nil
}

// GetOrphanStats returns orphan file statistics by category.
// Uses LEFT JOIN on relative_path column which is pre-computed and indexed.
func (s *Storage) GetOrphanStats(ctx context.Context) ([]models.CategoryStats, error) {
	query := `
		SELECT 
			l.category,
			COUNT(*) as file_count,
			COALESCE(SUM(l.size), 0) as total_size
		FROM local_files l
		LEFT JOIN torrent_files t ON l.relative_path = t.relative_path
		WHERE t.relative_path IS NULL
		GROUP BY l.category
		ORDER BY l.category ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphan stats: %w", err)
	}
	defer rows.Close()

	var stats []models.CategoryStats
	for rows.Next() {
		var cs models.CategoryStats
		if err := rows.Scan(&cs.Category, &cs.FileCount, &cs.TotalSize); err != nil {
			return nil, fmt.Errorf("failed to scan orphan stats: %w", err)
		}
		stats = append(stats, cs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orphan stats: %w", err)
	}

	return stats, nil
}

// allowedTables defines the whitelist of tables allowed for folder stats queries.
var allowedTables = map[string]bool{
	"torrent_files": true,
	"local_files":   true,
}

// GetFolderStats returns statistics by folder.
// Extracts the folder from file_path and groups by folder.
func (s *Storage) GetFolderStats(ctx context.Context, table string) ([]models.FolderStats, error) {
	// Validate table name to prevent SQL injection
	if !allowedTables[table] {
		return nil, fmt.Errorf("invalid table name: %s", table)
	}

	// Extract folder from file_path using SQLite's path manipulation
	// We use substr and instr to extract the first directory component from the path
	// For paths like "movies/action/file.mkv", this extracts "movies"
	// For paths like "file.mkv" (no folder), this returns the filename itself
	query := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN instr(file_path, '/') > 0 THEN substr(file_path, 1, instr(file_path, '/') - 1)
				ELSE file_path
			END as folder,
			COUNT(*) as file_count,
			COALESCE(SUM(size), 0) as total_size
		FROM %s
		GROUP BY folder
		ORDER BY total_size DESC
	`, table)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query folder stats: %w", err)
	}
	defer rows.Close()

	var stats []models.FolderStats
	for rows.Next() {
		var fs models.FolderStats
		if err := rows.Scan(&fs.Folder, &fs.FileCount, &fs.TotalSize); err != nil {
			return nil, fmt.Errorf("failed to scan folder stats: %w", err)
		}
		stats = append(stats, fs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating folder stats: %w", err)
	}

	return stats, nil
}

// Close closes the database connection.
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
