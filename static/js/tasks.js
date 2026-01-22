// Tasks Page Logic
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
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadTasks();
        });

        document.getElementById('status-filter').addEventListener('change', (e) => {
            this.filter = e.target.value;
            this.renderTasks();
        });

        document.querySelector('.modal-close').addEventListener('click', () => {
            this.closeModal();
        });

        document.getElementById('task-detail-modal').addEventListener('click', (e) => {
            if (e.target.id === 'task-detail-modal') {
                this.closeModal();
            }
        });
    }

    async loadTasks() {
        const container = document.getElementById('tasks-list');
        container.innerHTML = `
            <div class="loading-state">
                <div class="loading-spinner"></div>
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
                    <div class="task-title">
                        <h3>${this.escapeHtml(task.video_title || '未知标题')}</h3>
                    </div>
                    <div class="task-badges">
                        <span class="status-badge ${task.status}">
                            <span class="status-dot"></span>
                            ${statusText[task.status] || task.status}
                        </span>
                    </div>
                </div>
                <div class="task-meta">
                    <div class="task-meta-item">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5.586a1 1 0 0 1 .707.293l5.414 5.414a1 1 0 0 1 .293.707V19a2 2 0 0 1-2 2z"/>
                        </svg>
                        <span>ID: ${task.task_id.substring(0, 8)}...</span>
                    </div>
                    <div class="task-meta-item">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
                            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
                        </svg>
                        <span>${task.video_id || '-'}</span>
                    </div>
                    <div class="task-meta-item">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"/>
                            <polyline points="12 6 12 12 16 14"/>
                        </svg>
                        <span>${task.start_time}</span>
                    </div>
                </div>
                <div class="task-info-row">
                    <div class="task-meta-item">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                        </svg>
                        <span>${task.comment_count || 0} 条评论</span>
                    </div>
                    ${task.error ? `
                    <div class="task-meta-item" style="color: var(--error);">
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
            <div class="loading-state">
                <div class="loading-spinner"></div>
            </div>
        `;

        document.getElementById('task-detail-modal').style.display = 'flex';

        try {
            const task = await API.getTaskDetail(taskId);

            const statusText = {
                'running': '运行中',
                'completed': '已完成',
                'failed': '失败'
            };

            const comments = task.comments || [];
            const progress = task.progress || {};

            content.innerHTML = `
                <div class="detail-section">
                    <h3 class="detail-section-title">基本信息</h3>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <div class="detail-label">任务状态</div>
                            <div class="detail-value">
                                <span class="status-badge ${task.status}">
                                    <span class="status-dot"></span>
                                    ${statusText[task.status] || task.status}
                                </span>
                            </div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">视频标题</div>
                            <div class="detail-value">${this.escapeHtml(task.video_title || '-')}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">视频ID</div>
                            <div class="detail-value">${task.video_id || '-'}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">任务ID</div>
                            <div class="detail-value" style="font-family: monospace; font-size: 12px;">${task.task_id}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">开始时间</div>
                            <div class="detail-value">${task.start_time}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">结束时间</div>
                            <div class="detail-value">${task.end_time || '-'}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">评论数量</div>
                            <div class="detail-value">${progress.total_comments || task.comment_count || 0} 条</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">爬取页数</div>
                            <div class="detail-value">${progress.current_page || 0} / ${progress.page_limit || 0}</div>
                        </div>
                    </div>
                    ${task.error ? `
                    <div class="error-box" style="margin-top: var(--space-md);">
                        <strong>错误信息：</strong>${this.escapeHtml(task.error)}
                    </div>
                    ` : ''}
                </div>

                ${task.status === 'completed' && comments.length > 0 ? `
                <div class="detail-section">
                    <h3 class="detail-section-title">评论预览 (前 5 条)</h3>
                    <div class="comments-preview">
                        ${comments.slice(0, 5).map(comment => `
                            <div class="comment-preview-item">
                                <div class="comment-preview-header">
                                    <span class="comment-preview-author">${this.escapeHtml(comment.author || '匿名用户')}</span>
                                    <span class="comment-preview-time">${comment.time}</span>
                                </div>
                                <div class="comment-preview-content">${this.escapeHtml(comment.content || '')}</div>
                            </div>
                        `).join('')}
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

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    new TasksPage();
});
