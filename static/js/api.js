// API封装模块
const API = {
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
    }
};
