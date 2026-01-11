// 图表渲染模块
const Charts = {
    timeChart: null,
    likesChart: null,

    /**
     * 渲染评论时间分布图
     * @param {Array} comments - 评论数据
     */
    renderTimeChart(comments) {
        const ctx = document.getElementById('time-chart');
        if (!ctx) return;

        // 按日期统计评论数
        const dateCounts = {};
        comments.forEach(comment => {
            const date = comment.time.split(' ')[0]; // 提取日期部分
            dateCounts[date] = (dateCounts[date] || 0) + 1;
        });

        // 排序日期
        const sortedDates = Object.keys(dateCounts).sort();
        const counts = sortedDates.map(date => dateCounts[date]);

        // 销毁旧图表
        if (this.timeChart) {
            this.timeChart.destroy();
        }

        // 创建新图表
        this.timeChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: sortedDates,
                datasets: [{
                    label: '评论数',
                    data: counts,
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            precision: 0
                        }
                    }
                }
            }
        });
    },

    /**
     * 渲染点赞数分布图
     * @param {Array} comments - 评论数据
     */
    renderLikesChart(comments) {
        const ctx = document.getElementById('likes-chart');
        if (!ctx) return;

        // 统计点赞数分布
        const distribution = {
            '0-10': 0,
            '11-50': 0,
            '51-100': 0,
            '100+': 0
        };

        comments.forEach(comment => {
            const likes = comment.likes;
            if (likes <= 10) {
                distribution['0-10']++;
            } else if (likes <= 50) {
                distribution['11-50']++;
            } else if (likes <= 100) {
                distribution['51-100']++;
            } else {
                distribution['100+']++;
            }
        });

        // 销毁旧图表
        if (this.likesChart) {
            this.likesChart.destroy();
        }

        // 创建新图表
        this.likesChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: Object.keys(distribution),
                datasets: [{
                    label: '评论数',
                    data: Object.values(distribution),
                    backgroundColor: [
                        'rgba(102, 126, 234, 0.8)',
                        'rgba(118, 75, 162, 0.8)',
                        'rgba(237, 100, 166, 0.8)',
                        'rgba(255, 154, 158, 0.8)'
                    ],
                    borderColor: [
                        '#667eea',
                        '#764ba2',
                        '#ed64a6',
                        '#ff9a9e'
                    ],
                    borderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return `评论数: ${context.parsed.y}`;
                            }
                        }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            precision: 0
                        }
                    }
                }
            }
        });
    },

    /**
     * 渲染所有图表
     * @param {Array} comments - 评论数据
     */
    renderAll(comments) {
        if (!comments || comments.length === 0) {
            console.warn('No comments data for charts');
            return;
        }

        this.renderTimeChart(comments);
        this.renderLikesChart(comments);
    },

    /**
     * 销毁所有图表
     */
    destroyAll() {
        if (this.timeChart) {
            this.timeChart.destroy();
            this.timeChart = null;
        }
        if (this.likesChart) {
            this.likesChart.destroy();
            this.likesChart = null;
        }
    }
};
