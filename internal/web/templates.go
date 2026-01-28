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
        table { width: 100%; border-collapse: collapse; background: #16213e; border-radius: 12px; overflow: hidden; table-layout: fixed; }
        th, td { padding: 12px 15px; text-align: left; border-bottom: 1px solid #222; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        th { background: #0f1729; color: #888; font-size: 12px; text-transform: uppercase; cursor: pointer; user-select: none; }
        th:hover { color: #00d9ff; }
        tr:hover { background: #1f3460; }
        .size { color: #00d9ff; font-weight: 500; white-space: nowrap; }
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
            const pieChartRef = useRef(null);
            const orphanChartRef = useRef(null);
            const healthChartRef = useRef(null);
            const pieChartInstance = useRef(null);
            const orphanChartInstance = useRef(null);
            const healthChartInstance = useRef(null);
            
            const [torrentStats, setTorrentStats] = useState({ total_files: 0, total_torrents: 0, total_size: 0 });
            const [localStats, setLocalStats] = useState([]);
            const [orphanStats, setOrphanStats] = useState([]);
            const [extensionStats, setExtensionStats] = useState([]);
            const [loading, setLoading] = useState(true);

            useEffect(() => {
                Promise.all([
                    fetch('/api/torrent/stats').then(r => r.json()),
                    fetch('/api/local/stats').then(r => r.json()),
                    fetch('/api/orphans/stats').then(r => r.json()),
                    fetch('/api/unknown/extensions').then(r => r.json())
                ]).then(([ts, ls, os, es]) => {
                    setTorrentStats(ts);
                    setLocalStats(ls.categories || []);
                    setOrphanStats(os.categories || []);
                    setExtensionStats(es.extensions || []);
                    setLoading(false);
                });
            }, []);

            useEffect(() => {
                if (!healthChartRef.current || localStats.length === 0) return;
                if (healthChartInstance.current) healthChartInstance.current.destroy();
                const totalLocal = localStats.reduce((a, c) => a + c.file_count, 0);
                const totalOrphan = orphanStats.reduce((a, c) => a + c.file_count, 0);
                const healthy = totalLocal - totalOrphan;
                const ctx = healthChartRef.current.getContext('2d');
                healthChartInstance.current = new Chart(ctx, {
                    type: 'doughnut',
                    data: {
                        labels: ['Sains', 'Orphelins'],
                        datasets: [{ data: [healthy, totalOrphan], backgroundColor: ['#2ecc71', '#e74c3c'], borderWidth: 0 }]
                    },
                    options: { responsive: true, maintainAspectRatio: false, cutout: '75%', plugins: { legend: { display: false } } }
                });
                return () => { if (healthChartInstance.current) healthChartInstance.current.destroy(); };
            }, [localStats, orphanStats]);

            useEffect(() => {
                if (!pieChartRef.current || localStats.length === 0) return;
                if (pieChartInstance.current) pieChartInstance.current.destroy();
                const colors = { '4k': '#f39c12', 'movies': '#e74c3c', 'shows': '#3498db', 'unknown': '#95a5a6' };
                const ctx = pieChartRef.current.getContext('2d');
                pieChartInstance.current = new Chart(ctx, {
                    type: 'doughnut',
                    data: {
                        labels: localStats.map(s => s.category.toUpperCase()),
                        datasets: [{ data: localStats.map(s => s.total_size), backgroundColor: localStats.map(s => colors[s.category] || '#666'), borderWidth: 0 }]
                    },
                    options: {
                        responsive: true, maintainAspectRatio: false,
                        plugins: { legend: { position: 'right', labels: { color: '#ccc', padding: 15 } }, tooltip: { callbacks: { label: (ctx) => ctx.label + ': ' + formatSize(ctx.raw) } } }
                    }
                });
                return () => { if (pieChartInstance.current) pieChartInstance.current.destroy(); };
            }, [localStats]);

            useEffect(() => {
                if (!orphanChartRef.current || localStats.length === 0) return;
                if (orphanChartInstance.current) orphanChartInstance.current.destroy();
                const categories = ['4k', 'movies', 'shows', 'unknown'];
                const localData = categories.map(c => { const s = localStats.find(x => x.category === c); return s ? s.total_size / (1024*1024*1024) : 0; });
                const orphanData = categories.map(c => { const s = orphanStats.find(x => x.category === c); return s ? s.total_size / (1024*1024*1024) : 0; });
                const ctx = orphanChartRef.current.getContext('2d');
                orphanChartInstance.current = new Chart(ctx, {
                    type: 'bar',
                    data: {
                        labels: categories.map(c => c.toUpperCase()),
                        datasets: [
                            { label: 'Local (GB)', data: localData, backgroundColor: '#3498db', borderRadius: 4 },
                            { label: 'Orphelins (GB)', data: orphanData, backgroundColor: '#e74c3c', borderRadius: 4 }
                        ]
                    },
                    options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { labels: { color: '#888' } } }, scales: { x: { ticks: { color: '#888' }, grid: { color: '#222' } }, y: { ticks: { color: '#888' }, grid: { color: '#222' } } } }
                });
                return () => { if (orphanChartInstance.current) orphanChartInstance.current.destroy(); };
            }, [localStats, orphanStats]);

            if (loading) return <div className="loading">Chargement...</div>;

            const totalLocalFiles = localStats.reduce((a, c) => a + c.file_count, 0);
            const totalLocalSize = localStats.reduce((a, c) => a + c.total_size, 0);
            const totalOrphanFiles = orphanStats.reduce((a, c) => a + c.file_count, 0);
            const totalOrphanSize = orphanStats.reduce((a, c) => a + c.total_size, 0);
            const orphanPercent = totalLocalFiles > 0 ? ((totalOrphanFiles / totalLocalFiles) * 100).toFixed(1) : 0;
            const orphanSizePercent = totalLocalSize > 0 ? ((totalOrphanSize / totalLocalSize) * 100).toFixed(1) : 0;
            const healthyFiles = totalLocalFiles - totalOrphanFiles;
            const healthPercent = totalLocalFiles > 0 ? ((healthyFiles / totalLocalFiles) * 100).toFixed(0) : 100;

            const ProgressBar = ({ percent, color }) => (
                <div style={{background: '#0f1729', borderRadius: '4px', height: '8px', width: '100%', marginTop: '8px'}}>
                    <div style={{background: color, borderRadius: '4px', height: '100%', width: percent + '%'}}></div>
                </div>
            );
            return (
                <div>
                    <h2 style={{color: '#00d9ff', marginBottom: '20px', fontSize: '18px'}}>📊 Vue d'ensemble</h2>
                    <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px', marginBottom: '30px'}}>
                        <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px'}}>
                            <Card title="Torrents" value={(torrentStats.total_torrents || 0).toLocaleString()} sub={torrentStats.total_files?.toLocaleString() + ' fichiers'} />
                            <Card title="Espace Torrents" value={formatSize(torrentStats.total_size || 0)} />
                            <Card title="Fichiers Locaux" value={totalLocalFiles.toLocaleString()} />
                            <Card title="Espace Local" value={formatSize(totalLocalSize)} />
                        </div>
                        <div className="card">
                            <h3>💚 Santé du stockage</h3>
                            <div style={{display: 'flex', alignItems: 'center', gap: '20px', marginTop: '15px', height: 'calc(100% - 40px)'}}>
                                <div style={{width: '120px', height: '120px', position: 'relative', flexShrink: 0}}>
                                    <canvas ref={healthChartRef}></canvas>
                                    <div style={{position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', textAlign: 'center'}}>
                                        <div style={{fontSize: '22px', fontWeight: 'bold', color: healthPercent > 80 ? '#2ecc71' : healthPercent > 50 ? '#f39c12' : '#e74c3c'}}>{healthPercent}%</div>
                                        <div style={{fontSize: '9px', color: '#888'}}>SAIN</div>
                                    </div>
                                </div>
                                <div style={{flex: 1}}>
                                    <div style={{marginBottom: '15px'}}>
                                        <div style={{display: 'flex', justifyContent: 'space-between', fontSize: '13px', marginBottom: '6px'}}><span style={{color: '#2ecc71'}}>● Fichiers sains</span><span>{healthyFiles.toLocaleString()}</span></div>
                                        <ProgressBar percent={100 - orphanPercent} color="#2ecc71" />
                                    </div>
                                    <div>
                                        <div style={{display: 'flex', justifyContent: 'space-between', fontSize: '13px', marginBottom: '6px'}}><span style={{color: '#e74c3c'}}>● Fichiers orphelins</span><span>{totalOrphanFiles.toLocaleString()}</span></div>
                                        <ProgressBar percent={orphanPercent} color="#e74c3c" />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <h2 style={{color: '#00d9ff', margin: '30px 0 20px', fontSize: '18px'}}>🗑️ Orphelins</h2>
                    <div className="cards">
                        <div className="card"><h3>Fichiers orphelins</h3><div className="value" style={{color: '#e74c3c'}}>{totalOrphanFiles.toLocaleString()}</div><div className="sub">{orphanPercent}% du total</div><ProgressBar percent={orphanPercent} color="#e74c3c" /></div>
                        <div className="card"><h3>Espace orphelin</h3><div className="value" style={{color: '#e74c3c'}}>{formatSize(totalOrphanSize)}</div><div className="sub">{orphanSizePercent}% du stockage</div><ProgressBar percent={orphanSizePercent} color="#e74c3c" /></div>
                        <div className="card"><h3>Espace récupérable</h3><div className="value" style={{color: '#f39c12'}}>{formatSize(totalOrphanSize)}</div><div className="sub">Si nettoyage complet</div></div>
                    </div>

                    <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '20px', margin: '30px 0'}}>
                        <div className="chart-container" style={{height: '280px', padding: '15px'}}>
                            <h3 style={{color: '#888', marginBottom: '15px', fontSize: '14px'}}>📁 Répartition par catégorie</h3>
                            <div style={{height: 'calc(100% - 30px)'}}><canvas ref={pieChartRef}></canvas></div>
                        </div>
                        <div className="chart-container" style={{height: '280px', padding: '15px'}}>
                            <h3 style={{color: '#888', marginBottom: '15px', fontSize: '14px'}}>📊 Local vs Orphelins (GB)</h3>
                            <div style={{height: 'calc(100% - 30px)'}}><canvas ref={orphanChartRef}></canvas></div>
                        </div>
                    </div>

                    <h2 style={{color: '#00d9ff', marginBottom: '20px', fontSize: '18px'}}>📋 Détail par catégorie</h2>
                    <table>
                        <thead><tr><th>Catégorie</th><th>Fichiers</th><th>Taille</th><th>Orphelins</th><th>Taille orph.</th><th>% Orph.</th><th>Santé</th></tr></thead>
                        <tbody>
                            {['4k', 'movies', 'shows', 'unknown'].map(cat => {
                                const local = localStats.find(s => s.category === cat) || { file_count: 0, total_size: 0 };
                                const orphan = orphanStats.find(s => s.category === cat) || { file_count: 0, total_size: 0 };
                                const pct = local.file_count > 0 ? ((orphan.file_count / local.file_count) * 100).toFixed(1) : 0;
                                const health = 100 - pct;
                                return (
                                    <tr key={cat}>
                                        <td><span className={'category ' + cat}>{cat.toUpperCase()}</span></td>
                                        <td>{local.file_count.toLocaleString()}</td>
                                        <td className="size">{formatSize(local.total_size)}</td>
                                        <td style={{color: '#e74c3c'}}>{orphan.file_count.toLocaleString()}</td>
                                        <td style={{color: '#e74c3c'}}>{formatSize(orphan.total_size)}</td>
                                        <td style={{color: pct > 50 ? '#e74c3c' : pct > 20 ? '#f39c12' : '#2ecc71', fontWeight: 'bold'}}>{pct}%</td>
                                        <td><div style={{display: 'flex', alignItems: 'center', gap: '8px'}}><div style={{flex: 1, background: '#0f1729', borderRadius: '4px', height: '6px'}}><div style={{background: health > 80 ? '#2ecc71' : health > 50 ? '#f39c12' : '#e74c3c', borderRadius: '4px', height: '100%', width: health + '%'}}></div></div><span style={{fontSize: '11px', color: '#888'}}>{health.toFixed(0)}%</span></div></td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
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
