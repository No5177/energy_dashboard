// 節能看板 JavaScript 控制器
class EnergyDashboard {
    constructor() {
        this.charts = {};
        this.updateInterval = 5000; // 5秒更新一次
        this.init();
    }

    // 初始化儀表板
    init() {
        this.updateDateTime();
        this.createCharts();
        this.startAutoUpdate();
        this.loadEnergyData();
        
        // 設定自動更新
        setInterval(() => {
            this.updateDateTime();
        }, 1000);

        setInterval(() => {
            this.loadEnergyData();
        }, this.updateInterval);
    }

    // 更新日期時間顯示
    updateDateTime() {
        const now = new Date();
        
        // 格式化日期 (yyyy/mm/dd)
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        const dateStr = `${year}/${month}/${day}`;
        
        // 格式化時間 (HH:MM:SS)
        const hours = String(now.getHours()).padStart(2, '0');
        const minutes = String(now.getMinutes()).padStart(2, '0');
        const seconds = String(now.getSeconds()).padStart(2, '0');
        const timeStr = `${hours}:${minutes}:${seconds}`;
        
        // 更新DOM元素
        const dateElement = document.getElementById('currentDate');
        const timeElement = document.getElementById('currentTime');
        
        if (dateElement) dateElement.textContent = dateStr;
        if (timeElement) timeElement.textContent = timeStr;
    }

    // 載入能源資料
    async loadEnergyData() {
        try {
            // 嘗試從API載入資料
            let response;
            try {
                response = await fetch('/api/latest');
            } catch (error) {
                // 回退到final.json
                response = await fetch('final.json?' + new Date().getTime());
            }
            
            if (response.ok) {
                const data = await response.json();
                this.updateDisplayData(data);
            } else {
                // 使用模擬資料
                this.updateDisplayData(this.generateMockData());
            }
        } catch (error) {
            console.warn('載入資料失敗，使用模擬資料:', error);
            this.updateDisplayData(this.generateMockData());
        }
    }

    // 更新顯示資料
    updateDisplayData(data) {
        // 如果資料是從final.json來的，保持原有格式
        if (Array.isArray(data)) {
            // 更新小卡片資料
            this.updateSmallCards(data);
        } else {
            // 如果是其他格式，使用模擬資料
            this.updateSmallCards(this.generateMockData());
        }
        
        // 更新圖表資料
        this.updateChartData();
    }

    // 更新小卡片資料
    updateSmallCards(data) {
        // 資料映射表
        const dataMapping = {
            'Voltage': { index: 0, suffix: 'V' },
            'Current': { index: 1, suffix: 'A' },
            'Frequency': { index: 2, suffix: 'Hz' },
            'PF': { index: 5, suffix: '' },
            'THD_V': { index: 3, suffix: '%' },
            'THD_A': { index: 4, suffix: '%' },
            'Daily Energy Usage': { index: 6, suffix: 'kW' },
            'Daily Energy': { index: 7, suffix: 'kW' }
        };

        // 更新每個小卡片
        Object.entries(dataMapping).forEach(([label, config]) => {
            const cards = document.querySelectorAll('.small-card');
            cards.forEach(card => {
                const cardLabel = card.querySelector('.card-label');
                if (cardLabel && cardLabel.textContent === label) {
                    const valueElement = card.querySelector('.card-value');
                    if (valueElement) {
                        if (data[config.index]) {
                            const value = parseFloat(data[config.index].value) || 0;
                            valueElement.innerHTML = `${value.toFixed(label === 'PF' ? 3 : 1)} ${config.suffix ? `<span class="unit">${config.suffix}</span>` : ''}`;
                        }
                    }
                }
            });
        });
    }

    // 生成模擬資料
    generateMockData() {
        return [
            { index: 0, name: "Voltage", value: (220 + Math.random() * 10 - 5).toFixed(1), unit: "V" },
            { index: 1, name: "Current", value: (20 + Math.random() * 5 - 2.5).toFixed(2), unit: "A" },
            { index: 2, name: "Frequency", value: (60 + Math.random() * 0.5 - 0.25).toFixed(1), unit: "Hz" },
            { index: 3, name: "THD_V", value: (99 + Math.random() * 2 - 1).toFixed(1), unit: "%" },
            { index: 4, name: "THD_A", value: (1.5 + Math.random() * 1 - 0.5).toFixed(1), unit: "%" },
            { index: 5, name: "PF", value: (0.998 + Math.random() * 0.004 - 0.002).toFixed(3), unit: "" },
            { index: 6, name: "Daily Energy Usage", value: (9999 + Math.random() * 200 - 100).toFixed(0), unit: "kW" },
            { index: 7, name: "Daily Energy", value: (8888 + Math.random() * 200 - 100).toFixed(0), unit: "kW" }
        ];
    }

