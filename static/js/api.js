// V2 API Client - 新版页面专用（更简洁的接口）
const API = {
    baseURL: '/api',

    // 请求封装
    async request(endpoint, options = {}) {
        const url = this.baseURL + endpoint;
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        const response = await fetch(url, { ...defaultOptions, ...options });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: '请求失败' }));
            throw new Error(error.message || error.error || '请求失败');
        }

        return response.json();
    },

    // ========== 评论爬取相关（使用原有API）==========

    async startScrape(data) {
        return this.request('/comments/scrape', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    async getProgress(taskId) {
        return this.request(`/comments/progress/${taskId}`);
    },

    async getResults(taskId, params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/comments/result/${taskId}?${query}`);
    },

    async exportComments(data) {
        return this.request('/comments/export', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    // ========== V2 API - 新版页面专用 ==========

    // 获取所有任务
    // GET /api/v2/tasks -> 直接返回数组
    async getAllTasks() {
        return this.request('/v2/tasks');
    },

    // 获取单个任务详情
    // GET /api/v2/tasks/:id
    async getTaskDetail(taskId) {
        return this.request(`/v2/tasks/${taskId}`);
    },

    // 获取所有模板
    // GET /api/v2/templates -> 直接返回数组
    async getTemplates() {
        return this.request('/v2/templates');
    },

    // 预览Prompt
    // POST /api/v2/preview
    async previewPrompt(data) {
        return this.request('/v2/preview', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    // ========== 流式分析（简化SSE格式）==========

    // SSE格式: data: 文本内容\n\n
    // 完成信号: data: [DONE]\n\n
    // 错误信号: data: [ERROR] 错误信息\n\n
    analyzeStream(request, onChunk, onDone, onError) {
        console.log('[API] Starting stream analysis');

        fetch(this.baseURL + '/v2/analyze-stream', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        })
        .then(response => {
            console.log('[API] Stream response status:', response.status);

            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || err.message || '分析失败');
                });
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';
            let doneCalled = false;

            const read = () => {
                reader.read().then(({ done, value }) => {
                    if (done) {
                        console.log('[API] Stream fully done');
                        if (!doneCalled) {
                            doneCalled = true;
                            onDone();
                        }
                        return;
                    }

                    try {
                        buffer += decoder.decode(value, { stream: true });
                        const lines = buffer.split('\n');
                        buffer = lines.pop() || '';

                        for (const line of lines) {
                            // 跳过空行
                            if (line.trim() === '') continue;

                            // 处理 data: 行
                            if (line.startsWith('data:')) {
                                // 移除 "data: " 前缀
                                let content = line.slice(5).trim();

                                // 检查完成信号
                                if (content === '[DONE]') {
                                    console.log('[API] Received DONE signal');
                                    if (!doneCalled) {
                                        doneCalled = true;
                                        onDone();
                                    }
                                    return;
                                }

                                // 检查错误信号
                                if (content.startsWith('[ERROR]')) {
                                    const errorMsg = content.slice(7).trim();
                                    console.error('[API] Error:', errorMsg);
                                    if (!doneCalled) {
                                        doneCalled = true;
                                        onError(new Error(errorMsg));
                                    }
                                    return;
                                }

                                // 处理内容：反转义换行符
                                const decodedContent = content.replace(/\\n/g, '\n');
                                console.log('[API] Chunk, length:', decodedContent.length);
                                onChunk(decodedContent);
                            }
                        }

                        read();
                    } catch (err) {
                        console.error('[API] Stream error:', err);
                        if (!doneCalled) {
                            doneCalled = true;
                            onError(err);
                        }
                    }
                }).catch(error => {
                    console.error('[API] Read error:', error);
                    if (!doneCalled) {
                        doneCalled = true;
                        onError(error);
                    }
                });
            };

            read();
        })
        .catch(error => {
            console.error('[API] Request error:', error);
            onError(error);
        });
    }
};
