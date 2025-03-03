import { getCountryName } from './countryMap.js';

class App {
    constructor() {
        this.container = document.getElementById('app');
        this.templates = {};
        this.currentFilters = {
            country: '',
            search: ''
        };
        this.init();
    }

    async init() {
        await this.loadTemplates();
        window.addEventListener('popstate', () => this.handleRoute());
        this.handleRoute();
    }

    async loadTemplates() {
        try {
            const [landingResponse, dashboardResponse, aboutResponse, relatedWorkResponse] = await Promise.all([
                fetch('/static/pages/landing.html'),
                fetch('/static/pages/dashboard.html'),
                fetch('/static/pages/about.html'),
                fetch('/static/pages/related-work.html')
            ]);
            
            this.templates.landing = await landingResponse.text();
            this.templates.dashboard = await dashboardResponse.text();
            this.templates.about = await aboutResponse.text();
            this.templates.relatedWork = await relatedWorkResponse.text();
        } catch (error) {
            console.error('Error loading templates:', error);
        }
    }
    async handleRoute() {
        const path = window.location.pathname;
        
        switch(path) {
            case '/':
                await this.renderLandingPage();
                break;
            case '/dashboard':
                await this.renderDashboard();
                break;
            case '/about':
                await this.renderAboutPage();
                break;
            case '/related-work':
                await this.renderRelatedWork();
                break;
            default:
                this.renderNotFound();
        }
    }

    async renderLandingPage() {
        try {
            const response = await fetch('/api/accounts');
            if (!response.ok) throw new Error('Failed to fetch landing page data');
            
            const data = await response.json();

            // Get the last update time from the first account (most recent)
            const lastUpdate = data.accounts && data.accounts.length > 0 
            ? new Date(data.accounts[0].last_reported_at).toLocaleString()
            : 'N/A';
            
            let template = this.templates.landing;
            template = template.replace('{{totalAccounts}}', data.totalCount);
            template = template.replace('{{totalCountries}}', data.uniqueCountries.length);
            template = template.replace('{{lastUpdate}}', lastUpdate);
            
            this.container.innerHTML = template;
        } catch (error) {
            console.error('Error rendering landing page:', error);
            this.renderError('Failed to load landing page data');
        }
    }

   
    async renderDashboard(page = 1) {
        try {
            // Construct URL with filters
            const params = new URLSearchParams({
                page: page.toString(),
                ...this.currentFilters
            });
    
            const response = await fetch(`/api/accounts?${params}`);
            if (!response.ok) throw new Error('Failed to fetch dashboard data');
            const data = await response.json();

            // Get the last update time from the first account (most recent)
            const lastUpdate = data.accounts && data.accounts.length > 0 
            ? new Date(data.accounts[0].last_reported_at).toLocaleString()
            : 'N/A';
    
            // Update the template with the data
            let template = this.templates.dashboard;
            template = template.replace('{{lastUpdate}}', lastUpdate);
            template = template.replace('{{totalAccounts}}', data.totalCount);
            template = template.replace('{{totalCountries}}', data.uniqueCountries.length);
            template = template.replace('{{currentPage}}', data.currentPage);
            template = template.replace('{{totalPages}}', data.totalPages);
    
            // Set the main template
            this.container.innerHTML = template;
    
            // Populate country filter options
            const countryFilter = document.getElementById('countryFilter');
            if (countryFilter) {
                // Sort countries by their full names
                const countryOptions = data.uniqueCountries
                    .sort((a, b) => getCountryName(a).localeCompare(getCountryName(b)))
                    .map(code => ({
                        code: code,
                        name: getCountryName(code)
                    }));
    
                countryFilter.innerHTML = '<option value="">All Countries</option>' +
                    countryOptions.map(country => 
                        `<option value="${country.code}">${country.name}</option>`
                    ).join('');
                
                // Restore selected value if there was one
                countryFilter.value = this.currentFilters.country;
            }
    
            // Restore search filter value
            const searchFilter = document.getElementById('searchFilter');
            if (searchFilter) {
                searchFilter.value = this.currentFilters.search;
            }
    
            // Render accounts list
            const accountsList = document.getElementById('accountsList');
            if (accountsList) {
                if (!data.accounts || data.accounts.length === 0) {
                    accountsList.innerHTML = '<tr><td colspan="3">No accounts found</td></tr>';
                } else {
                    const rows = data.accounts.map(account => {
                        const username = account.name;
                        const countries = account.countries
                            .map(code => `<span class="country-item">${getCountryName(code)}</span>`)
                            .join(' ');
                        const date = new Date(account.last_reported_at).toLocaleString();
                        
                        return `
                            <tr>
                                <td class="username-col">@${username}</td>
                                <td class="countries-col">${countries}</td>
                                <td class="date-col">${date}</td>
                            </tr>
                        `;
                    });
                    accountsList.innerHTML = rows.join('\n');
                }
            }
    
            // Render pagination
            const paginationControls = document.getElementById('paginationControls');
            if (paginationControls) {
                const pages = [];
                
                // Previous button
                pages.push(`<a href="#" class="page-link ${data.currentPage === 1 ? 'disabled' : ''}" data-page="${data.currentPage - 1}">[PREV]</a>`);
                
                // Previous page number (if not first page)
                if (data.currentPage > 1) {
                    pages.push(`<a href="#" class="page-link" data-page="${data.currentPage - 1}">[${data.currentPage - 1}]</a>`);
                }
                
                // Current page
                pages.push(`<a href="#" class="page-link active" data-page="${data.currentPage}">[${data.currentPage}]</a>`);
                
                // Next page number (if not last page)
                if (data.currentPage < data.totalPages) {
                    pages.push(`<a href="#" class="page-link" data-page="${data.currentPage + 1}">[${data.currentPage + 1}]</a>`);
                }
                
                // Next button
                pages.push(`<a href="#" class="page-link ${data.currentPage === data.totalPages ? 'disabled' : ''}" data-page="${data.currentPage + 1}">[NEXT]</a>`);
                
                paginationControls.innerHTML = pages.join('');
                this.setupPaginationListeners();
            }
            
            // Setup filter listeners
            this.setupFilterListeners();
    
        } catch (error) {
            console.error('Error rendering dashboard:', error);
            this.renderError('Failed to load dashboard data');
        }
    }

