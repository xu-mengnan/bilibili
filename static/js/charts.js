// Charts Module
const Charts = {
    timeChart: null,
    likesChart: null,

    renderAll(comments) {
        this.renderTimeChart(comments);
        this.renderLikesChart(comments);
    },

    renderTimeChart(comments) {
        const ctx = document.getElementById('time-chart');
        if (!ctx) return;

        // Destroy existing chart
        if (this.timeChart) {
            this.timeChart.destroy();
        }

        // Process data - group by date
        const timeData = {};
        comments.forEach(comment => {
            let date = 'Unknown';
            if (comment.timestamp) {
                date = new Date(comment.timestamp).toLocaleDateString('zh-CN');
            } else if (comment.time) {
                // Try to parse time string
                const match = comment.time.match(/(\d{4}-\d{2}-\d{2})/);
                if (match) {
                    date = match[1];
                }
            }
            timeData[date] = (timeData[date] || 0) + 1;
        });

        const labels = Object.keys(timeData).sort();
        const data = labels.map(label => timeData[label]);

        this.timeChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: '评论数',
                    data: data,
                    borderColor: '#06B6D4',
                    backgroundColor: 'rgba(6, 182, 212, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4,
                    pointBackgroundColor: '#06B6D4',
                    pointBorderColor: '#fff',
                    pointBorderWidth: 2,
                    pointRadius: 4,
                    pointHoverRadius: 6
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
                        backgroundColor: 'rgba(15, 23, 42, 0.9)',
                        titleColor: '#fff',
                        bodyColor: '#E2E8F0',
                        borderColor: 'rgba(6, 182, 212, 0.3)',
                        borderWidth: 1,
                        cornerRadius: 8,
                        padding: 12
                    }
                },
                scales: {
                    x: {
                        grid: {
                            display: false
                        },
                        ticks: {
                            color: '#64748B',
                            font: {
                                size: 11
                            }
                        }
                    },
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(226, 232, 240, 0.6)'
                        },
                        ticks: {
                            color: '#64748B',
                            font: {
                                size: 11
                            },
                            stepSize: 1
                        }
                    }
                }
            }
        });
    },

    renderLikesChart(comments) {
        const ctx = document.getElementById('likes-chart');
        if (!ctx) return;

        // Destroy existing chart
        if (this.likesChart) {
            this.likesChart.destroy();
        }

        // Process data - group likes into ranges
        const ranges = {
            '0': 0,
            '1-10': 0,
            '11-50': 0,
            '51-100': 0,
            '101+': 0
        };

        comments.forEach(comment => {
            const likes = comment.likes || 0;
            if (likes === 0) {
                ranges['0']++;
            } else if (likes <= 10) {
                ranges['1-10']++;
            } else if (likes <= 50) {
                ranges['11-50']++;
            } else if (likes <= 100) {
                ranges['51-100']++;
            } else {
                ranges['101+']++;
            }
        });

        const labels = Object.keys(ranges);
        const data = Object.values(ranges);

        // Generate gradient colors
        const gradientColors = labels.map((_, i) => {
            const alpha = 0.4 + (i * 0.15);
            return `rgba(6, 182, 212, ${alpha})`;
        });

        this.likesChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: '评论数',
                    data: data,
                    backgroundColor: gradientColors,
                    borderColor: '#06B6D4',
                    borderWidth: 0,
                    borderRadius: 6,
                    borderSkipped: false
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
                        backgroundColor: 'rgba(15, 23, 42, 0.9)',
                        titleColor: '#fff',
                        bodyColor: '#E2E8F0',
                        borderColor: 'rgba(6, 182, 212, 0.3)',
                        borderWidth: 1,
                        cornerRadius: 8,
                        padding: 12
                    }
                },
                scales: {
                    x: {
                        grid: {
                            display: false
                        },
                        ticks: {
                            color: '#64748B',
                            font: {
                                size: 11
                            }
                        }
                    },
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(226, 232, 240, 0.6)'
                        },
                        ticks: {
                            color: '#64748B',
                            font: {
                                size: 11
                            },
                            stepSize: 1
                        }
                    }
                }
            }
        });
    },

    destroy() {
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
