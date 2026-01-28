# 🧹 GoDataCleaner

Application CLI en Go pour indexer et gérer les fichiers de torrents qBittorrent avec une WebUI de visualisation.

## Fonctionnalités

- **Synchronisation qBittorrent** : Récupère tous les fichiers de tous les torrents via l'API qBittorrent v2
- **Scan local** : Parcourt récursivement un répertoire pour indexer les fichiers locaux
- **Détection des orphelins** : Identifie les fichiers présents localement mais absents de qBittorrent
- **WebUI React** : Interface web pour explorer, rechercher et comparer les données
- **Export CSV** : Exporte la liste des fichiers orphelins

## Installation

### Prérequis

- Go 1.21+
- CGO activé (pour SQLite)
- qBittorrent avec l'API Web activée

### Build

```bash
make build
```

Le binaire sera créé dans `./build/godatacleaner`

### Cross-compilation

Builds disponibles pour plusieurs plateformes :

```bash
make build-linux-amd64    # Linux 64-bit
make build-linux-arm64    # Linux ARM64
make build-darwin-amd64   # macOS Intel
make build-darwin-arm64   # macOS Apple Silicon
make build-windows-amd64  # Windows 64-bit
make build-all            # Toutes les plateformes
```

**Note** : La cross-compilation nécessite des cross-compilers C (CGO requis pour SQLite) :

```bash
# macOS : installer les cross-compilers via Homebrew
brew tap messense/macos-cross-toolchains
brew install x86_64-unknown-linux-gnu    # pour Linux AMD64
brew install aarch64-unknown-linux-gnu   # pour Linux ARM64
brew install mingw-w64                   # pour Windows
```

## Utilisation

### Commandes

```bash
# Synchroniser les données qBittorrent et les fichiers locaux vers SQLite
./build/godatacleaner sync

# Démarrer le serveur WebUI
./build/godatacleaner web

# Afficher les statistiques
./build/godatacleaner stats

# Afficher l'aide
./build/godatacleaner help
```

### Configuration

L'application supporte deux méthodes de configuration :
1. **Fichier JSON** (`config.json`)
2. **Variables d'environnement**

**Priorité** : Variables d'environnement > Fichier config > Valeurs par défaut

#### Fichier de configuration (config.json)

Créez un fichier `config.json` à la racine du projet (ou spécifiez le chemin via `CONFIG_PATH`) :

```json
{
  "local_host": "localhost",
  "local_port": 61913,
  "qbittorrent_host": "192.168.1.100",
  "qbittorrent_port": 8080,
  "qbittorrent_username": "admin",
  "qbittorrent_password": "monmotdepasse",
  "qbittorrent_max_workers": 10,
  "sqlite_path": "./data/torrents.db",
  "sqlite_batch_size": 1000,
  "local_path": "/mnt/media/torrents"
}
```

#### Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `CONFIG_PATH` | ./config.json | Chemin du fichier de configuration |
| `LOCAL_HOST` | localhost | Hôte du serveur HTTP |
| `LOCAL_PORT` | 61913 | Port du serveur HTTP |
| `QBITTORRENT_HOST` | qbt.home | Hôte qBittorrent |
| `QBITTORRENT_PORT` | 80 | Port qBittorrent |
| `QBITTORRENT_USERNAME` | admin | Utilisateur qBittorrent |
| `QBITTORRENT_PASSWORD` | adminadmin | Mot de passe qBittorrent |
| `QBITTORRENT_MAX_WORKERS` | 10 | Workers parallèles pour la sync |
| `SQLITE_PATH` | ./data/torrents.db | Chemin de la base SQLite |
| `SQLITE_BATCH_SIZE` | 1000 | Taille des lots d'insertion |
| `LOCAL_PATH` | ./data/torrents | Répertoire à scanner |

### Exemple

```bash
# Option 1 : Utiliser un fichier config.json
cp config.example.json config.json
# Éditer config.json avec vos paramètres
./build/godatacleaner sync

# Option 2 : Utiliser les variables d'environnement
export QBITTORRENT_HOST=192.168.1.100
export QBITTORRENT_PORT=8080
export LOCAL_PATH=/mnt/media
./build/godatacleaner sync

# Option 3 : Mixer les deux (env vars ont la priorité)
# config.json contient la config de base
# Les env vars permettent de surcharger ponctuellement
LOCAL_PORT=8080 ./build/godatacleaner web
# Ouvrir http://localhost:8080
```

## WebUI

Interface React avec 4 onglets :

- **Torrents** : Liste des fichiers indexés depuis qBittorrent avec recherche et tri
- **Local** : Liste des fichiers scannés localement avec filtrage par catégorie
- **Orphelins** : Fichiers présents localement mais absents de qBittorrent (à nettoyer)
- **Stats** : Graphique de distribution par dossier

### Catégories

Les fichiers locaux sont automatiquement catégorisés selon leur chemin :
- `4k` : Fichiers dans un dossier contenant `/4k/`
- `movies` : Fichiers dans un dossier contenant `/movies/`
- `shows` : Fichiers dans un dossier contenant `/shows/`
- `unknown` : Autres fichiers

## Architecture

```
cmd/godatacleaner/main.go     # Point d'entrée CLI
internal/
├── config/config.go          # Configuration via env vars
├── models/data.go            # Structures de données
├── qbittorrent/client.go     # Client API qBittorrent v2
├── scanner/scanner.go        # Scanner de fichiers locaux
├── storage/sqlite.go         # Storage SQLite optimisé
└── web/
    ├── server.go             # Serveur HTTP
    ├── handlers.go           # Handlers API REST
    └── templates.go          # Template WebUI React
```

## API REST

| Endpoint | Description |
|----------|-------------|
| `GET /` | WebUI HTML |
| `GET /api/torrent/files` | Fichiers torrents paginés |
| `GET /api/torrent/stats` | Stats globales torrents |
| `GET /api/torrent/folders` | Stats par dossier |
| `GET /api/local/files` | Fichiers locaux paginés |
| `GET /api/local/stats` | Stats par catégorie |
| `GET /api/orphans/files` | Fichiers orphelins paginés |
| `GET /api/orphans/stats` | Stats orphelins par catégorie |
| `GET /api/orphans/export` | Export CSV des orphelins |

### Paramètres de pagination

- `page` : Numéro de page (défaut: 1)
- `per_page` : Éléments par page (défaut: 100, max: 1000)
- `sort` : Colonne de tri (file_name, file_path, size, category)
- `order` : Ordre de tri (asc, desc)
- `search` : Recherche dans le nom/chemin
- `category` : Filtrer par catégorie (4k, movies, shows)

## Optimisations

- **SQLite** : Mode WAL, cache 10000 pages, busy_timeout 5000ms
- **HTTP** : Pool de connexions (max 100), compression
- **Sync** : Workers parallèles avec errgroup
- **Scan** : Streaming via channels (pas de chargement complet en mémoire)

## Dépendances

- `github.com/mattn/go-sqlite3` - Driver SQLite
- `github.com/autobrr/go-qbittorrent` - Client API qBittorrent
- `golang.org/x/sync` - errgroup pour workers parallèles

## Licence

MIT
