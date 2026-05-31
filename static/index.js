/**
 * API module handles all backend communication
 */
const ApiClient = {
    async request(url, options = {}) {
        try {
            const res = await fetch(url, options);
            if (!res.ok) {
                const data = await res.json().catch(() => ({}));
                throw new Error(data.message || 'Request failed');
            }
            return await res.json();
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },

    async fetchTitles(filter, page, pageSize) {
        let url = `/api/titles?page=${page}&page_size=${pageSize}`;
        if (filter === 'watched') url = `/api/titles/watched?page=${page}&page_size=${pageSize}`;
        else if (filter === 'wished') url = `/api/titles/search?wished=true&page=${page}&page_size=${pageSize}`;
        return this.request(url);
    },

    async searchTitles(query, page, pageSize) {
        return this.request(`/api/titles/search?title_names=${encodeURIComponent(query)}&page=${page}&page_size=${pageSize}&watched=true&wished=true`);
    },

    async fetchDetails(name, year) {
        let url = `/api/titles/details?movie_name=${encodeURIComponent(name)}`;
        if (year) url += `&release_year=${encodeURIComponent(year)}`;
        return this.request(url);
    },

    async fetchTitlesByIds(ids) {
        return this.request('/api/titles/bulk', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ids })
        });
    },

    async addTitle(titleData) {
        return this.request('/api/titles', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ titles: [titleData] })
        });
    },

    async updateTitleStatus(id, actionType, isRemove = false) {
        return this.request(`/api/titles/${id}/${actionType}`, { method: isRemove ? 'DELETE' : 'PUT' });
    },

    async deleteTitle(id) {
        return this.request(`/api/titles/${id}`, { method: 'DELETE' });
    }
};

/**
 * State module manages application state
 */
const AppState = {
    filter: 'watched',
    page: 1,
    pageSize: 15,
    searchQuery: '',
    currentManualData: null,
    
    resetPagination() {
        this.page = 1;
    }
};

/**
 * Main Application orchestrator
 */
class App {
    constructor() {
        this.elements = this.cacheDOM();
        this.searchTimeout = null;
        
        // Expose global actions for inline onclick handlers
        window.markAction = this.handleMarkAction.bind(this);
        window.closeModalById = (id) => this.closeModal(document.getElementById(id));
    }

    cacheDOM() {
        return {
            // Navigation
            navWatched: document.getElementById("nav-watched"),
            navWished: document.getElementById("nav-wished"),
            navGame: document.getElementById("nav-game"),

            // Main views
            controls: document.querySelector(".controls"),
            grid: document.getElementById("titlesGrid"),
            pagination: document.getElementById("pagination"),
            gameContainer: document.getElementById("gameContainer"),
            
            // Game UI
            gameLoadingState: document.getElementById("gameLoadingState"),
            gameContent: document.getElementById("gameContent"),
            gamePoster: document.getElementById("gamePoster"),
            gameTitle: document.getElementById("gameTitle"),
            gameYear: document.getElementById("gameYear"),
            gameRatingContainer: document.getElementById("gameRatingContainer"),
            gameRating: document.getElementById("gameRating"),
            gameGenres: document.getElementById("gameGenres"),
            gameOverview: document.getElementById("gameOverview"),
            btnGameSkip: document.getElementById("btnGameSkip"),
            btnGameWish: document.getElementById("btnGameWish"),
            btnGameWatch: document.getElementById("btnGameWatch"),

            // Buttons & Inputs
            searchInput: document.getElementById("searchInput"),
            btnPrev: document.getElementById("btnPrev"),
            btnNext: document.getElementById("btnNext"),
            pageInfo: document.getElementById("pageInfo"),

            // Modals
            addModal: document.getElementById("addModal"),
            detailsModal: document.getElementById("detailsModal"),
            
            // Details Elements
            btnCloseDetails: document.getElementById("btnCloseDetails"),
            detailsTitle: document.getElementById("detailsTitle"),
            detailsPoster: document.getElementById("detailsPoster"),
            detailsYear: document.getElementById("detailsYear"),
            detailsRating: document.getElementById("detailsRating"),
            detailsRatingValue: document.getElementById("detailsRatingValue"),
            detailsGenres: document.getElementById("detailsGenres"),
            detailsOverview: document.getElementById("detailsOverview"),
            detailsModalActions: document.getElementById("detailsModalActions"),
            
            // Manual Add Elements
            btnAddManual: document.getElementById("btnAddManual"),
            btnCancelAdd: document.getElementById("btnCancelAdd"),
            btnBackManual: document.getElementById("btnBackManual"),
            btnSearchManual: document.getElementById("btnSearchManual"),
            btnConfirmAddManual: document.getElementById("btnConfirmAddManual"),
            titleName: document.getElementById("titleName"),
            titleYear: document.getElementById("titleYear"),
            manualCheckboxGroup: document.getElementById("manualCheckboxGroup"),
            addManualWatched: document.getElementById("addManualWatched"),
            addManualWished: document.getElementById("addManualWished"),
            manualInputFields: document.getElementById("manualInputFields"),
            manualResultContainer: document.getElementById("manualResultContainer"),
            manualSearchActions: document.getElementById("manualSearchActions"),
            manualConfirmActions: document.getElementById("manualConfirmActions"),
            manualStatus: document.getElementById("manualStatus"),
            manualResultTitle: document.getElementById("manualResultTitle"),
            manualResultYear: document.getElementById("manualResultYear"),
            manualResultOverview: document.getElementById("manualResultOverview"),
            manualResultPoster: document.getElementById("manualResultPoster"),
        };
    }

