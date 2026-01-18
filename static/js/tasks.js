// 任务管理页面逻辑
class TasksPage {
    constructor() {
        this.tasks = [];
        this.filter = '';
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadTasks();
    }

    bindEvents() {
        // 刷新按钮
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadTasks();
        });

        // 状态筛选
        document.getElementById('status-filter').addEventListener('change', (e) => {
            this.filter = e.target.value;
            this.renderTasks();
        });

        // 模态框关闭
        document.querySelector('.close-btn').addEventListener('click', () => {
            this.closeModal();
        });

        // 点击模态框外部关闭
        document.getElementById('task-detail-modal').addEventListener('click', (e) => {
            if (e.target.id === 'task-detail-modal') {
                this.closeModal();
            }
        });
    }

    async loadTasks() {
        const container = document.getElementById('tasks-list');
        container.innerHTML = `
            <div class="loading-spinner">
                <div class="spinner"></div>
            </div>
        `;

        try {
            this.tasks = await API.getAllTasks();
            this.renderTasks();
            this.updateStats();
        } catch (error) {
            console.error('加载任务失败:', error);
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">
                        <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <circle cx="12" cy="12" r="10"/>
                            <line x1="12" y1="8" x2="12" y2="12"/>
                            <line x1="12" y1="16" x2="12.01" y2="16"/>
                        </svg>
                    </div>
                    <div class="empty-state-title">加载失败</div>
                    <div class="empty-state-description">${this.escapeHtml(error.message)}</div>
                </div>
            `;
        }
    }

    renderTasks() {
        const container = document.getElementById('tasks-list');

        // 筛选任务
        let filteredTasks = this.tasks;
        if (this.filter) {
            filteredTasks = this.tasks.filter(t => t.status === this.filter);
        }

        if (filteredTasks.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">
                        <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <rect x="3" y="3" width="7" height="7"/>
                            <rect x="14" y="3" width="7" height="7"/>
                            <rect x="14" y="14" width="7" height="7"/>
                            <rect x="3" y="14" width="7" height="7"/>
                        </svg>
                    </div>
                    <div class="empty-state-title">暂无任务</div>
                    <div class="empty-state-description">${this.filter ? '没有符合筛选条件的任务' : '还未创建任何爬取任务'}</div>
                </div>
            `;
            return;
        }

        container.innerHTML = filteredTasks.map(task => this.renderTaskItem(task)).join('');

        // 绑定点击事件
        container.querySelectorAll('.task-item').forEach(item => {
            item.addEventListener('click', () => {
                const taskId = item.dataset.taskId;
                this.showTaskDetail(taskId);
            });
        });
    }

    renderTaskItem(task) {
        const statusText = {
            'running': '运行中',
            'completed': '已完成',
            'failed': '失败'
        };

        return `
            <div class="task-item" data-task-id="${task.task_id}">
                <div class="task-item-header">
                    <div class="task-item-title">
                        <h3>${this.escapeHtml(task.video_title || '未知标题')}</h3>
                    </div>
                    <span class="status-badge ${task.status}">
                        <span class="status-dot"></span>
                        ${statusText[task.status] || task.status}
                    </span>
                </div>
                <div class="task-item-meta">
                    <span>任务ID: ${task.task_id.substring(0, 8)}...</span>
                    <span>视频ID: ${task.video_id || '-'}</span>
                    <span>开始时间: ${task.start_time}</span>
                </div>
                <div class="task-info-row">
                    <div class="task-info-item">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                        </svg>
                        <span>${task.comment_count || 0} 条评论</span>
                    </div>
                    ${task.error ? `
                    <div class="task-info-item" style="color: var(--error-color);">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"/>
                            <line x1="12" y1="8" x2="12" y2="12"/>
                            <line x1="12" y1="16" x2="12.01" y2="16"/>
                        </svg>
                        <span>${this.escapeHtml(task.error)}</span>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
    }

    updateStats() {
        const total = this.tasks.length;
        const running = this.tasks.filter(t => t.status === 'running').length;
        const completed = this.tasks.filter(t => t.status === 'completed').length;
        const failed = this.tasks.filter(t => t.status === 'failed').length;

        document.getElementById('stat-total').textContent = total;
        document.getElementById('stat-running').textContent = running;
        document.getElementById('stat-completed').textContent = completed;
        document.getElementById('stat-failed').textContent = failed;
    }

    async showTaskDetail(taskId) {
        const content = document.getElementById('task-detail-content');
        content.innerHTML = `
            <div class="loading-spinner">
                <div class="spinner"></div>
            </div>
        `;

        document.getElementById('task-detail-modal').style.display = 'flex';

        try {
            const task = await API.getTaskProgress(taskId);

            const statusText = {
                'running': '运行中',
                'completed': '已完成',
                'failed': '失败'
            };

            // 调试：打印任务数据
            console.log('Task data:', task);
            console.log('Comments:', task.comments);
            console.log('Comments length:', task.comments?.length);

            content.innerHTML = `
                <div class="task-detail-section">
                    <h4>基本信息</h4>
                    <div class="task-detail-grid">
                        <div class="task-detail-item">
                            <span class="task-detail-label">任务状态</span>
                            <span class="task-detail-value">
                                <span class="status-badge ${task.status}">
                                    <span class="status-dot"></span>
                                    ${statusText[task.status] || task.status}
                                </span>
                            </span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">视频标题</span>
                            <span class="task-detail-value">${this.escapeHtml(task.video_title || '-')}</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">视频ID</span>
                            <span class="task-detail-value">${task.video_id || '-'}</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">任务ID</span>
                            <span class="task-detail-value" style="font-family: monospace; font-size: var(--font-size-xs);">${task.task_id}</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">开始时间</span>
                            <span class="task-detail-value">${task.start_time}</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">结束时间</span>
                            <span class="task-detail-value">${task.end_time || '-'}</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">评论数量</span>
                            <span class="task-detail-value">${task.progress?.total_comments || 0} 条</span>
                        </div>
                        <div class="task-detail-item">
                            <span class="task-detail-label">爬取页数</span>
                            <span class="task-detail-value">${task.progress?.current_page || 0} / ${task.progress?.page_limit || 0}</span>
                        </div>
                    </div>
                    ${task.error ? `
                    <div style="margin-top: var(--space-md); padding: var(--space-md); background: var(--error-light); border-radius: var(--radius-md); color: var(--error-color);">
                        <strong>错误信息：</strong>${this.escapeHtml(task.error)}
                    </div>
                    ` : ''}
                </div>

                ${task.status === 'completed' && task.comments && task.comments.length > 0 ? `
                <div class="task-detail-section">
                    <h4>评论预览 (前 5 条，共 ${task.comments.length} 条)</h4>
                    <div class="comments-preview">
                        ${task.comments.slice(0, 5).map(comment => `
                            <div class="comment-preview-item">
                                <div class="comment-preview-header">
                                    <span class="comment-preview-author">${this.escapeHtml(comment.member?.uname || '匿名用户')}</span>
                                    <span class="comment-preview-time">${new Date(comment.ctime * 1000).toLocaleString()}</span>
                                </div>
                                <div class="comment-preview-content">${this.escapeHtml(comment.content?.message || '')}</div>
                            </div>
                        `).join('')}
                    </div>
                </div>
                ` : task.status === 'completed' ? `
                <div class="task-detail-section">
                    <h4>评论数据</h4>
                    <div style="padding: var(--space-md); background: var(--warning-light); border-radius: var(--radius-md); color: var(--warning-color);">
                        评论数据正在加载中，请稍后刷新重试。
                    </div>
                </div>
                ` : ''}
            `;

        } catch (error) {
            console.error('加载任务详情失败:', error);
            content.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-title">加载失败</div>
                    <div class="empty-state-description">${this.escapeHtml(error.message)}</div>
                </div>
            `;
        }
    }

    closeModal() {
        document.getElementById('task-detail-modal').style.display = 'none';
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// 初始化页面
document.addEventListener('DOMContentLoaded', () => {
    new TasksPage();
});