    // 創建圖表
    createCharts() {
        // 創建小圖表
        this.createDailyUsageChart();
        this.createDailyReclaimedChart();
        
        // 創建大圖表
        this.createEnergyUsageChart();
        this.createEnergyReclaimedChart();
    }

    // 創建每日使用量圖表
    createDailyUsageChart() {
        const canvas = document.getElementById('dailyUsageChart');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const data = this.generateChartData(24); // 24小時資料
        
        this.drawLineChart(ctx, canvas, data, '#f1c40f');
    }

    // 創建每日回收量圖表
    createDailyReclaimedChart() {
        const canvas = document.getElementById('dailyReclaimedChart');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const data = this.generateChartData(24); // 24小時資料
        
        this.drawLineChart(ctx, canvas, data, '#f39c12');
    }

    // 創建能源使用大圖表
    createEnergyUsageChart() {
        const canvas = document.getElementById('energyUsageChart');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const data = this.generateChartData(48); // 48小時資料
        
        this.drawLargeChart(ctx, canvas, data, ['#3498db', '#2980b9', '#1abc9c']);
    }

    // 創建能源回收大圖表
    createEnergyReclaimedChart() {
        const canvas = document.getElementById('energyReclaimedChart');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const data = this.generateChartData(48); // 48小時資料
        
        this.drawLargeChart(ctx, canvas, data, ['#9b59b6', '#8e44ad', '#e74c3c']);
    }

