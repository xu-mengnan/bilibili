// API封装模块
const API = {
    // Debug: ensure API is defined
    _version: '2.0',
    /**
     * 启动爬取任务
     * @param {Object} data - 爬取配置
     * @returns {Promise<Object>} 任务信息
     */
    async startScrape(data) {
        const response = await fetch('/api/comments/scrape', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '启动爬取失败');
        }

        return await response.json();
    },

    /**
     * 获取任务进度
     * @param {string} taskId - 任务ID
     * @returns {Promise<Object>} 进度信息
     */
    async getProgress(taskId) {
        const response = await fetch(`/api/comments/progress/${taskId}`);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取进度失败');
        }

        return await response.json();
    },

    /**
     * 获取爬取结果
     * @param {string} taskId - 任务ID
     * @param {Object} params - 查询参数（sort, keyword, limit）
     * @returns {Promise<Object>} 评论结果
     */
    async getResults(taskId, params = {}) {
        const queryParams = new URLSearchParams();
        if (params.sort) queryParams.append('sort', params.sort);
        if (params.keyword) queryParams.append('keyword', params.keyword);
        if (params.limit) queryParams.append('limit', params.limit);

        const url = `/api/comments/result/${taskId}${queryParams.toString() ? '?' + queryParams.toString() : ''}`;
        const response = await fetch(url);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取结果失败');
        }

        return await response.json();
    },

    /**
     * 获取评论统计
     * @param {string} taskId - 任务ID
     * @returns {Promise<Object>} 统计数据
     */
    async getStats(taskId) {
        const response = await fetch(`/api/comments/stats/${taskId}`);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取统计失败');
        }

        return await response.json();
    },

    /**
     * 导出评论
     * @param {Object} data - 导出配置（task_id, format, sort, filename）
     * @returns {Promise<Object>} 文件信息
     */
    async exportComments(data) {
        const response = await fetch('/api/comments/export', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '导出失败');
        }

        return await response.json();
    },

    /**
     * 获取视频信息
     * @param {string} videoInput - 视频ID或URL
     * @returns {Promise<Object>} 视频信息
     */
    async getVideoInfo(videoInput) {
        const response = await fetch('/api/videos/info', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                video_url_or_id: videoInput
            })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取视频信息失败');
        }

        return await response.json();
    },

    // ========== AI分析相关 ==========

    /**
     * 获取预设Prompt模板
     * @returns {Promise<Array>} 模板列表
     */
    async getTemplates() {
        const response = await fetch('/api/analysis/templates');

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取模板失败');
        }

        const data = await response.json();
        return data.templates;
    },

    /**
     * 获取所有已完成的任务
     * @returns {Promise<Array>} 任务列表
     */
    async getCompletedTasks() {
        const response = await fetch('/api/analysis/tasks/completed');

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取任务列表失败');
        }

        const data = await response.json();
        return data.tasks;
    },

    /**
     * 获取任务评论摘要
     * @param {string} taskId - 任务ID
     * @returns {Promise<Object>} 任务评论摘要
     */
    async getTaskSummary(taskId) {
        const response = await fetch(`/api/analysis/tasks/${taskId}`);

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '获取任务信息失败');
        }

        return await response.json();
    },

    /**
     * 执行评论分析
     * @param {Object} data - 分析配置
     * @returns {Promise<Object>} 分析结果
     */
    async analyze(data) {
        const response = await fetch('/api/analysis/analyze', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '分析失败');
        }

        return await response.json();
    },

    /**
     * 预览Prompt渲染结果
     * @param {Object} data - 预览配置（task_id, template_id）
     * @returns {Promise<Object>} 渲染后的Prompt
     */
    async previewPrompt(data) {
        const response = await fetch('/api/analysis/preview', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '预览失败');
        }

        return await response.json();
    },

    /**
     * 执行流式评论分析（使用 Fetch + ReadableStream）
     * @param {Object} data - 分析配置
     * @param {Function} onChunk - 接收数据块的回调函数 (chunk: string) => void
     * @param {Function} onDone - 完成时的回调函数
     * @param {Function} onError - 错误时的回调函数
     * @returns {Object} 包含 close 方法的对象，用于关闭连接
     */
    analyzeStream(data, onChunk, onDone, onError) {
        // 发送 POST 请求启动分析
        console.log('[API] Starting analysis stream with data:', data);
        fetch('/api/analysis/analyze-stream', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            console.log('[API] Response status:', response.status, response.headers.get('content-type'));
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || '启动分析失败');
                });
            }

            // 使用 ReadableStream 读取响应
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';
            let currentEvent = '';
            let doneCalled = false;
            let chunkCount = 0;

            const readStream = () => {
                reader.read().then(({ done, value }) => {
                    if (done) {
                        console.log('[API] Stream done, total chunks:', chunkCount);
                        if (!doneCalled) {
                            doneCalled = true;
                            onDone && onDone();
                        }
                        return;
                    }

                    try {
                        // 解码数据块
                        buffer += decoder.decode(value, { stream: true });

                        // 处理 SSE 格式数据
                        // SSE 格式: event: xxx\ndata: yyy\n\n
                        const lines = buffer.split('\n');
                        buffer = lines.pop() || ''; // 保留未完成的行

                        for (const line of lines) {
                            // 兼容有无空格的格式: "event: content" 或 "event:content"
                            if (line.startsWith('event:')) {
                                const eventValue = line.includes('event: ') ? line.slice(7).trim() : line.slice(6).trim();
                                currentEvent = eventValue;
                            } else if (line.startsWith('data:')) {
                                // 兼容有无空格的格式: "data: xxx" 或 "data:xxx"
                                const content = line.includes('data: ') ? line.slice(6).trim() : line.slice(5).trim();
                                if (content === '[DONE]') {
                                    console.log('[API] Received [DONE] signal');
                                    if (!doneCalled) {
                                        doneCalled = true;
                                        onDone && onDone();
                                    }
                                    return;
                                }
                                if (currentEvent === 'content' && content.length > 0) {
                                    chunkCount++;
                                    // 解析 JSON 编码的内容
                                    let decodedContent;
                                    try {
                                        decodedContent = JSON.parse(content);
                                    } catch (e) {
                                        // 如果不是 JSON 格式，直接使用原内容
                                        decodedContent = content;
                                    }
                                    console.log('[API] Chunk', chunkCount, 'length:', decodedContent.length, 'content:', decodedContent.substring(0, 50) + '...');
                                    onChunk && onChunk(decodedContent);
                                } else if (currentEvent === 'error' && content.length > 0) {
                                    console.error('[API] Error from server:', content);
                                    if (!doneCalled) {
                                        doneCalled = true;
                                        onError && onError(new Error(content));
                                    }
                                    return;
                                }
                            } else if (line === '') {
                                // 空行表示事件结束，重置当前事件
                                currentEvent = '';
                            }
                        }

                        // 继续读取
                        readStream();
                    } catch (err) {
                        console.error('[API] Error in readStream:', err);
                        if (!doneCalled) {
                            doneCalled = true;
                            onError && onError(err);
                        }
                    }
                }).catch(err => {
                    console.error('[API] Error reading stream:', err);
                    if (!doneCalled) {
                        doneCalled = true;
                        onError && onError(err);
                    }
                });
            };

            readStream();
        }).catch(err => {
            console.error('[API] Error starting analysis:', err);
            onError && onError(err);
        });

        // 返回控制对象
        return {
            close: () => {
                // 用于清理资源
            }
        };
    }
};
