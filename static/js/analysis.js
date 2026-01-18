// AI 分析页面逻辑
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
        // 刷新任务列表
        document.getElementById('refresh-tasks').addEventListener('click', () => this.loadTasks());

        // 分析按钮
        document.getElementById('analyze-btn').addEventListener('click', () => this.startAnalysis());

        // 模态框关闭
        document.querySelector('.close-btn').addEventListener('click', () => this.closeModal());
        document.querySelector('.close-modal-btn').addEventListener('click', () => this.closeModal());

        // 点击模态框外部关闭
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
            container.innerHTML = '<p class="empty">暂无模板</p>';
            return;
        }

        container.innerHTML = this.templates.map(t => `
            <div class="template-item ${this.selectedTemplateId === t.id ? 'selected' : ''}"
                 data-id="${t.id}">
                <div class="template-header">
                    <input type="radio" name="template" value="${t.id}"
                           ${this.selectedTemplateId === t.id ? 'checked' : ''}>
                    <div class="template-info">
                        <h4>${t.name}</h4>
                        <p>${t.description}</p>
                    </div>
                </div>
                ${t.id === 'custom' ? `
                <div class="template-actions">
                    <button class="btn-text edit-prompt-btn">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                        </svg>
                        编辑 Prompt
                    </button>
                </div>` : `
                <div class="template-actions">
                    <button class="btn-text preview-btn">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px;">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                            <circle cx="12" cy="12" r="3"/>
                        </svg>
                        预览 Prompt
                    </button>
                </div>`}
            </div>
        `).join('');

        // 绑定模板选择事件
        container.querySelectorAll('.template-item').forEach(item => {
            item.addEventListener('click', (e) => {
                if (!e.target.classList.contains('preview-btn') &&
                    !e.target.classList.contains('edit-prompt-btn')) {
                    this.selectTemplate(item.dataset.id);
                }
            });
        });

        // 绑定预览按钮事件
        container.querySelectorAll('.preview-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const templateId = btn.closest('.template-item').dataset.id;
                this.previewPrompt(templateId);
            });
        });

        // 绑定编辑按钮事件
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
        document.querySelectorAll('.template-item').forEach(item => {
            item.classList.toggle('selected', item.dataset.id === templateId);
        });

        const customSection = document.getElementById('custom-prompt-section');
        customSection.style.display = templateId === 'custom' ? 'block' : 'none';

        this.updateAnalyzeButton();
    }

    async loadTasks() {
        try {
            const container = document.getElementById('task-list');
            container.innerHTML = '<p class="loading">加载中...</p>';

            this.tasks = await API.getCompletedTasks();
            this.renderTasks();
        } catch (error) {
            this.showError('加载任务失败: ' + error.message);
        }
    }

    renderTasks() {
        const container = document.getElementById('task-list');
        if (this.tasks.length === 0) {
            container.innerHTML = '<p class="empty">暂无已完成任务，请先在主页进行爬取</p>';
            return;
        }

        container.innerHTML = this.tasks.map(t => `
            <div class="task-item ${this.selectedTaskId === t.task_id ? 'selected' : ''}"
                 data-id="${t.task_id}">
                <div class="task-header">
                    <input type="radio" name="task" value="${t.task_id}"
                           ${this.selectedTaskId === t.task_id ? 'checked' : ''}>
                    <div class="task-info">
                        <h4>${this.escapeHtml(t.video_title || '未知标题')}</h4>
                        <p>评论数: ${t.comment_count} | 时间: ${t.start_time}</p>
                    </div>
                </div>
                <div class="task-id">${t.task_id}</div>
            </div>
        `).join('');

        // 绑定任务选择事件
        container.querySelectorAll('.task-item').forEach(item => {
            item.addEventListener('click', () => {
                this.selectTask(item.dataset.id);
            });
        });

        this.updateAnalyzeButton();
    }

    selectTask(taskId) {
        this.selectedTaskId = taskId;
        document.querySelectorAll('.task-item').forEach(item => {
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
        btn.textContent = '分析中...';

        const resultContainer = document.getElementById('result-container');

        // 初始化流式显示容器
        resultContainer.innerHTML = `
            <div class="result-header">
                <h3>分析进行中</h3>
                <span class="timestamp">正在生成分析结果...</span>
            </div>
            <div class="result-content markdown-body streaming-content" id="streaming-content"></div>
            <div class="streaming-indicator">
                <span class="dot"></span>
                <span class="dot"></span>
                <span class="dot"></span>
            </div>
        `;

        const streamingContent = document.getElementById('streaming-content');
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

            // 使用流式 API
            API.analyzeStream(
                request,
                // onChunk - 接收累积的完整内容
                (accumulatedContent) => {
                    fullContent = accumulatedContent; // 直接使用累积的完整内容
                    // 流式渲染期间显示纯文本（保留换行），避免解析不完整的 markdown
                    streamingContent.innerHTML = `<pre style="white-space: pre-wrap; font-family: inherit; line-height: 1.8;">${this.escapeHtml(fullContent)}</pre>`;
                    // 自动滚动到底部
                    streamingContent.scrollTop = streamingContent.scrollHeight;
                },
                // onDone - 分析完成
                () => {
                    console.log('[Analysis] Stream completed, total content length:', fullContent.length);
                    console.log('[Analysis] First 200 chars:', fullContent.substring(0, 200));
                    const timestamp = new Date().toLocaleString();
                    let finalHtml;
                    try {
                        // 检查 mdParser 是否可用
                        if (window.mdParser && typeof window.mdParser.render === 'function') {
                            finalHtml = window.mdParser.render(fullContent);
                            console.log('[Analysis] Rendered HTML length:', finalHtml.length);
                        } else {
                            // 不可用，显示纯文本（保留换行）
                            finalHtml = `<pre style="white-space: pre-wrap; font-family: inherit;">${this.escapeHtml(fullContent)}</pre>`;
                        }
                    } catch (e) {
                        console.error('Markdown parse error:', e);
                        finalHtml = `<pre style="white-space: pre-wrap; font-family: inherit;">${this.escapeHtml(fullContent)}</pre>`;
                    }

                    resultContainer.innerHTML = `
                        <div class="result-header">
                            <h3>分析完成</h3>
                            <span class="timestamp">生成时间: ${this.escapeHtml(timestamp)}</span>
                        </div>
                        <div class="result-content markdown-body">${finalHtml}</div>
                        <div class="result-actions">
                            <button class="btn copy-btn">复制结果</button>
                            <button class="btn download-btn">下载Markdown</button>
                        </div>
                    `;

                    // 绑定复制按钮
                    const copyBtn = resultContainer.querySelector('.copy-btn');
                    if (copyBtn) {
                        copyBtn.addEventListener('click', () => {
                            navigator.clipboard.writeText(fullContent);
                            this.showMessage('已复制到剪贴板');
                        });
                    }

                    // 绑定下载按钮
                    const downloadBtn = resultContainer.querySelector('.download-btn');
                    if (downloadBtn) {
                        downloadBtn.addEventListener('click', () => {
                            this.downloadMarkdown(fullContent);
                        });
                    }

                    btn.disabled = false;
                    btn.textContent = '开始分析';
                    this.updateAnalyzeButton();
                },
                // onError - 发生错误
                (error) => {
                    const errorMsg = error.message || '未知错误';
                    resultContainer.innerHTML = `
                        <div class="error-result">
                            <h3>分析失败</h3>
                            <p>${this.escapeHtml(errorMsg)}</p>
                        </div>
                    `;
                    btn.disabled = false;
                    btn.textContent = '开始分析';
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
            btn.textContent = '开始分析';
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
        // 手动转义 HTML 特殊字符，保留换行符
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
}

// 初始化页面
document.addEventListener('DOMContentLoaded', () => {
    new AnalysisPage();
});
