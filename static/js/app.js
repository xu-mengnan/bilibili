// 主应用逻辑
const App = {
    currentTaskId: null,
    progressInterval: null,
    currentComments: [],

    /**
     * 初始化应用
     */
    init() {
        this.bindEvents();
        this.setupAuthSwitch();
    },

    /**
     * 绑定事件
     */
    bindEvents() {
        // 开始爬取
        document.getElementById('start-btn').addEventListener('click', () => this.startScraping());

        // 筛选应用
        document.getElementById('filter-btn').addEventListener('click', () => this.applyFilter());

        // 导出
        document.getElementById('export-btn').addEventListener('click', () => this.exportComments());
    },

    /**
     * 设置认证方式切换
     */
    setupAuthSwitch() {
        const authRadios = document.querySelectorAll('input[name="auth-type"]');
        const cookieAuth = document.getElementById('cookie-auth');
        const appAuth = document.getElementById('app-auth');

        authRadios.forEach(radio => {
            radio.addEventListener('change', (e) => {
                cookieAuth.style.display = e.target.value === 'cookie' ? 'block' : 'none';
                appAuth.style.display = e.target.value === 'app' ? 'block' : 'none';
            });
        });
    },

    /**
     * 启动爬取
     */
    async startScraping() {
        try {
            // 获取表单数据
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

            // 禁用开始按钮
            const startBtn = document.getElementById('start-btn');
            startBtn.disabled = true;
            startBtn.textContent = '启动中...';

            // 启动爬取
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

            // 显示进度区域
            document.getElementById('progress-section').style.display = 'block';
            document.getElementById('results-section').style.display = 'none';

            // 开始轮询进度
            this.startProgressPolling();

        } catch (error) {
            this.showError(error.message);
            document.getElementById('start-btn').disabled = false;
            document.getElementById('start-btn').textContent = '开始爬取';
        }
    },

    /**
     * 开始轮询进度
     */
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
                    this.resetUI();
                }
            } catch (error) {
                console.error('获取进度失败:', error);
                clearInterval(this.progressInterval);
                this.showError('获取进度失败');
                this.resetUI();
            }
        }, 1000);
    },

    /**
     * 更新进度UI
     */
    updateProgress(progress) {
        // 更新视频标题
        if (progress.video_title) {
            document.getElementById('video-title').textContent = progress.video_title;
        }

        // 更新进度条
        const percent = progress.progress.page_limit > 0
            ? (progress.progress.current_page / progress.progress.page_limit) * 100
            : 0;
        document.getElementById('progress-fill').style.width = percent + '%';

        // 更新文本
        document.getElementById('progress-text').textContent =
            `正在爬取第 ${progress.progress.current_page}/${progress.progress.page_limit} 页...`;

        // 更新统计
        document.getElementById('current-page').textContent = progress.progress.current_page;
        document.getElementById('total-comments').textContent = progress.progress.total_comments;
        document.getElementById('elapsed-time').textContent = progress.elapsed_seconds + 's';
    },

    /**
     * 爬取完成处理
     */
    async onScrapeCompleted() {
        try {
            // 更新进度UI
            document.getElementById('progress-fill').style.width = '100%';
            document.getElementById('progress-text').textContent = '爬取完成！';

            // 加载结果
            await this.loadResults();

            // 显示结果区域
            document.getElementById('results-section').style.display = 'block';

            // 重置开始按钮
            this.resetUI();

        } catch (error) {
            this.showError('加载结果失败: ' + error.message);
            this.resetUI();
        }
    },

    /**
     * 加载结果
     */
    async loadResults() {
        const sortBy = document.getElementById('sort-select').value;
        const keyword = document.getElementById('keyword-input').value.trim();

        const results = await API.getResults(this.currentTaskId, {
            sort: sortBy,
            keyword: keyword,
            limit: 1000
        });

        this.currentComments = results.comments;

        // 更新评论数量
        const summaryEl = document.getElementById('result-summary');
        if (summaryEl) {
            summaryEl.innerHTML = `共获取 <strong>${results.total_count}</strong> 条评论`;
        }

        // 渲染评论表格
        this.renderCommentsTable(results.comments);

        // 渲染图表
        Charts.renderAll(results.comments);
    },

    /**
     * 渲染评论表格
     */
    renderCommentsTable(comments) {
        const container = document.getElementById('comments-list');
        container.innerHTML = '';

        if (comments.length === 0) {
            container.innerHTML = '<div style="text-align: center; padding: 40px; color: var(--gray-400);">暂无评论</div>';
            return;
        }

        comments.forEach((comment, index) => {
            // 渲染主评论卡片
            this.renderCommentCard(container, comment, index);
        });
    },

    /**
     * 渲染评论卡片
     */
    renderCommentCard(container, comment, index) {
        const hasReplies = comment.replies && comment.replies.length > 0;
        const replyCount = hasReplies ? comment.replies.length : 0;

        const avatar = comment.avatar || '';
        const author = comment.author || '未知用户';
        const content = comment.content || '';
        const likes = comment.likes || 0;
        const time = comment.time || '';
        const levelInfo = comment.level || 0;

        // 创建主评论卡片
        const card = document.createElement('div');
        card.className = 'comment-card';

        // 构建卡片 HTML
        let cardHtml = `
            <div class="comment-header">
                <div class="comment-user">
                    <img src="${avatar}" alt="${author}" class="comment-avatar">
                    <div class="comment-user-info">
                        <strong class="comment-author-name">${this.escapeHtml(author)}</strong>
                        <span class="comment-user-level">Lv${levelInfo}</span>
                    </div>
                </div>
                <div class="comment-meta">
                    <span class="comment-likes">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                        </svg>
                        ${likes}
                    </span>
                    <span class="comment-time">${time}</span>
                </div>
            </div>
            <div class="comment-content">${this.escapeHtml(content)}</div>
        `;

        // 如果有子评论，添加展开按钮和子评论容器
        if (hasReplies) {
            cardHtml += `
                <div class="comment-actions">
                    <button class="expand-btn" data-index="${index}">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                            <path d="M4.427 7.427l3.396 3.396a.25.25 0 00.354 0l3.396-3.396A.25.25 0 0011.396 7H4.604a.25.25 0 00-.177.427z"/>
                        </svg>
                        <span>${replyCount} 条回复</span>
                    </button>
                </div>
                <div class="replies-container" data-parent-index="${index}" style="display: none;">
                    <div class="replies-list">
                        ${comment.replies.map(reply => this.renderReplyCard(reply)).join('')}
                    </div>
                </div>
            `;
        }

        card.innerHTML = cardHtml;
        container.appendChild(card);

        // 绑定展开/折叠事件
        if (hasReplies) {
            const expandBtn = card.querySelector('.expand-btn');
            const repliesContainer = card.querySelector('.replies-container');
            const svg = expandBtn.querySelector('svg');
            const textSpan = expandBtn.querySelector('span');

            expandBtn.addEventListener('click', () => {
                const isExpanded = repliesContainer.style.display !== 'none';
                repliesContainer.style.display = isExpanded ? 'none' : 'block';
                svg.style.transform = isExpanded ? 'rotate(-90deg)' : 'rotate(0deg)';
                textSpan.textContent = isExpanded ? `${replyCount} 条回复` : '收起回复';
            });
        }
    },

    /**
     * 渲染子评论卡片
     */
    renderReplyCard(reply) {
        const avatar = reply.avatar || '';
        const author = reply.author || '未知用户';
        const content = reply.content || '';
        const likes = reply.likes || 0;
        const time = reply.time || '';
        const levelInfo = reply.level || 0;

        return `
            <div class="reply-card">
                <div class="reply-line"></div>
                <img src="${avatar}" alt="${author}" class="comment-avatar reply-avatar">
                <div class="reply-content">
                    <div class="reply-header">
                        <strong class="reply-author">${this.escapeHtml(author)}</strong>
                        <span class="comment-user-level">Lv${levelInfo}</span>
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

    /**
     * 应用筛选
     */
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

    /**
     * 导出评论
     */
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

            // 显示下载链接
            const downloadLink = document.getElementById('download-link');
            const downloadUrl = document.getElementById('download-url');
            downloadUrl.href = result.download_url;
            downloadUrl.textContent = result.filename;
            downloadLink.style.display = 'block';

            document.getElementById('export-btn').disabled = false;
            document.getElementById('export-btn').textContent = '导出';

        } catch (error) {
            this.showError('导出失败: ' + error.message);
            document.getElementById('export-btn').disabled = false;
            document.getElementById('export-btn').textContent = '导出';
        }
    },

    /**
     * 显示错误消息
     */
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

    /**
     * 重置UI
     */
    resetUI() {
        const startBtn = document.getElementById('start-btn');
        startBtn.disabled = false;
        startBtn.textContent = '开始爬取';
    },

    /**
     * HTML转义
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});
