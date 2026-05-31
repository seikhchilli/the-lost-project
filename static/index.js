document.addEventListener('DOMContentLoaded', () => {
    const grid = document.getElementById('titlesGrid');
    const searchInput = document.getElementById('searchInput');
    
    // Navigation
    const navWatched = document.getElementById('nav-watched');
    const navWished = document.getElementById('nav-wished');
    
    // Modals
    const addModal = document.getElementById('addModal');
    const tmdbModal = document.getElementById('tmdbModal');
    
    let currentFilter = 'watched'; // watched, wished
    let currentPage = 1;
    const pageSize = 15;

    const btnPrev = document.getElementById('btnPrev');
    const btnNext = document.getElementById('btnNext');
    const pageInfo = document.getElementById('pageInfo');
    const pagination = document.getElementById('pagination');

    // Initial Load
    fetchTitles(true, false);

    // Event Listeners for Navigation
    navWatched.addEventListener('click', () => { setActiveNav(navWatched); currentFilter = 'watched'; currentPage = 1; fetchTitles(true, false); });
    navWished.addEventListener('click', () => { setActiveNav(navWished); currentFilter = 'wished'; currentPage = 1; fetchTitles(false, true); });

    // Pagination Listeners
    btnPrev.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            executeFetch();
        }
    });

    btnNext.addEventListener('click', () => {
        currentPage++;
        executeFetch();
    });

    function executeFetch() {
        const val = searchInput.value.trim();
        if (val) {
            searchTitles(val);
        } else {
            if(currentFilter === 'watched') fetchTitles(true, false);
            else if(currentFilter === 'wished') fetchTitles(false, true);
        }
    }

    // Live Search
    let searchTimeout;
    searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            const val = e.target.value.trim();
            currentPage = 1;
            executeFetch();
        }, 300);
    });

    // Add Manually Modal
    document.getElementById('btnAddManual').addEventListener('click', () => addModal.classList.add('active'));
    document.getElementById('btnCancelAdd').addEventListener('click', () => addModal.classList.remove('active'));
    document.getElementById('btnSubmitAdd').addEventListener('click', async () => {
        const name = document.getElementById('titleName').value.trim();
        const year = parseInt(document.getElementById('titleYear').value);
        if(!name) return alert('Name is required');
        
        try {
            const res = await fetch('/api/titles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ titles: [{ name, release_year: year }] })
            });
            if(res.ok) {
                addModal.classList.remove('active');
                document.getElementById('titleName').value = '';
                document.getElementById('titleYear').value = '';
                fetchTitles();
            } else {
                const data = await res.json();
                alert('Error: ' + data.message);
            }
        } catch(e) { console.error(e); }
    });

    // TMDB Search Modal
    document.getElementById('btnSearchTMDB').addEventListener('click', () => tmdbModal.classList.add('active'));
    document.getElementById('btnCancelTMDB').addEventListener('click', () => {
        tmdbModal.classList.remove('active');
        document.getElementById('tmdbStatus').innerText = '';
    });
    document.getElementById('btnSubmitTMDB').addEventListener('click', async () => {
        const query = document.getElementById('tmdbQuery').value.trim();
        const statusEl = document.getElementById('tmdbStatus');
        if(!query) return;
        
        statusEl.innerText = 'Fetching metadata from TMDB...';
        try {
            const res = await fetch(`/api/titles/details?movie_name=${encodeURIComponent(query)}`);
            const data = await res.json();
            
            if(res.ok && data.status === 'success' && data.metadata) {
                statusEl.innerText = 'Adding to database...';
                
                // Now POST to add
                const addRes = await fetch('/api/titles', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        titles: [{
                            name: data.metadata.title || data.metadata.name,
                            release_year: data.metadata.release_date ? parseInt(data.metadata.release_date.split('-')[0]) : (data.metadata.first_air_date ? parseInt(data.metadata.first_air_date.split('-')[0]) : 0),
                            overview: data.metadata.overview,
                            imdb_id: data.metadata.imdb_id,
                            imdb_rating: data.metadata.vote_average,
                            poster_path: data.metadata.poster_path,
                            backdrop_path: data.metadata.backdrop_path,
                            genres: data.metadata.genres ? data.metadata.genres.map(g => g.name) : []
                        }]
                    })
                });
                
                if(addRes.ok) {
                    tmdbModal.classList.remove('active');
                    document.getElementById('tmdbQuery').value = '';
                    statusEl.innerText = '';
                    fetchTitles();
                } else {
                    statusEl.innerText = 'Error adding to DB.';
                }
            } else {
                statusEl.innerText = data.message || 'Not found.';
            }
        } catch(e) {
            statusEl.innerText = 'Network error.';
        }
    });

    // --- Core Functions ---
    function setActiveNav(el) {
        [navWatched, navWished].forEach(n => n.classList.remove('active'));
        el.classList.add('active');
        searchInput.value = '';
    }

    async function fetchTitles(watched = null, wished = null) {
        let url = `/api/titles?page=${currentPage}&page_size=${pageSize}`;
        if (watched) url = `/api/titles/watched?page=${currentPage}&page_size=${pageSize}`;
        else if (wished) url = `/api/titles/search?wished=true&page=${currentPage}&page_size=${pageSize}`;

        try {
            const res = await fetch(url);
            const data = await res.json();
            const titles = data.titles || data.watched_titles || [];
            renderGrid(titles);
            updatePaginationUI(data.total);
        } catch (e) {
            console.error('Failed to fetch titles', e);
            renderError();
        }
    }

    async function searchTitles(query) {
        try {
            const res = await fetch(`/api/titles/search?title_names=${encodeURIComponent(query)}&page=${currentPage}&page_size=${pageSize}`);
            const data = await res.json();
            renderGrid(data.titles || []);
            updatePaginationUI(data.total);
        } catch(e) {
            console.error(e);
        }
    }

    function updatePaginationUI(total) {
        if (!total || total <= pageSize) {
            pagination.style.display = 'none';
        } else {
            pagination.style.display = 'flex';
            const totalPages = Math.ceil(total / pageSize);
            pageInfo.innerText = `Page ${currentPage} of ${totalPages}`;
            btnPrev.disabled = currentPage === 1;
            btnNext.disabled = currentPage === totalPages;
        }
    }

    function renderGrid(titles) {
        grid.innerHTML = '';
        if (!titles || titles.length === 0) {
            grid.innerHTML = `
                <div class="empty-state">
                    <h3>No titles found</h3>
                    <p>Try adding some new movies or TV shows!</p>
                </div>
            `;
            return;
        }

        titles.forEach(t => {
            const card = document.createElement('div');
            card.className = 'card';
            
            // Build badges HTML
            let badgesHtml = '';
            if (t.watched) badgesHtml += `<span class="badge watched">Watched</span>`;
            if (t.wished) badgesHtml += `<span class="badge wished">Wishlist</span>`;

            let imgHtml = '';
            if (t.poster_path) {
                imgHtml = `<img src="https://image.tmdb.org/t/p/w500${t.poster_path}" alt="${escapeHTML(t.name)} poster" class="card-image" loading="lazy">`;
            } else {
                imgHtml = `<div class="card-image" style="display:flex; align-items:center; justify-content:center; color:var(--text-muted);">
                    <ion-icon name="film-outline" style="font-size: 4rem; opacity: 0.5;"></ion-icon>
                </div>`;
            }

            card.innerHTML = `
                ${imgHtml}
                <div class="card-content">
                    <div class="card-title">${escapeHTML(t.name)}</div>
                    <div class="card-year">${t.release_year ? t.release_year : 'Unknown Year'}</div>
                    <div class="badges">${badgesHtml}</div>
                    <div class="card-actions">
                        <button class="icon-btn watched" onclick="markAction(${t.id}, 'watch')" title="Mark Watched">
                            <ion-icon name="eye-outline"></ion-icon>
                        </button>
                        <button class="icon-btn wished" onclick="markAction(${t.id}, 'wish')" title="Add to Wishlist">
                            <ion-icon name="bookmark-outline"></ion-icon>
                        </button>
                        <div style="flex:1"></div>
                        <button class="icon-btn delete" onclick="deleteAction(${t.id}, ${t.watched})" title="Remove">
                            <ion-icon name="trash-outline"></ion-icon>
                        </button>
                    </div>
                </div>
            `;
            grid.appendChild(card);
        });
    }

    function renderError() {
        grid.innerHTML = `
            <div class="empty-state">
                <h3>Oops! Something went wrong.</h3>
                <p>Could not connect to the server.</p>
            </div>
        `;
    }

    // Global action handlers
    window.markAction = async (id, type) => {
        try {
            await fetch(`/api/titles/${id}/${type}`, { method: 'PUT' });
            // refresh current view
            if(currentFilter === 'watched') fetchTitles(true, false);
            else if(currentFilter === 'wished') fetchTitles(false, true);
        } catch(e) { console.error(e); }
    };
    
    window.deleteAction = async (id, isWatched) => {
        try {
            // Depending on status, we might remove from watched or wished.
            // For simplicity in this UI, if we hit delete, let's just remove watched state if it's watched, else wished.
            const type = isWatched ? 'watch' : 'wish';
            await fetch(`/api/titles/${id}/${type}`, { method: 'DELETE' });
            if(currentFilter === 'watched') fetchTitles(true, false);
            else if(currentFilter === 'wished') fetchTitles(false, true);
        } catch(e) { console.error(e); }
    };

    function escapeHTML(str) {
        return str.replace(/[&<>'"]/g, 
            tag => ({
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                "'": '&#39;',
                '"': '&quot;'
            }[tag] || tag)
        );
    }
});