    async renderAboutPage() {
        try {
            const template = this.templates.about;
            this.container.innerHTML = template;
        } catch (error) {
            console.error('Error rendering about page:', error);
            this.renderError('Failed to load about page');
        }
    }
    
    async renderRelatedWork() {
        try {
            const template = this.templates.relatedWork;
            this.container.innerHTML = template;
        } catch (error) {
            console.error('Error rendering related work page:', error);
            this.renderError('Failed to load related work page');
        }
    }

    setupPaginationListeners() {
        const controls = document.getElementById('paginationControls');
        if (!controls) return;
    
        controls.addEventListener('click', (e) => {
            e.preventDefault();
            const link = e.target.closest('.page-link');
            if (!link || link.classList.contains('disabled')) return;
    
            const page = parseInt(link.dataset.page);
            this.renderDashboard(page);
        });
    }
    
    setupFilterListeners() {
        const countryFilter = document.getElementById('countryFilter');
        const searchFilter = document.getElementById('searchFilter');
        const searchButton = document.getElementById('searchButton');
        const clearFilters = document.getElementById('clearFilters');
    
        if (countryFilter) {
            countryFilter.addEventListener('change', () => {
                this.currentFilters.country = countryFilter.value;
                this.renderDashboard(1);
            });
        }
    
        // Add enter key support for search
        if (searchFilter) {
            searchFilter.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.currentFilters.search = searchFilter.value;
                    this.renderDashboard(1);
                }
            });
        }
    
        // Add search button handler
        if (searchButton) {
            searchButton.addEventListener('click', () => {
                const searchValue = document.getElementById('searchFilter').value;
                this.currentFilters.search = searchValue;
                this.renderDashboard(1);
            });
        }
    
        if (clearFilters) {
            clearFilters.addEventListener('click', () => {
                this.currentFilters = {
                    country: '',
                    search: ''
                };
                if (countryFilter) countryFilter.value = '';
                if (searchFilter) searchFilter.value = '';
                this.renderDashboard(1);
            });
        }
    }
    
    renderError(message) {
        this.container.innerHTML = `
            <div class="error-container">
                <div class="data-block">
                    <div class="block-header">>_ Error</div>
                    <pre class="data-display">${message}</pre>
                    <a href="/" class="action-link">>_ Return Home</a>
                </div>
            </div>
        `;
    }

    renderNotFound() {
        this.renderError('Page not found');
    }
}

// Initialize the app
const app = new App();