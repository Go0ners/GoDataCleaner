// Package qbittorrent provides a client for the qBittorrent Web API v2.
package qbittorrent

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	qbt "github.com/autobrr/go-qbittorrent"
	"golang.org/x/sync/errgroup"

	"godatacleaner/internal/models"
)

// Client wraps the qBittorrent API client with additional functionality.
type Client struct {
	client     *qbt.Client
	maxWorkers int
}

// NewClient creates a new qBittorrent client with connection pooling.
// The HTTP transport is configured with:
// - MaxIdleConns: 100 (maximum idle connections across all hosts)
// - MaxIdleConnsPerHost: 100 (maximum idle connections per host)
// - IdleConnTimeout: 90 seconds
// - DisableCompression: false (compression enabled)
func NewClient(host, username, password string, maxWorkers int) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("qbittorrent: host cannot be empty")
	}
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default to 10 workers
	}

	// Configure HTTP transport with connection pooling (max 100 connections)
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	// Create HTTP client with custom transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Create qBittorrent client with configuration
	qbtClient := qbt.NewClient(qbt.Config{
		Host:     host,
		Username: username,
		Password: password,
		Timeout:  30, // 30 seconds timeout
	})

	// Apply custom HTTP client with connection pooling
	qbtClient = qbtClient.WithHTTPClient(httpClient)

	return &Client{
		client:     qbtClient,
		maxWorkers: maxWorkers,
	}, nil
}

// Login authenticates the client with the qBittorrent API.
// Returns an error if authentication fails with the HTTP status code.
func (c *Client) Login(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("qbittorrent: client not initialized")
	}

	err := c.client.LoginCtx(ctx)
	if err != nil {
		return fmt.Errorf("qbittorrent: authentication failed: %w", err)
	}

	return nil
}

// GetTorrents retrieves the list of all torrents from qBittorrent.
// Returns a slice of Torrent models with hash, name, size, and save path.
func (c *Client) GetTorrents(ctx context.Context) ([]models.Torrent, error) {
	if c.client == nil {
		return nil, fmt.Errorf("qbittorrent: client not initialized")
	}

	// Get all torrents without any filter
	qbtTorrents, err := c.client.GetTorrentsCtx(ctx, qbt.TorrentFilterOptions{})
	if err != nil {
		return nil, fmt.Errorf("qbittorrent: failed to get torrents: %w", err)
	}

	// Convert qBittorrent torrents to our model
	torrents := make([]models.Torrent, 0, len(qbtTorrents))
	for _, t := range qbtTorrents {
		torrents = append(torrents, models.Torrent{
			Hash:     t.Hash,
			Name:     t.Name,
			Size:     t.Size,
			SavePath: t.SavePath,
		})
	}

	return torrents, nil
}

// GetTorrentFiles retrieves the files of a specific torrent by its hash.
// Returns a slice of TorrentFile models with file details.
func (c *Client) GetTorrentFiles(ctx context.Context, hash string) ([]models.TorrentFile, error) {
	if c.client == nil {
		return nil, fmt.Errorf("qbittorrent: client not initialized")
	}

	if hash == "" {
		return nil, fmt.Errorf("qbittorrent: torrent hash cannot be empty")
	}

	// Get files for the specified torrent using GetFilesInformationCtx
	qbtFiles, err := c.client.GetFilesInformationCtx(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("qbittorrent: failed to get files for torrent %s: %w", hash, err)
	}

	// We need to get the torrent info to get the name and save path
	torrents, err := c.client.GetTorrentsCtx(ctx, qbt.TorrentFilterOptions{
		Hashes: []string{hash},
	})
	if err != nil {
		return nil, fmt.Errorf("qbittorrent: failed to get torrent info for %s: %w", hash, err)
	}

	var torrentName, savePath string
	if len(torrents) > 0 {
		torrentName = torrents[0].Name
		savePath = torrents[0].SavePath
	}

	// Handle nil response
	if qbtFiles == nil {
		return []models.TorrentFile{}, nil
	}

	// Convert qBittorrent files to our model
	files := make([]models.TorrentFile, 0, len(*qbtFiles))
	for _, f := range *qbtFiles {
		// Build the full file path: savePath + torrentName + file.Name
		// The file.Name from qBittorrent is relative to the torrent root
		fullPath := filepath.Join(savePath, torrentName, f.Name)

		files = append(files, models.TorrentFile{
			TorrentHash: hash,
			TorrentName: torrentName,
			FileName:    filepath.Base(f.Name),
			FilePath:    fullPath,
			Size:        f.Size,
		})
	}

	return files, nil
}

// GetMaxWorkers returns the configured maximum number of workers.
func (c *Client) GetMaxWorkers() int {
	return c.maxWorkers
}

// SyncAll synchronizes all torrents and their files in parallel.
// Uses errgroup with worker limit for parallel processing.
// Returns two channels:
// - files: streams TorrentFile as they are retrieved
// - errs: streams errors encountered during synchronization
// Both channels are closed when synchronization is complete.
func (c *Client) SyncAll(ctx context.Context) (<-chan models.TorrentFile, <-chan error) {
	files := make(chan models.TorrentFile)
	errs := make(chan error, 1) // Buffered to avoid blocking on error send

	go func() {
		defer close(files)
		defer close(errs)

		// Get all torrents first
		torrents, err := c.GetTorrents(ctx)
		if err != nil {
			select {
			case errs <- fmt.Errorf("failed to get torrents: %w", err):
			case <-ctx.Done():
			}
			return
		}

		// Create errgroup with context for parallel processing
		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(c.maxWorkers)

		// Mutex to protect channel writes
		var mu sync.Mutex

		// Process each torrent in parallel with worker limit
		for _, torrent := range torrents {
			t := torrent // Capture loop variable

			g.Go(func() error {
				// Check if context is cancelled
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				// Get files for this torrent
				torrentFiles, err := c.GetTorrentFiles(gCtx, t.Hash)
				if err != nil {
					// Send error to error channel (non-blocking)
					select {
					case errs <- fmt.Errorf("failed to get files for torrent %s: %w", t.Hash, err):
					default:
						// Error channel full, skip this error
					}
					// Continue processing other torrents, don't fail the whole sync
					return nil
				}

				// Stream files through the channel
				mu.Lock()
				defer mu.Unlock()

				for _, file := range torrentFiles {
					select {
					case files <- file:
					case <-gCtx.Done():
						return gCtx.Err()
					}
				}

				return nil
			})
		}

		// Wait for all goroutines to complete
		if err := g.Wait(); err != nil {
			select {
			case errs <- fmt.Errorf("sync failed: %w", err):
			case <-ctx.Done():
			default:
				// Error channel full
			}
		}
	}()

	return files, errs
}
