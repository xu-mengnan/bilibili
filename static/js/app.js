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
                sort_mode: sortMode
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
        document.getElementById('comment-count').textContent = results.total_count;

        // 渲染评论表格
        this.renderCommentsTable(results.comments);

        // 渲染图表
        Charts.renderAll(results.comments);
    },

    /**
     * 渲染评论表格
     */
    renderCommentsTable(comments) {
        const tbody = document.getElementById('comments-tbody');
        tbody.innerHTML = '';

        if (comments.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" style="text-align: center;">暂无评论</td></tr>';
            return;
        }

        comments.forEach(comment => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>
                    <div class="comment-author">
                        <img src="${comment.avatar}" alt="${comment.author}" class="comment-avatar">
                        <div>
                            <div><strong>${this.escapeHtml(comment.author)}</strong></div>
                            <div style="font-size: 12px; color: #999;">Lv${comment.level}</div>
                        </div>
                    </div>
                </td>
                <td class="comment-content">${this.escapeHtml(comment.content)}</td>
                <td>${comment.likes}</td>
                <td>${comment.time}</td>
            `;
            tbody.appendChild(row);
        });
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
