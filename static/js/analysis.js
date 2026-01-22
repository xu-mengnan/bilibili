// AI Analysis Page Logic
class AnalysisPage {
    constructor() {
        this.selectedTaskId = null;
        this.selectedTemplateId = null;
        this.templates = [];
        this.tasks = [];
        this.init();
    }

    async init() {
        this.bindEvents();
        await this.loadTemplates();
        await this.loadTasks();
    }

    bindEvents() {
        document.getElementById('refresh-tasks').addEventListener('click', (e) => {
            const btn = e.currentTarget;
            btn.classList.add('spinning');
            this.loadTasks().finally(() => {
                setTimeout(() => btn.classList.remove('spinning'), 500);
            });
        });

        document.getElementById('analyze-btn').addEventListener('click', () => this.startAnalysis());

        document.querySelector('.modal-close').addEventListener('click', () => this.closeModal());
        document.querySelector('.close-modal-btn').addEventListener('click', () => this.closeModal());

        document.getElementById('preview-modal').addEventListener('click', (e) => {
            if (e.target.id === 'preview-modal') {
                this.closeModal();
            }
        });
    }

    async loadTemplates() {
        try {
            this.templates = await API.getTemplates();
            this.renderTemplates();
        } catch (error) {
            this.showError('加载模板失败: ' + error.message);
        }
    }

