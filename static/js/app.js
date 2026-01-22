// Main Application Logic
const App = {
    currentTaskId: null,
    progressInterval: null,
    currentComments: [],

    init() {
        this.bindEvents();
        this.setupAuthSwitch();
    },

    bindEvents() {
        document.getElementById('start-btn').addEventListener('click', () => this.startScraping());
        document.getElementById('filter-btn').addEventListener('click', () => this.applyFilter());
        document.getElementById('export-btn').addEventListener('click', () => this.exportComments());
    },

    setupAuthSwitch() {
        const authRadios = document.querySelectorAll('input[name="auth-type"]');
        const cookieGroup = document.getElementById('cookie-auth-group');
        const appGroup = document.getElementById('app-auth-group');

        authRadios.forEach(radio => {
            radio.addEventListener('change', (e) => {
                cookieGroup.style.display = e.target.value === 'cookie' ? 'block' : 'none';
                appGroup.style.display = e.target.value === 'app' ? 'block' : 'none';
            });
        });
    },

    async startScraping() {
        try {
            const videoInput = document.getElementById('video-input').value.trim();
            if (!videoInput) {
                this.showError('请输入视频BV号或链接');
                return;
            }

            const authType = document.querySelector('input[name="auth-type"]:checked').value;
            const cookie = authType === 'cookie' ? document.getElementById('sessdata').value : '';
            const appKey = authType === 'app' ? document.getElementById('app-key').value : '';
            const appSecret = authType === 'app' ? document.getElementById('app-secret').value : '';
            const pageLimit = parseInt(document.getElementById('page-limit').value) || 10;
            const delayMs = parseInt(document.getElementById('delay').value) || 300;
            const sortMode = document.querySelector('input[name="sort-mode"]:checked').value;
            const includeReplies = document.getElementById('include-replies').checked;

            const startBtn = document.getElementById('start-btn');
            startBtn.disabled = true;
            startBtn.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinner">
                    <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                </svg>
                启动中...
            `;

            const response = await API.startScrape({
                video_id: videoInput,
                auth_type: authType,
                cookie: cookie,
                app_key: appKey,
                app_secret: appSecret,
                page_limit: pageLimit,
                delay_ms: delayMs,
                sort_mode: sortMode,
                include_replies: includeReplies
            });

            this.currentTaskId = response.task_id;

            document.getElementById('progress-section').style.display = 'block';
            document.getElementById('results-section').style.display = 'none';

            this.startProgressPolling();

        } catch (error) {
            this.showError(error.message);
            this.resetStartButton();
        }
    },

    startProgressPolling() {
        if (this.progressInterval) {
            clearInterval(this.progressInterval);
        }

        this.progressInterval = setInterval(async () => {
            try {
                const progress = await API.getProgress(this.currentTaskId);
                this.updateProgress(progress);

                if (progress.status === 'completed') {
                    clearInterval(this.progressInterval);
                    this.onScrapeCompleted();
                } else if (progress.status === 'failed') {
                    clearInterval(this.progressInterval);
                    this.showError(progress.error || '爬取失败');
                    this.resetStartButton();
                }
            } catch (error) {
                console.error('获取进度失败:', error);
                clearInterval(this.progressInterval);
                this.showError('获取进度失败');
                this.resetStartButton();
            }
        }, 1000);
    },

    updateProgress(progress) {
        if (progress.video_title) {
            document.getElementById('video-title').textContent = progress.video_title;
        }

        const percent = progress.progress.page_limit > 0
            ? (progress.progress.current_page / progress.progress.page_limit) * 100
            : 0;
        document.getElementById('progress-fill').style.width = percent + '%';

        document.getElementById('progress-text').textContent =
            `正在爬取第 ${progress.progress.current_page}/${progress.progress.page_limit} 页...`;

        document.getElementById('current-page').textContent = progress.progress.current_page;
        document.getElementById('total-comments').textContent = progress.progress.total_comments;
        document.getElementById('elapsed-time').textContent = progress.elapsed_seconds + 's';
    },

    async onScrapeCompleted() {
        try {
            document.getElementById('progress-fill').style.width = '100%';
            document.getElementById('progress-text').textContent = '爬取完成！';

            await this.loadResults();

            document.getElementById('results-section').style.display = 'block';

            this.resetStartButton();

        } catch (error) {
            this.showError('加载结果失败: ' + error.message);
            this.resetStartButton();
        }
    },

    async loadResults() {
        const sortBy = document.getElementById('sort-select').value;
        const keyword = document.getElementById('keyword-input').value.trim();

        const results = await API.getResults(this.currentTaskId, {
            sort: sortBy,
            keyword: keyword,
            limit: 1000
        });

        this.currentComments = results.comments;

        const summaryEl = document.getElementById('result-summary');
        if (summaryEl) {
            summaryEl.innerHTML = `共获取 <strong>${results.total_count}</strong> 条评论`;
        }

        this.renderCommentsTable(results.comments);
        Charts.renderAll(results.comments);
    },

    renderCommentsTable(comments) {
        const container = document.getElementById('comments-list');
        container.innerHTML = '';

        if (comments.length === 0) {
            container.innerHTML = '<div class="empty-state"><p style="color: var(--gray-400);">暂无评论</p></div>';
            return;
        }

        comments.forEach((comment, index) => {
            this.renderCommentCard(container, comment, index);
        });
    },

    renderCommentCard(container, comment, index) {
        const hasReplies = comment.replies && comment.replies.length > 0;
        const replyCount = hasReplies ? comment.replies.length : 0;

        const avatar = comment.avatar || '';
        const author = comment.author || '未知用户';
        const content = comment.content || '';
        const likes = comment.likes || 0;
        const time = comment.time || '';
        const levelInfo = comment.level || 0;

        const card = document.createElement('div');
        card.className = 'comment-card';

        let cardHtml = `
            <div class="comment-header">
                <img src="${avatar}" alt="${author}" class="comment-avatar">
                <div class="comment-user-info">
                    <div class="comment-author">${this.escapeHtml(author)}</div>
                    <span class="comment-level">Lv${levelInfo}</span>
                    <div class="comment-meta">
                        <span class="comment-likes">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                            </svg>
                            ${likes}
                        </span>
                        <span class="comment-time">${time}</span>
                    </div>
                </div>
            </div>
            <div class="comment-content">${this.escapeHtml(content)}</div>
        `;

        if (hasReplies) {
            cardHtml += `
                <button class="expand-btn" data-index="${index}">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M4.427 7.427l3.396 3.396a.25.25 0 00.354 0l3.396-3.396A.25.25 0 0011.396 7H4.604a.25.25 0 00-.177.427z"/>
                    </svg>
                    <span>${replyCount} 条回复</span>
                </button>
                <div class="replies-container" data-parent-index="${index}" style="display: none;">
                    <div class="replies-list">
                        ${comment.replies.map(reply => this.renderReplyCard(reply)).join('')}
                    </div>
                </div>
            `;
        }

        card.innerHTML = cardHtml;
        container.appendChild(card);

        if (hasReplies) {
            const expandBtn = card.querySelector('.expand-btn');
            const repliesContainer = card.querySelector('.replies-container');

            expandBtn.addEventListener('click', () => {
                const isExpanded = repliesContainer.style.display !== 'none';
                repliesContainer.style.display = isExpanded ? 'none' : 'block';
                expandBtn.classList.toggle('expanded', !isExpanded);
                expandBtn.querySelector('span').textContent = isExpanded ? `${replyCount} 条回复` : '收起回复';
            });
        }
    },

    renderReplyCard(reply) {
        const avatar = reply.avatar || '';
        const author = reply.author || '未知用户';
        const content = reply.content || '';
        const likes = reply.likes || 0;
        const time = reply.time || '';
        const levelInfo = reply.level || 0;

        return `
            <div class="reply-card">
                <img src="${avatar}" alt="${author}" class="reply-avatar">
                <div class="reply-content">
                    <div class="reply-header">
                        <span class="reply-author">${this.escapeHtml(author)}</span>
                        <span class="comment-level" style="font-size: 10px;">Lv${levelInfo}</span>
                        <span class="reply-likes">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                                <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                            </svg>
                            ${likes}
                        </span>
                    </div>
                    <div class="reply-text">${this.escapeHtml(content)}</div>
                    <div class="reply-time">${time}</div>
                </div>
            </div>
        `;
    },

    async applyFilter() {
        if (!this.currentTaskId) {
            this.showError('请先完成爬取');
            return;
        }

        try {
            document.getElementById('filter-btn').disabled = true;
            document.getElementById('filter-btn').textContent = '应用中...';

            await this.loadResults();

            document.getElementById('filter-btn').disabled = false;
            document.getElementById('filter-btn').textContent = '应用筛选';
        } catch (error) {
            this.showError('筛选失败: ' + error.message);
            document.getElementById('filter-btn').disabled = false;
            document.getElementById('filter-btn').textContent = '应用筛选';
        }
    },

    async exportComments() {
        if (!this.currentTaskId) {
            this.showError('请先完成爬取');
            return;
        }

        try {
            const format = document.getElementById('export-format').value;
            const filename = document.getElementById('export-filename').value.trim();
            const sortBy = document.getElementById('sort-select').value;

            document.getElementById('export-btn').disabled = true;
            document.getElementById('export-btn').textContent = '导出中...';

            const result = await API.exportComments({
                task_id: this.currentTaskId,
                format: format,
                sort: sortBy,
                filename: filename
            });

            const downloadLink = document.getElementById('download-link');
            const downloadUrl = document.getElementById('download-url');
            downloadUrl.href = result.download_url;
            downloadUrl.textContent = result.filename;
            downloadLink.classList.add('show');

            document.getElementById('export-btn').disabled = false;
            document.getElementById('export-btn').innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="7 10 12 15 17 10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                </svg>
                导出
            `;

        } catch (error) {
            this.showError('导出失败: ' + error.message);
            document.getElementById('export-btn').disabled = false;
            document.getElementById('export-btn').innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="7 10 12 15 17 10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                </svg>
                导出
            `;
        }
    },

    showError(message) {
        const errorDiv = document.getElementById('error-message');
        if (!errorDiv) {
            console.error('Error element not found:', message);
            return;
        }
        errorDiv.textContent = message;
        errorDiv.style.display = 'block';

        setTimeout(() => {
            errorDiv.style.display = 'none';
        }, 5000);
    },

    resetStartButton() {
        const startBtn = document.getElementById('start-btn');
        startBtn.disabled = false;
        startBtn.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            开始爬取
        `;
    },

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});
