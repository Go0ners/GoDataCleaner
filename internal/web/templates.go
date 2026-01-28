// Package web provides HTML templates for the WebUI.
package web

import "net/http"

// renderTemplate renders the WebUI HTML template.
func renderTemplate(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexTemplate))
}

const indexTemplate = `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoDataCleaner</title>
    <script src="https://unpkg.com/react@18/umd/react.production.min.js" crossorigin></script>
    <script src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js" crossorigin></script>
    <script src="https://unpkg.com/@babel/standalone/babel.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #1a1a2e; color: #eee; min-height: 100vh; }
        .container { width: 100%; padding: 20px; }
        h1 { color: #00d9ff; margin-bottom: 20px; }
        .tabs { display: flex; gap: 10px; margin-bottom: 20px; }
        .tab { padding: 12px 24px; background: #16213e; border: none; color: #888; cursor: pointer; border-radius: 8px; font-size: 14px; transition: all 0.2s; }
        .tab:hover { background: #1f3460; color: #fff; }
        .tab.active { background: #00d9ff; color: #1a1a2e; font-weight: 600; }
        .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .card { background: #16213e; padding: 20px; border-radius: 12px; }
        .card h3 { color: #888; font-size: 12px; text-transform: uppercase; margin-bottom: 8px; }
        .card .value { font-size: 28px; font-weight: 700; color: #00d9ff; }
        .card .sub { font-size: 12px; color: #666; margin-top: 4px; }
        .controls { display: flex; gap: 10px; margin-bottom: 15px; flex-wrap: wrap; }
        .search { flex: 1; min-width: 200px; padding: 10px 15px; background: #16213e; border: 1px solid #333; border-radius: 8px; color: #fff; font-size: 14px; }
        .search:focus { outline: none; border-color: #00d9ff; }
        select { padding: 10px 15px; background: #16213e; border: 1px solid #333; border-radius: 8px; color: #fff; font-size: 14px; cursor: pointer; }
        table { width: 100%; border-collapse: collapse; background: #16213e; border-radius: 12px; overflow: hidden; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid #222; }
        th { background: #0f1729; color: #888; font-size: 12px; text-transform: uppercase; cursor: pointer; user-select: none; }
        th:hover { color: #00d9ff; }
        tr:hover { background: #1f3460; }
        .size { color: #00d9ff; font-weight: 500; }
        .category { padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
        .category.movies { background: #e74c3c33; color: #e74c3c; }
        .category.shows { background: #3498db33; color: #3498db; }
        .category.4k { background: #f39c1233; color: #f39c12; }
        .category.unknown { background: #95a5a633; color: #95a5a6; }
        .pagination { display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 20px; }
        .pagination button { padding: 8px 16px; background: #16213e; border: 1px solid #333; border-radius: 6px; color: #fff; cursor: pointer; }
        .pagination button:hover:not(:disabled) { background: #1f3460; border-color: #00d9ff; }
        .pagination button:disabled { opacity: 0.5; cursor: not-allowed; }
        .pagination span { color: #888; }
        .export-btn { padding: 10px 20px; background: #00d9ff; border: none; border-radius: 8px; color: #1a1a2e; font-weight: 600; cursor: pointer; }
        .export-btn:hover { background: #00b8d9; }
        .chart-container { background: #16213e; padding: 20px; border-radius: 12px; height: 400px; }
        .loading { text-align: center; padding: 40px; color: #888; }
        .path { max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: #aaa; }
    </style>
</head>
<body>
    <div id="root"></div>
    <script type="text/babel">
        const { useState, useEffect, useRef } = React;

        function formatSize(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function Card({ title, value, sub }) {
            return (
                <div className="card">
                    <h3>{title}</h3>
                    <div className="value">{value}</div>
                    {sub && <div className="sub">{sub}</div>}
                </div>
            );
        }

        function DataTable({ data, columns, sort, order, onSort, loading }) {
            if (loading) return <div className="loading">Chargement...</div>;
            return (
                <table>
                    <thead>
                        <tr>
                            {columns.map(col => (
                                <th key={col.key} onClick={() => onSort(col.key)}>
                                    {col.label} {sort === col.key ? (order === 'asc' ? '↑' : '↓') : ''}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {data.map((row, i) => (
                            <tr key={i}>
                                {columns.map(col => (
                                    <td key={col.key} className={col.className}>
                                        {col.render ? col.render(row[col.key], row) : row[col.key]}
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            );
        }

        function Pagination({ page, totalPages, onPageChange }) {
            return (
                <div className="pagination">
                    <button onClick={() => onPageChange(1)} disabled={page <= 1}>««</button>
                    <button onClick={() => onPageChange(page - 1)} disabled={page <= 1}>«</button>
                    <span>Page {page} / {totalPages || 1}</span>
                    <button onClick={() => onPageChange(page + 1)} disabled={page >= totalPages}>»</button>
                    <button onClick={() => onPageChange(totalPages)} disabled={page >= totalPages}>»»</button>
                </div>
            );
        }

        function TorrentsTab() {
            const [data, setData] = useState([]);
            const [stats, setStats] = useState({ total_files: 0, total_torrents: 0, total_size: 0 });
            const [page, setPage] = useState(1);
            const [totalPages, setTotalPages] = useState(1);
            const [search, setSearch] = useState('');
            const [sort, setSort] = useState('size');
            const [order, setOrder] = useState('desc');
            const [loading, setLoading] = useState(true);

            useEffect(() => {
                let ignore = false;
                setLoading(true);
                fetch('/api/torrent/stats').then(r => r.json()).then(d => { if (!ignore) setStats(d); });
                fetch('/api/torrent/files?page=' + page + '&per_page=50&sort=' + sort + '&order=' + order + '&search=' + encodeURIComponent(search))
                    .then(r => r.json())
                    .then(d => {
                        if (!ignore) {
                            setData(d.data || []);
                            setTotalPages(d.total_pages || 1);
                            setLoading(false);
                        }
                    });
                return () => { ignore = true; };
            }, [page, sort, order, search]);

            const handleSort = (col) => {
                if (sort === col) setOrder(order === 'asc' ? 'desc' : 'asc');
                else { setSort(col); setOrder('desc'); }
                setPage(1);
            };

            const columns = [
                { key: 'file_name', label: 'Fichier', className: '', render: (v) => v },
                { key: 'file_path', label: 'Chemin', className: 'path', render: (v) => v },
                { key: 'torrent_name', label: 'Torrent', className: '', render: (v) => v },
                { key: 'size', label: 'Taille', className: 'size', render: (v) => formatSize(v) },
            ];

            return (
                <div>
                    <div className="cards">
                        <Card title="Torrents" value={(stats.total_torrents || 0).toLocaleString()} />
                        <Card title="Fichiers" value={(stats.total_files || 0).toLocaleString()} />
                        <Card title="Poids total" value={formatSize(stats.total_size || 0)} />
                    </div>
                    <div className="controls">
                        <input className="search" placeholder="Rechercher..." value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} />
                    </div>
                    <DataTable data={data} columns={columns} sort={sort} order={order} onSort={handleSort} loading={loading} />
                    <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                </div>
            );
        }

        function LocalTab() {
            const [data, setData] = useState([]);
            const [stats, setStats] = useState([]);
            const [page, setPage] = useState(1);
            const [totalPages, setTotalPages] = useState(1);
            const [search, setSearch] = useState('');
            const [category, setCategory] = useState('');
            const [sort, setSort] = useState('size');
            const [order, setOrder] = useState('desc');
            const [loading, setLoading] = useState(true);

            useEffect(() => {
                let ignore = false;
                setLoading(true);
                fetch('/api/local/stats').then(r => r.json()).then(d => { if (!ignore) setStats(d.categories || []); });
                fetch('/api/local/files?page=' + page + '&per_page=50&sort=' + sort + '&order=' + order + '&search=' + encodeURIComponent(search) + '&category=' + category)
                    .then(r => r.json())
                    .then(d => {
                        if (!ignore) {
                            setData(d.data || []);
                            setTotalPages(d.total_pages || 1);
                            setLoading(false);
                        }
                    });
                return () => { ignore = true; };
            }, [page, sort, order, search, category]);

            const handleSort = (col) => {
                if (sort === col) setOrder(order === 'asc' ? 'desc' : 'asc');
                else { setSort(col); setOrder('desc'); }
                setPage(1);
            };

            const columns = [
                { key: 'file_name', label: 'Fichier', render: (v) => v },
                { key: 'file_path', label: 'Chemin', className: 'path', render: (v) => v },
                { key: 'category', label: 'Catégorie', render: (v) => <span className={'category ' + v}>{v}</span> },
                { key: 'size', label: 'Taille', className: 'size', render: (v) => formatSize(v) },
            ];

            const totalFiles = stats.reduce((a, c) => a + c.file_count, 0);
            const totalSize = stats.reduce((a, c) => a + c.total_size, 0);

            return (
                <div>
                    <div className="cards">
                        <Card title="Fichiers" value={totalFiles.toLocaleString()} />
                        <Card title="Poids total" value={formatSize(totalSize)} />
                    </div>
                    <div className="controls">
                        <input className="search" placeholder="Rechercher..." value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} />
                        <select value={category} onChange={e => { setCategory(e.target.value); setPage(1); }}>
                            <option value="">Toutes catégories</option>
                            <option value="4k">4K</option>
                            <option value="movies">Movies</option>
                            <option value="shows">Shows</option>
                        </select>
                    </div>
                    <DataTable data={data} columns={columns} sort={sort} order={order} onSort={handleSort} loading={loading} />
                    <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                </div>
            );
        }

        function OrphansTab() {
            const [data, setData] = useState([]);
            const [stats, setStats] = useState([]);
            const [page, setPage] = useState(1);
            const [totalPages, setTotalPages] = useState(1);
            const [search, setSearch] = useState('');
            const [category, setCategory] = useState('');
            const [sort, setSort] = useState('size');
            const [order, setOrder] = useState('desc');
            const [loading, setLoading] = useState(true);

            useEffect(() => {
                let ignore = false;
                setLoading(true);
                fetch('/api/orphans/stats').then(r => r.json()).then(d => { if (!ignore) setStats(d.categories || []); });
                fetch('/api/orphans/files?page=' + page + '&per_page=50&sort=' + sort + '&order=' + order + '&search=' + encodeURIComponent(search) + '&category=' + category)
                    .then(r => r.json())
                    .then(d => {
                        if (!ignore) {
                            setData(d.data || []);
                            setTotalPages(d.total_pages || 1);
                            setLoading(false);
                        }
                    });
                return () => { ignore = true; };
            }, [page, sort, order, search, category]);

            const handleSort = (col) => {
                if (sort === col) setOrder(order === 'asc' ? 'desc' : 'asc');
                else { setSort(col); setOrder('desc'); }
                setPage(1);
            };

            const columns = [
                { key: 'file_name', label: 'Fichier', render: (v) => v },
                { key: 'file_path', label: 'Chemin', className: 'path', render: (v) => v },
                { key: 'category', label: 'Catégorie', render: (v) => <span className={'category ' + v}>{v}</span> },
                { key: 'size', label: 'Taille', className: 'size', render: (v) => formatSize(v) },
            ];

            const totalFiles = stats.reduce((a, c) => a + c.file_count, 0);
            const totalSize = stats.reduce((a, c) => a + c.total_size, 0);

            return (
                <div>
                    <div className="cards">
                        <Card title="Fichiers" value={totalFiles.toLocaleString()} />
                        <Card title="Poids total" value={formatSize(totalSize)} />
                    </div>
                    <div className="controls">
                        <input className="search" placeholder="Rechercher..." value={search} onChange={e => { setSearch(e.target.value); setPage(1); }} />
                        <select value={category} onChange={e => { setCategory(e.target.value); setPage(1); }}>
                            <option value="">Toutes catégories</option>
                            <option value="4k">4K</option>
                            <option value="movies">Movies</option>
                            <option value="shows">Shows</option>
                        </select>
                        <a href="/api/orphans/export" className="export-btn">Exporter CSV</a>
                    </div>
                    <DataTable data={data} columns={columns} sort={sort} order={order} onSort={handleSort} loading={loading} />
                    <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                </div>
            );
        }

        function StatsTab() {
            const chartRef = useRef(null);
            const chartInstance = useRef(null);
            const [folders, setFolders] = useState([]);

            useEffect(() => {
                fetch('/api/torrent/folders').then(r => r.json()).then(d => setFolders(d.folders || []));
            }, []);

            useEffect(() => {
                if (!chartRef.current || folders.length === 0) return;
                if (chartInstance.current) chartInstance.current.destroy();

                const ctx = chartRef.current.getContext('2d');
                chartInstance.current = new Chart(ctx, {
                    type: 'bar',
                    data: {
                        labels: folders.map(f => f.folder || 'Racine'),
                        datasets: [{
                            label: 'Taille (GB)',
                            data: folders.map(f => f.total_size / (1024 * 1024 * 1024)),
                            backgroundColor: '#00d9ff',
                            borderRadius: 4,
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } },
                        scales: {
                            x: { ticks: { color: '#888' }, grid: { color: '#333' } },
                            y: { ticks: { color: '#888' }, grid: { color: '#333' } }
                        }
                    }
                });

                return () => { if (chartInstance.current) chartInstance.current.destroy(); };
            }, [folders]);

            return (
                <div>
                    <div className="cards">
                        <Card title="Dossiers" value={folders.length} />
                    </div>
                    <div className="chart-container">
                        <canvas ref={chartRef}></canvas>
                    </div>
                </div>
            );
        }

        function App() {
            const [tab, setTab] = useState('torrents');

            return (
                <div className="container">
                    <h1>🧹 GoDataCleaner</h1>
                    <div className="tabs">
                        <button className={'tab' + (tab === 'torrents' ? ' active' : '')} onClick={() => setTab('torrents')}>Torrents</button>
                        <button className={'tab' + (tab === 'local' ? ' active' : '')} onClick={() => setTab('local')}>Local</button>
                        <button className={'tab' + (tab === 'orphans' ? ' active' : '')} onClick={() => setTab('orphans')}>Orphelins</button>
                        <button className={'tab' + (tab === 'stats' ? ' active' : '')} onClick={() => setTab('stats')}>Stats</button>
                    </div>
                    {tab === 'torrents' && <TorrentsTab />}
                    {tab === 'local' && <LocalTab />}
                    {tab === 'orphans' && <OrphansTab />}
                    {tab === 'stats' && <StatsTab />}
                </div>
            );
        }

        ReactDOM.createRoot(document.getElementById('root')).render(<App />);
    </script>
</body>
</html>`
