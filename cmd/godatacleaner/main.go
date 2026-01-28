// Package main provides the CLI entry point for GoDataCleaner.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"godatacleaner/internal/config"
	"godatacleaner/internal/models"
	"godatacleaner/internal/qbittorrent"
	"godatacleaner/internal/scanner"
	"godatacleaner/internal/storage"
	"godatacleaner/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]
	switch command {
	case "sync":
		runSync()
	case "web":
		runWeb()
	case "stats":
		runStats()
	case "help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Commande inconnue: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func runSync() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur de configuration: %v", err)
	}

	// Créer le répertoire pour la DB si nécessaire
	if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0755); err != nil {
		log.Fatalf("Erreur création répertoire DB: %v", err)
	}

	// Initialiser le storage
	store, err := storage.NewStorage(cfg.SQLitePath, cfg.SQLiteBatchSize)
	if err != nil {
		log.Fatalf("Erreur connexion SQLite: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		log.Fatalf("Erreur initialisation DB: %v", err)
	}

	// Sync qBittorrent
	log.Println("🔄 Synchronisation qBittorrent...")
	qbtClient, err := qbittorrent.NewClient(cfg.QBittorrentURL(), cfg.QBittorrentUsername, cfg.QBittorrentPassword, cfg.QBittorrentMaxWorkers)
	if err != nil {
		log.Fatalf("Erreur création client qBittorrent: %v", err)
	}

	if err := qbtClient.Login(ctx); err != nil {
		log.Printf("⚠️  Impossible de se connecter à qBittorrent: %v", err)
	} else {
		// Clear et sync torrents
		if err := store.ClearTorrentFiles(ctx); err != nil {
			log.Fatalf("Erreur clear torrent_files: %v", err)
		}

		torrents, err := qbtClient.GetTorrents(ctx)
		if err != nil {
			log.Printf("⚠️  Erreur récupération torrents: %v", err)
		} else {
			total := len(torrents)
			fmt.Printf("📦 %d torrents trouvés\n", total)
			var allFiles []models.TorrentFile
			for i, t := range torrents {
				files, err := qbtClient.GetTorrentFiles(ctx, t.Hash)
				if err != nil {
					continue
				}
				allFiles = append(allFiles, files...)
				// Progress on single line
				percent := float64(i+1) / float64(total) * 100
				fmt.Printf("\r⏳ Progression: %d/%d (%.1f%%) - %d fichiers", i+1, total, percent, len(allFiles))
			}
			fmt.Println() // New line after progress
			if err := store.InsertTorrentFiles(ctx, allFiles); err != nil {
				log.Fatalf("Erreur insertion fichiers torrents: %v", err)
			}
			fmt.Printf("✅ %d fichiers torrents synchronisés\n", len(allFiles))
		}
	}

	// Sync local
	fmt.Println("🔄 Scan des fichiers locaux...")
	if err := store.ClearLocalFiles(ctx); err != nil {
		log.Fatalf("Erreur clear local_files: %v", err)
	}

	scan := scanner.NewScanner(cfg.LocalPath)
	filesChan, errsChan := scan.Scan(ctx)

	var localFiles []models.LocalFile
	count := 0
	for f := range filesChan {
		localFiles = append(localFiles, f)
		count++
		if count%100 == 0 {
			fmt.Printf("\r⏳ Scan: %d fichiers trouvés", count)
		}
	}
	fmt.Println() // New line after progress
	if err := <-errsChan; err != nil {
		log.Printf("⚠️  Erreur scan: %v", err)
	}

	fmt.Printf("💾 Insertion de %d fichiers en base...\n", len(localFiles))
	if err := store.InsertLocalFiles(ctx, localFiles); err != nil {
		log.Fatalf("Erreur insertion fichiers locaux: %v", err)
	}
	fmt.Printf("✅ %d fichiers locaux synchronisés\n", len(localFiles))

	fmt.Println("🎉 Synchronisation terminée!")
}