    init() {
        this.bindEvents();
        this.loadTitles();
    }

    bindEvents() {
        const els = this.elements;

        // Navigation
        els.navWatched.addEventListener("click", () => this.switchTab("watched", els.navWatched));
        els.navWished.addEventListener("click", () => this.switchTab("wished", els.navWished));
        els.navGame.addEventListener("click", () => this.switchTab("game", els.navGame));

        // Game actions
        els.btnGameSkip.addEventListener("click", () => this.handleGameAction(false, false));
        els.btnGameWish.addEventListener("click", () => this.handleGameAction(false, true));
        els.btnGameWatch.addEventListener("click", () => this.handleGameAction(true, false));

        // Pagination
        els.btnPrev.addEventListener("click", () => {
            if (AppState.page > 1) {
                AppState.page--;
                this.executeFetch();
            }
        });
        els.btnNext.addEventListener("click", () => {
            AppState.page++;
            this.executeFetch();
        });

        // Search
        els.searchInput.addEventListener("input", (e) => this.handleLiveSearch(e.target.value));

        // Manual Add flow
        els.btnAddManual.addEventListener("click", () => this.openManualModal());
        els.btnCancelAdd.addEventListener("click", () => this.closeModal(els.addModal));
        els.btnBackManual.addEventListener("click", () => this.resetManualSearch());
        els.btnSearchManual.addEventListener("click", () => this.searchManualDetails());
        els.btnConfirmAddManual.addEventListener("click", () => this.confirmAddManual());

        // Details Flow
        els.btnCloseDetails.addEventListener("click", () => this.closeModal(els.detailsModal));
    }

    // --- State Management & UI Updates ---

    switchTab(filter, activeNavEl) {
        const els = this.elements;
        [els.navWatched, els.navWished, els.navGame].forEach(n => n.classList.remove("active"));
        activeNavEl.classList.add("active");
        
        if (filter === "game") {
            AppState.filter = "game";
            els.controls.style.display = "none";
            els.grid.style.display = "none";
            els.pagination.style.display = "none";
            els.gameContainer.style.display = "flex";
            this.loadNextGameMovie();
        } else {
            AppState.filter = filter;
            AppState.resetPagination();
            els.controls.style.display = "flex";
            els.grid.style.display = "grid";
            els.pagination.style.display = "flex";
            els.gameContainer.style.display = "none";
            AppState.searchQuery = "";
            els.searchInput.value = "";
            this.executeFetch();
        }
    }