    // 繪製線形圖表
    drawLineChart(ctx, canvas, data, color) {
        const width = canvas.width;
        const height = canvas.height;
        const padding = 20;
        const chartWidth = width - padding * 2;
        const chartHeight = height - padding * 2;

        // 清除畫布
        ctx.clearRect(0, 0, width, height);

        // 找出最大值和最小值
        const values = data.map(d => d.value);
        const maxValue = Math.max(...values);
        const minValue = Math.min(...values);
        const valueRange = maxValue - minValue || 1;

        // 創建漸變
        const gradient = ctx.createLinearGradient(0, padding, 0, height - padding);
        gradient.addColorStop(0, color + '80');
        gradient.addColorStop(1, color + '20');

        // 繪製填充區域
        ctx.beginPath();
        ctx.moveTo(padding, height - padding);
        
        data.forEach((point, index) => {
            const x = padding + (index / (data.length - 1)) * chartWidth;
            const y = height - padding - ((point.value - minValue) / valueRange) * chartHeight;
            
            if (index === 0) {
                ctx.lineTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        
        ctx.lineTo(width - padding, height - padding);
        ctx.closePath();
        ctx.fillStyle = gradient;
        ctx.fill();

        // 繪製線條
        ctx.beginPath();
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        
        data.forEach((point, index) => {
            const x = padding + (index / (data.length - 1)) * chartWidth;
            const y = height - padding - ((point.value - minValue) / valueRange) * chartHeight;
            
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        
        ctx.stroke();
    }

    // 繪製大型圖表
    drawLargeChart(ctx, canvas, data, colors) {
        const width = canvas.width;
        const height = canvas.height;
        const padding = 30;
        const chartWidth = width - padding * 2;
        const chartHeight = height - padding * 2;

        // 清除畫布
        ctx.clearRect(0, 0, width, height);

        // 生成多條數據線
        const datasets = [
            data.map(d => ({ ...d, value: d.value + Math.random() * 20 - 10 })),
            data.map(d => ({ ...d, value: d.value + Math.random() * 15 - 7.5 })),
            data.map(d => ({ ...d, value: d.value + Math.random() * 25 - 12.5 }))
        ];

        // 找出所有數據的最大值和最小值
        const allValues = datasets.flat().map(d => d.value);
        const maxValue = Math.max(...allValues);
        const minValue = Math.min(...allValues);
        const valueRange = maxValue - minValue || 1;

        // 繪製每條數據線
        datasets.forEach((dataset, datasetIndex) => {
            const color = colors[datasetIndex];
            
            // 創建漸變
            const gradient = ctx.createLinearGradient(0, padding, 0, height - padding);
            gradient.addColorStop(0, color + '40');
            gradient.addColorStop(1, color + '10');

            // 繪製填充區域
            if (datasetIndex === 0) {
                ctx.beginPath();
                ctx.moveTo(padding, height - padding);
                
                dataset.forEach((point, index) => {
                    const x = padding + (index / (dataset.length - 1)) * chartWidth;
                    const y = height - padding - ((point.value - minValue) / valueRange) * chartHeight;
                    ctx.lineTo(x, y);
                });
                
                ctx.lineTo(width - padding, height - padding);
                ctx.closePath();
                ctx.fillStyle = gradient;
                ctx.fill();
            }

            // 繪製線條
            ctx.beginPath();
            ctx.strokeStyle = color;
            ctx.lineWidth = 2;
            
            dataset.forEach((point, index) => {
                const x = padding + (index / (dataset.length - 1)) * chartWidth;
                const y = height - padding - ((point.value - minValue) / valueRange) * chartHeight;
                
                if (index === 0) {
                    ctx.moveTo(x, y);
                } else {
                    ctx.lineTo(x, y);
                }
            });
            
            ctx.stroke();
        });

        // 繪製座標軸
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
        ctx.lineWidth = 1;
        
        // Y軸
        ctx.beginPath();
        ctx.moveTo(padding, padding);
        ctx.lineTo(padding, height - padding);
        ctx.stroke();
        
        // X軸
        ctx.beginPath();
        ctx.moveTo(padding, height - padding);
        ctx.lineTo(width - padding, height - padding);
        ctx.stroke();
    }

    // 生成圖表資料
    generateChartData(points) {
        const data = [];
        const baseValue = 100;
        
        for (let i = 0; i < points; i++) {
            const timeVariation = Math.sin((i / points) * Math.PI * 2) * 30;
            const randomVariation = Math.random() * 20 - 10;
            const value = baseValue + timeVariation + randomVariation;
            
            data.push({
                time: i,
                value: Math.max(0, value)
            });
        }
        
        return data;
    }

    // 更新圖表資料
    updateChartData() {
        // 重新創建所有圖表
        this.createCharts();
    }

    // 開始自動更新
    startAutoUpdate() {
        console.log('節能看板已啟動，每5秒自動更新數據');
    }
}

// 工具函數
class Utils {
    // 格式化數值
    static formatNumber(value, decimals = 1) {
        return parseFloat(value).toFixed(decimals);
    }

    // 格式化大數值
    static formatLargeNumber(value) {
        if (value >= 1000000) {
            return (value / 1000000).toFixed(1) + 'M';
        } else if (value >= 1000) {
            return (value / 1000).toFixed(1) + 'K';
        }
        return value.toString();
    }

    // 生成隨機顏色
    static getRandomColor() {
        const colors = [
            '#f1c40f', '#f39c12', '#e67e22', '#d35400',
            '#27ae60', '#2ecc71', '#16a085', '#1abc9c',
            '#3498db', '#2980b9', '#9b59b6', '#8e44ad'
        ];
        return colors[Math.floor(Math.random() * colors.length)];
    }
}

// 初始化儀表板
document.addEventListener('DOMContentLoaded', function() {
    window.energyDashboard = new EnergyDashboard();
    
    // 綁定樹木倍數選擇器事件
    const scaleItems = document.querySelectorAll('.scale-item');
    scaleItems.forEach(item => {
        item.addEventListener('click', function() {
            scaleItems.forEach(si => si.classList.remove('active'));
            this.classList.add('active');
            
            // 更新樹木數量顯示
            const multiplier = this.querySelector('span').textContent;
            const baseValue = 9999;
            let newValue;
            
            switch(multiplier) {
                case 'x1': newValue = baseValue; break;
                case 'x10': newValue = baseValue * 10; break;
                case 'x100': newValue = baseValue * 100; break;
                case 'x1000': newValue = baseValue * 1000; break;
                default: newValue = baseValue;
            }
            
            document.querySelector('.trees-number').textContent = Utils.formatLargeNumber(newValue);
        });
    });
    
    console.log('Think Power 節能看板系統已載入完成');
}); 