    renderTemplates() {
        const container = document.getElementById('template-list');
        if (this.templates.length === 0) {
            container.innerHTML = '<p style="color: var(--gray-400); text-align: center; padding: var(--space-lg);">暂无模板</p>';
            return;
        }

        container.innerHTML = this.templates.map(t => `
            <div class="template-option ${this.selectedTemplateId === t.id ? 'selected' : ''}" data-id="${t.id}">
                <div class="template-header">
                    <input type="radio" name="template" value="${t.id}" ${this.selectedTemplateId === t.id ? 'checked' : ''}>
                    <div class="template-info">
                        <div class="template-name">${t.name}</div>
                        <div class="template-description">${t.description}</div>
                    </div>
                </div>
                ${t.id === 'custom' ? `
                <div class="template-actions">
                    <button class="template-action-btn edit-prompt-btn">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                        </svg>
                        编辑
                    </button>
                </div>` : `
                <div class="template-actions">
                    <button class="template-action-btn preview-btn">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                            <circle cx="12" cy="12" r="3"/>
                        </svg>
                        预览
                    </button>
                </div>`}
            </div>
        `).join('');

        // Bind selection events
        container.querySelectorAll('.template-option').forEach(item => {
            item.addEventListener('click', (e) => {
                if (!e.target.classList.contains('preview-btn') && !e.target.classList.contains('edit-prompt-btn')) {
                    this.selectTemplate(item.dataset.id);
                }
            });
        });

        // Bind preview buttons
        container.querySelectorAll('.preview-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const templateId = btn.closest('.template-option').dataset.id;
                this.previewPrompt(templateId);
            });
        });

        // Bind edit buttons
        container.querySelectorAll('.edit-prompt-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.selectTemplate('custom');
            });
        });

        this.updateAnalyzeButton();
    }

    selectTemplate(templateId) {
        this.selectedTemplateId = templateId;
        document.querySelectorAll('.template-option').forEach(item => {
            item.classList.toggle('selected', item.dataset.id === templateId);
        });

        const customSection = document.getElementById('custom-prompt-section');
        customSection.classList.toggle('show', templateId === 'custom');

        this.updateAnalyzeButton();
    }

    async loadTasks() {
        try {
            const container = document.getElementById('task-list');
            container.innerHTML = '<p style="color: var(--gray-400); text-align: center; padding: var(--space-lg);">加载中...</p>';

            // 获取所有任务，过滤已完成的
            const allTasks = await API.getAllTasks();
            this.tasks = allTasks.filter(t => t.status === 'completed');
            this.renderTasks();
        } catch (error) {
            this.showError('加载任务失败: ' + error.message);
        }
    }

    renderTasks() {
        const container = document.getElementById('task-list');
        if (this.tasks.length === 0) {
            container.innerHTML = '<p style="color: var(--gray-400); text-align: center; padding: var(--space-lg);">暂无已完成任务</p>';
            return;
        }

        container.innerHTML = this.tasks.map(t => `
            <div class="task-option ${this.selectedTaskId === t.task_id ? 'selected' : ''}" data-id="${t.task_id}">
                <input type="radio" name="task" value="${t.task_id}" ${this.selectedTaskId === t.task_id ? 'checked' : ''}>
                <div class="task-info">
                    <div class="task-title">${this.escapeHtml(t.video_title || '未知标题')}</div>
                    <div class="task-meta">评论数: ${t.comment_count} | 时间: ${t.start_time}</div>
                </div>
            </div>
        `).join('');

        container.querySelectorAll('.task-option').forEach(item => {
            item.addEventListener('click', () => {
                this.selectTask(item.dataset.id);
            });
        });

        this.updateAnalyzeButton();
    }

    selectTask(taskId) {
        this.selectedTaskId = taskId;
        document.querySelectorAll('.task-option').forEach(item => {
            item.classList.toggle('selected', item.dataset.id === taskId);
        });

        this.updateAnalyzeButton();
    }

    updateAnalyzeButton() {
        const btn = document.getElementById('analyze-btn');
        btn.disabled = !(this.selectedTaskId && this.selectedTemplateId);
    }

    async previewPrompt(templateId) {
        if (!this.selectedTaskId) {
            this.showError('请先选择任务');
            return;
        }

        try {
            const result = await API.previewPrompt({
                task_id: this.selectedTaskId,
                template_id: templateId
            });

            document.getElementById('preview-content').textContent = result.prompt;
            document.getElementById('preview-modal').style.display = 'flex';
        } catch (error) {
            this.showError('预览失败: ' + error.message);
        }
    }

    closeModal() {
        document.getElementById('preview-modal').style.display = 'none';
    }

    async startAnalysis() {
        if (!this.selectedTaskId || !this.selectedTemplateId) {
            this.showError('请先选择任务和模板');
            return;
        }

        const btn = document.getElementById('analyze-btn');
        btn.disabled = true;
        btn.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinner">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            分析中...
        `;

        const resultContainer = document.getElementById('result-container');

        resultContainer.innerHTML = `
            <div class="analyzing-state">
                <div class="analyzing-spinner"></div>
                <div class="analyzing-text">AI 正在分析评论...</div>
                <div class="analyzing-hint">这可能需要几秒钟，请耐心等待</div>
                <div id="streaming-preview" style="margin-top: var(--space-md); padding: var(--space-md); background: var(--gray-50); border-radius: var(--radius-md); max-height: 300px; overflow-y: auto; text-align: left; font-size: 13px; color: var(--gray-600); white-space: pre-wrap; display: none;"></div>
            </div>
        `;

        const streamingPreview = document.getElementById('streaming-preview');
        let fullContent = '';

        try {
            const commentLimit = parseInt(document.getElementById('comment-limit').value) || 0;
            const customPrompt = document.getElementById('custom-prompt').value;

            const request = {
                task_id: this.selectedTaskId,
                template_id: this.selectedTemplateId,
                comment_limit: commentLimit
            };

            if (this.selectedTemplateId === 'custom' && customPrompt) {
                request.custom_prompt = customPrompt;
            }

            API.analyzeStream(
                request,
                (accumulatedContent) => {
                    fullContent = accumulatedContent;
                    // 实时显示进度（纯文本，避免解析不完整的markdown）
                    streamingPreview.style.display = 'block';
                    streamingPreview.textContent = fullContent;
                    // 自动滚动到底部
                    streamingPreview.scrollTop = streamingPreview.scrollHeight;
                },
                () => {
                    const timestamp = new Date().toLocaleString();
                    let finalHtml;

                    try {
                        if (window.mdParser && typeof window.mdParser.render === 'function') {
                            finalHtml = window.mdParser.render(fullContent);
                        } else {
                            finalHtml = `<pre style="white-space: pre-wrap; font-family: inherit;">${this.escapeHtml(fullContent)}</pre>`;
                        }
                    } catch (e) {
                        finalHtml = `<pre style="white-space: pre-wrap; font-family: inherit;">${this.escapeHtml(fullContent)}</pre>`;
                    }

                    resultContainer.innerHTML = `
                        <div class="result-header">
                            <div class="result-title">分析完成</div>
                            <div class="result-timestamp">${this.escapeHtml(timestamp)}</div>
                        </div>
                        <div class="result-content markdown-body">${finalHtml}</div>
                        <div class="result-actions">
                            <button class="action-btn copy">
                                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                                </svg>
                                复制结果
                            </button>
                            <button class="action-btn download">
                                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                    <polyline points="7 10 12 15 17 10"/>
                                    <line x1="12" y1="15" x2="12" y2="3"/>
                                </svg>
                                下载 Markdown
                            </button>
                        </div>
                    `;

                    // Bind copy button
                    const copyBtn = resultContainer.querySelector('.action-btn.copy');
                    if (copyBtn) {
                        copyBtn.addEventListener('click', () => {
                            navigator.clipboard.writeText(fullContent);
                            this.showMessage('已复制到剪贴板');
                        });
                    }

                    // Bind download button
                    const downloadBtn = resultContainer.querySelector('.action-btn.download');
                    if (downloadBtn) {
                        downloadBtn.addEventListener('click', () => {
                            this.downloadMarkdown(fullContent);
                        });
                    }

                    btn.disabled = false;
                    btn.innerHTML = `
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                            <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                            <line x1="12" y1="22.08" x2="12" y2="12"/>
                        </svg>
                        开始分析
                    `;
                    this.updateAnalyzeButton();
                },
                (error) => {
                    const errorMsg = error.message || '未知错误';
                    resultContainer.innerHTML = `
                        <div class="error-result">
                            <h3>分析失败</h3>
                            <p>${this.escapeHtml(errorMsg)}</p>
                        </div>
                    `;
                    btn.disabled = false;
                    btn.innerHTML = `
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                            <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                            <line x1="12" y1="22.08" x2="12" y2="12"/>
                        </svg>
                        开始分析
                    `;
                    this.updateAnalyzeButton();
                }
            );

        } catch (error) {
            const errorMsg = error.message || '未知错误';
            resultContainer.innerHTML = `
                <div class="error-result">
                    <h3>分析失败</h3>
                    <p>${this.escapeHtml(errorMsg)}</p>
                </div>
            `;
            btn.disabled = false;
            btn.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                    <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                    <line x1="12" y1="22.08" x2="12" y2="12"/>
                </svg>
                开始分析
            `;
            this.updateAnalyzeButton();
        }
    }

    downloadMarkdown(content) {
        const blob = new Blob([content], { type: 'text/markdown' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `analysis_${this.selectedTaskId}_${Date.now()}.md`;
        a.click();
        URL.revokeObjectURL(url);
    }

    showError(message) {
        const el = document.getElementById('error-message');
        if (!el) {
            console.error('Error element not found:', message);
            return;
        }
        el.textContent = message;
        el.style.display = 'block';
        setTimeout(() => {
            el.style.display = 'none';
        }, 5000);
    }

    showMessage(message) {
        const el = document.getElementById('error-message');
        if (!el) {
            console.error('Message element not found:', message);
            return;
        }
        el.textContent = message;
        el.className = 'success-message';
        el.style.display = 'block';
        setTimeout(() => {
            el.style.display = 'none';
            el.className = 'error-message';
        }, 3000);
    }

    escapeHtml(text) {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    new AnalysisPage();
});