    handleLiveSearch(value) {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            AppState.searchQuery = value.trim();
            AppState.resetPagination();
            this.executeFetch();
        }, 300);
    }

    executeFetch() {
        if (AppState.searchQuery) {
            this.performSearch(AppState.searchQuery);
        } else {
            this.loadTitles();
        }
    }

    async loadTitles() {
        try {
            const data = await ApiClient.fetchTitles(AppState.filter, AppState.page, AppState.pageSize);
            const titles = data.titles || data.watched_titles || [];
            this.renderGrid(titles);
            this.updatePaginationUI(data.total);
        } catch (e) {
            this.renderError();
        }
    }

    async performSearch(query) {
        try {
            const data = await ApiClient.searchTitles(query, AppState.page, AppState.pageSize);
            this.renderGrid(data.titles || []);
            this.updatePaginationUI(data.total);
        } catch (e) {
            this.renderError();
        }
    }

    // --- Manual Add Flow ---

    openManualModal() {
        const els = this.elements;
        els.addManualWatched.checked = AppState.filter !== "wished";
        els.addManualWished.checked = AppState.filter === "wished";
        
        this.resetManualSearch();
        els.titleName.value = "";
        els.titleYear.value = "";
        
        els.addModal.classList.add("active");
    }

    resetManualSearch() {
        const els = this.elements;
        els.manualInputFields.style.display = "";
        els.manualCheckboxGroup.style.display = "none";
        els.manualResultContainer.style.display = "none";
        els.manualSearchActions.style.display = "flex";
        els.manualConfirmActions.style.display = "none";
        els.manualStatus.innerText = "";
        AppState.currentManualData = null;
    }

    async searchManualDetails() {
        const els = this.elements;
        const name = els.titleName.value.trim();
        const year = els.titleYear.value.trim();
        
        if (!name) return alert("Name is required");
        
        els.manualStatus.innerText = "Fetching metadata from TMDB...";
        try {
            const data = await ApiClient.fetchDetails(name, year);
            if (data.status === "success" && data.metadata) {
                AppState.currentManualData = data.metadata;
                els.manualStatus.innerText = "";
                
                this.populatePreviewCard(
                    data.metadata,
                    els.manualResultTitle,
                    els.manualResultYear,
                    els.manualResultOverview,
                    els.manualResultPoster
                );
                
                els.manualInputFields.style.display = "none";
                els.manualCheckboxGroup.style.display = "flex";
                els.manualResultContainer.style.display = "flex";
                els.manualSearchActions.style.display = "none";
                els.manualConfirmActions.style.display = "flex";
            } else {
                els.manualStatus.innerText = data.message || "Not found.";
            }
        } catch (e) {
            els.manualStatus.innerText = "Network error.";
        }
    }

    async confirmAddManual() {
        if (!AppState.currentManualData) return;
        const els = this.elements;
        els.manualStatus.innerText = "Adding to database...";
        
        try {
            await this.addTitleToDB(
                AppState.currentManualData,
                els.addManualWatched.checked,
                els.addManualWished.checked
            );
            this.closeModal(els.addModal);
            this.loadTitles();
        } catch (e) {
            els.manualStatus.innerText = `Error: ${e.message}`;
        }
    }



    // --- Shared Utilities ---

    async openDetails(id) {
        try {
            const data = await ApiClient.fetchTitlesByIds([id]);
            if (data.status === "success" && data.titles && data.titles.length > 0) {
                this.renderDetailsModal(data.titles[0]);
            }
        } catch (e) {
            console.error("Failed to fetch details", e);
        }
    }

    renderDetailsModal(title) {
        const els = this.elements;
        els.detailsTitle.innerText = title.name;
        els.detailsYear.innerText = title.release_year || "Unknown Year";
        
        if (title.poster_path) {
            els.detailsPoster.src = `https://image.tmdb.org/t/p/w300${title.poster_path}`;
            els.detailsPoster.style.display = "block";
        } else {
            els.detailsPoster.style.display = "none";
        }

        if (title.imdb_rating) {
            els.detailsRatingValue.innerText = title.imdb_rating.toFixed(1);
            els.detailsRating.style.display = "inline-flex";
        } else {
            els.detailsRating.style.display = "none";
        }

        els.detailsGenres.innerHTML = "";
        if (title.genres && title.genres.length > 0) {
            title.genres.forEach(g => {
                const span = document.createElement("span");
                span.className = "badge";
                span.style.background = "rgba(255, 255, 255, 0.1)";
                span.innerText = g;
                els.detailsGenres.appendChild(span);
            });
        }

        els.detailsOverview.innerText = title.overview || "No overview available.";

        let modalActionsHtml = "";
        if (title.watched) {
            modalActionsHtml = `
                <button class="btn btn-danger" onclick="markAction(${title.id}, 'watch', true); closeModalById('detailsModal')" style="flex: 1;">
                    <ion-icon name="trash-outline" style="vertical-align: text-bottom; margin-right: 5px;"></ion-icon>
                    Remove from Watched
                </button>
            `;
        } else if (title.wished) {
            modalActionsHtml = `
                <button class="btn" onclick="markAction(${title.id}, 'watch', false); closeModalById('detailsModal')" style="flex: 1;">
                    <ion-icon name="eye-outline" style="vertical-align: text-bottom; margin-right: 5px;"></ion-icon>
                    Mark as Watched
                </button>
                <button class="btn btn-danger" onclick="markAction(${title.id}, 'wish', true); closeModalById('detailsModal')" style="flex: 1;">
                    <ion-icon name="trash-outline" style="vertical-align: text-bottom; margin-right: 5px;"></ion-icon>
                    Remove from Wishlist
                </button>
            `;
        }
        els.detailsModalActions.innerHTML = modalActionsHtml;

        els.detailsModal.classList.add("active");
    }

    closeModal(modalEl) {
        modalEl.classList.remove("active");
    }

    populatePreviewCard(metadata, titleEl, yearEl, overviewEl, posterEl) {
        const title = metadata.title || metadata.name;
        const releaseYear = metadata.release_date
            ? metadata.release_date.split("-")[0]
            : metadata.first_air_date
                ? metadata.first_air_date.split("-")[0]
                : "Unknown Year";

        titleEl.innerText = title;
        yearEl.innerText = releaseYear;
        overviewEl.innerText = metadata.overview || "No overview available.";

        if (metadata.poster_path) {
            posterEl.src = `https://image.tmdb.org/t/p/w200${metadata.poster_path}`;
            posterEl.style.display = "block";
        } else {
            posterEl.style.display = "none";
        }
    }

    async addTitleToDB(metadata, isWatched, isWished) {
        const releaseYear = metadata.release_date
            ? parseInt(metadata.release_date.split("-")[0])
            : metadata.first_air_date
                ? parseInt(metadata.first_air_date.split("-")[0])
                : 0;

        const titleData = {
            name: metadata.title || metadata.name,
            release_year: releaseYear,
            overview: metadata.overview,
            imdb_id: metadata.imdb_id,
            tmdb_id: String(metadata.id || ''),
            imdb_rating: metadata.vote_average,
            poster_path: metadata.poster_path,
            backdrop_path: metadata.backdrop_path,
            genres: metadata.genres ? metadata.genres.map((g) => g.name) : [],
            watched: isWatched,
            wished: isWished,
        };

        return ApiClient.addTitle(titleData);
    }

    async handleMarkAction(id, type, isRemove = false) {
        try {
            await ApiClient.updateTitleStatus(id, type, isRemove);
            this.executeFetch();
        } catch (e) {
            console.error("Action failed", e);
        }
    }


    // --- Game Logic ---

    async loadNextGameMovie() {
        const els = this.elements;
        els.gameContent.style.display = "none";
        els.gameLoadingState.style.display = "flex";

        try {
            const data = await ApiClient.request('/api/titles/game/next');
            if (data.status === "error") {
                els.gameLoadingState.innerHTML = `<p>${data.message || "Could not find a new movie right now."}</p>`;
                return;
            }
            
            const meta = data.metadata;
            this.currentGameMetadata = meta;

            // Render game UI
            els.gameTitle.innerText = meta.title || meta.name;
            const releaseYear = meta.release_date
                ? meta.release_date.split("-")[0]
                : meta.first_air_date
                    ? meta.first_air_date.split("-")[0]
                    : "Unknown Year";
            els.gameYear.innerText = releaseYear;

            if (meta.poster_path) {
                els.gamePoster.src = `https://image.tmdb.org/t/p/w300${meta.poster_path}`;
                els.gamePoster.style.display = "block";
            } else {
                els.gamePoster.style.display = "none";
            }

            if (meta.vote_average) {
                els.gameRating.innerText = meta.vote_average.toFixed(1);
                els.gameRatingContainer.style.display = "inline-flex";
            } else {
                els.gameRatingContainer.style.display = "none";
            }

            els.gameGenres.innerHTML = "";
            if (meta.genres && meta.genres.length > 0) {
                meta.genres.forEach(g => {
                    const span = document.createElement("span");
                    span.className = "badge";
                    span.style.background = "rgba(255, 255, 255, 0.1)";
                    span.innerText = g.name || g;
                    els.gameGenres.appendChild(span);
                });
            }

            els.gameOverview.innerText = meta.overview || "No overview available.";

            // Show content
            els.gameLoadingState.style.display = "none";
            els.gameContent.style.display = "block";

        } catch (e) {
            console.error("Game load failed", e);
            els.gameLoadingState.innerHTML = `<p>Error loading movie game.</p>`;
        }
    }

    async handleGameAction(isWatched, isWished) {
        if (!this.currentGameMetadata) return;
        const els = this.elements;
        
        // Show loading briefly
        els.gameContent.style.display = "none";
        els.gameLoadingState.innerHTML = `<div class="loader"></div><p>Saving to your list...</p>`;
        els.gameLoadingState.style.display = "flex";

        try {
            await this.addTitleToDB(this.currentGameMetadata, isWatched, isWished);
            this.currentGameMetadata = null;
            // Load next immediately
            this.loadNextGameMovie();
        } catch (e) {
            console.error("Failed to add game title", e);
            els.gameLoadingState.innerHTML = `<p>Failed to add title. Retrying game...</p>`;
            setTimeout(() => this.loadNextGameMovie(), 2000);
        }
    }

    // --- UI Rendering ---

    updatePaginationUI(total) {
        const { pagination, btnPrev, btnNext, pageInfo } = this.elements;
        
        if (!total || total <= AppState.pageSize) {
            pagination.style.display = "none";
        } else {
            pagination.style.display = "flex";
            const totalPages = Math.ceil(total / AppState.pageSize);
            pageInfo.innerText = `Page ${AppState.page} of ${totalPages}`;
            btnPrev.disabled = AppState.page === 1;
            btnNext.disabled = AppState.page === totalPages;
        }
    }

    renderGrid(titles) {
        const grid = this.elements.grid;
        grid.innerHTML = "";
        
        if (!titles || titles.length === 0) {
            grid.innerHTML = `
                <div class="empty-state">
                    <h3>No titles found</h3>
                    <p>Try adding some new movies or TV shows!</p>
                </div>
            `;
            return;
        }

        titles.forEach((t) => grid.appendChild(this.createTitleCard(t)));
    }

    createTitleCard(t) {
        const card = document.createElement("div");
        card.className = "card";
        card.style.cursor = "pointer";

        card.addEventListener("click", (e) => {
            if (!e.target.closest(".card-actions")) {
                this.openDetails(t.id);
            }
        });

        let badgesHtml = "";
        if (t.watched) badgesHtml += `<span class="badge watched">Watched</span>`;
        if (t.wished) badgesHtml += `<span class="badge wished">Wishlist</span>`;

        let imgHtml = "";
        if (t.poster_path) {
            imgHtml = `<img src="https://image.tmdb.org/t/p/w500${t.poster_path}" alt="${this.escapeHTML(t.name)} poster" class="card-image" loading="lazy">`;
        } else {
            imgHtml = `
                <div class="card-image" style="display:flex; align-items:center; justify-content:center; color:var(--text-muted);">
                    <ion-icon name="film-outline" style="font-size: 4rem; opacity: 0.5;"></ion-icon>
                </div>`;
        }

        let primaryActionHtml = "";
        let deleteActionType = "watch";

        if (t.watched) {
            deleteActionType = "watch";
        } else if (t.wished) {
            primaryActionHtml = `
                <button class="icon-btn watched" onclick="markAction(${t.id}, 'watch', false)" title="Mark as Watched">
                    <ion-icon name="eye-outline"></ion-icon>
                </button>
            `;
            deleteActionType = "wish";
        }

        card.innerHTML = `
            ${imgHtml}
            <div class="card-content">
                <div class="card-title">${this.escapeHTML(t.name)}</div>
                <div class="card-year">${t.release_year || "Unknown Year"}</div>
                <div class="badges">${badgesHtml}</div>
                <div class="card-actions">
                    ${primaryActionHtml}
                    <div style="flex:1"></div>
                    <button class="icon-btn delete" onclick="markAction(${t.id}, '${deleteActionType}', true)" title="Remove from List">
                        <ion-icon name="trash-outline"></ion-icon>
                    </button>
                </div>
            </div>
        `;
        return card;
    }

    renderError() {
        this.elements.grid.innerHTML = `
            <div class="empty-state">
                <h3>Oops! Something went wrong.</h3>
                <p>Could not connect to the server.</p>
            </div>
        `;
    }

    escapeHTML(str) {
        return str.replace(/[&<>'"]/g, (tag) => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            "'": "&#39;",
            '"': "&quot;",
        })[tag] || tag);
    }
}

// Bootstrap application
document.addEventListener("DOMContentLoaded", () => {
    const app = new App();
    app.init();
});