func runWeb() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur de configuration: %v", err)
	}

	store, err := storage.NewStorage(cfg.SQLitePath, cfg.SQLiteBatchSize)
	if err != nil {
		log.Fatalf("Erreur connexion SQLite: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		log.Fatalf("Erreur initialisation DB: %v", err)
	}

	server := web.NewServer(store, cfg.LocalHost, cfg.LocalPort)
	log.Printf("🌐 Démarrage du serveur sur http://%s:%d", cfg.LocalHost, cfg.LocalPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}

func runStats() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur de configuration: %v", err)
	}

	store, err := storage.NewStorage(cfg.SQLitePath, cfg.SQLiteBatchSize)
	if err != nil {
		log.Fatalf("Erreur connexion SQLite: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Stats torrents
	torrentStats, err := store.GetTorrentStats(ctx)
	if err != nil {
		log.Fatalf("Erreur stats torrents: %v", err)
	}

	// Stats locaux
	localStats, err := store.GetLocalStats(ctx)
	if err != nil {
		log.Fatalf("Erreur stats locaux: %v", err)
	}

	// Stats orphelins
	orphanStats, err := store.GetOrphanStats(ctx)
	if err != nil {
		log.Fatalf("Erreur stats orphelins: %v", err)
	}

	fmt.Println("📊 Statistiques GoDataCleaner")
	fmt.Println("═══════════════════════════════")
	fmt.Println()
	fmt.Println("🌐 Torrents:")
	fmt.Printf("   Fichiers: %d\n", torrentStats.TotalFiles)
	fmt.Printf("   Torrents: %d\n", torrentStats.TotalTorrents)
	fmt.Printf("   Taille:   %s\n", formatSize(torrentStats.TotalSize))
	fmt.Println()
	fmt.Println("💾 Fichiers locaux:")
	for _, s := range localStats {
		fmt.Printf("   %s: %d fichiers (%s)\n", s.Category, s.FileCount, formatSize(s.TotalSize))
	}
	fmt.Println()
	fmt.Println("🗑️  Orphelins:")
	var totalOrphans int64
	var totalOrphanSize int64
	for _, s := range orphanStats {
		fmt.Printf("   %s: %d fichiers (%s)\n", s.Category, s.FileCount, formatSize(s.TotalSize))
		totalOrphans += s.FileCount
		totalOrphanSize += s.TotalSize
	}
	fmt.Printf("   Total: %d fichiers (%s)\n", totalOrphans, formatSize(totalOrphanSize))
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func printHelp() {
	fmt.Println("GoDataCleaner - Gestionnaire de fichiers torrents")
	fmt.Println()
	fmt.Println("Usage: godatacleaner <commande>")
	fmt.Println()
	fmt.Println("Commandes:")
	fmt.Println("  sync   Synchroniser qBittorrent et fichiers locaux vers SQLite")
	fmt.Println("  web    Démarrer le serveur WebUI")
	fmt.Println("  stats  Afficher les statistiques de la base")
	fmt.Println("  help   Afficher cette aide")
	fmt.Println()
	fmt.Println("Variables d'environnement:")
	fmt.Println("  LOCAL_HOST              Hôte du serveur (défaut: localhost)")
	fmt.Println("  LOCAL_PORT              Port du serveur (défaut: 61913)")
	fmt.Println("  QBITTORRENT_HOST        Hôte qBittorrent (défaut: qbt.home)")
	fmt.Println("  QBITTORRENT_PORT        Port qBittorrent (défaut: 80)")
	fmt.Println("  QBITTORRENT_USERNAME    Utilisateur (défaut: admin)")
	fmt.Println("  QBITTORRENT_PASSWORD    Mot de passe (défaut: adminadmin)")
	fmt.Println("  SQLITE_PATH             Chemin de la DB (défaut: ./data/torrents.db)")
	fmt.Println("  LOCAL_PATH              Chemin à scanner (défaut: ./data/torrents)")
}
