// 彈窗控制器 - 整合設定和工程模式功能
class ModalController {
    constructor() {
        this.engineeringModeVisible = false;
        this.connectionStatus = {
            consumption: false,
            recovery: false
        };
        this.init();
    }

    init() {
        this.setupKeyboardShortcuts();
        this.setupModalEvents();
    }

    // ==================== 彈窗控制 ====================

    openSettingsModal() {
        const modal = document.getElementById('settingsModal');
        if (modal) {
            modal.style.display = 'flex';
            setTimeout(() => {
                modal.classList.add('show');
            }, 10);
            
            this.loadSettings();
            // 重置工程模式按鈕狀態
            this.engineeringModeVisible = false;
            document.getElementById('engineeringModeBtn').style.display = 'none';
        }
    }

    closeSettingsModal() {
        const modal = document.getElementById('settingsModal');
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.style.display = 'none';
            }, 300);
        }
    }

    openEngineeringModal() {
        const modal = document.getElementById('engineeringModal');
        if (modal) {
            modal.style.display = 'flex';
            setTimeout(() => {
                modal.classList.add('show');
            }, 10);
            
            this.loadEngineeringSettings();
            this.updateEngineeringTime();
            this.checkInitialConnections();
        }
    }

    closeEngineeringModal() {
        const modal = document.getElementById('engineeringModal');
        if (modal) {
            modal.classList.remove('show');
            setTimeout(() => {
                modal.style.display = 'none';
            }, 300);
        }
    }

    // 設定彈窗事件
    setupModalEvents() {
        // 點擊遮罩關閉彈窗
        document.getElementById('settingsModal').addEventListener('click', (e) => {
            if (e.target.id === 'settingsModal') {
                this.closeSettingsModal();
            }
        });

        document.getElementById('engineeringModal').addEventListener('click', (e) => {
            if (e.target.id === 'engineeringModal') {
                this.closeEngineeringModal();
            }
        });

        // ESC鍵關閉彈窗
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeSettingsModal();
                this.closeEngineeringModal();
            }
        });
    }

    // ==================== 設定功能 ====================

    loadSettings() {
        const settings = this.getStoredSettings();
        
        const workStationInput = document.getElementById('workStationName');
        const voltageLogSelect = document.getElementById('voltageLogTime');
        
        if (workStationInput && settings.workStationName) {
            workStationInput.value = settings.workStationName;
        }
        
        if (voltageLogSelect && settings.voltageLogTime) {
            voltageLogSelect.value = settings.voltageLogTime;
        }
    }

    getStoredSettings() {
        const defaultSettings = {
            workStationName: 'Work Station Name',
            voltageLogTime: '5'
        };
        
        try {
            const stored = localStorage.getItem('energyDashboardSettings');
            return stored ? { ...defaultSettings, ...JSON.parse(stored) } : defaultSettings;
        } catch (error) {
            console.warn('讀取設定失敗:', error);
            return defaultSettings;
        }
    }

    saveSettings() {
        const workStationName = document.getElementById('workStationName')?.value || 'Work Station Name';
        const voltageLogTime = document.getElementById('voltageLogTime')?.value || '5';
        
        const settings = {
            workStationName,
            voltageLogTime,
            savedAt: new Date().toISOString()
        };
        
        try {
            localStorage.setItem('energyDashboardSettings', JSON.stringify(settings));
            this.showNotification('設定已儲存', 'success');
            
            // 更新主頁面工作站名稱
            const stationNameElement = document.querySelector('.station-name h1');
            if (stationNameElement) {
                stationNameElement.textContent = workStationName;
            }
            
        } catch (error) {
            console.error('儲存設定失敗:', error);
            this.showNotification('儲存失敗', 'error');
        }
    }

    // ==================== 工程模式功能 ====================

    loadEngineeringSettings() {
        const settings = this.getStoredEngineeringSettings();
        
        document.getElementById('consumptionMeterIP').value = settings.consumptionMeterIP;
        document.getElementById('consumptionMeterPort').value = settings.consumptionMeterPort;
        document.getElementById('consumptionMeterID').value = settings.consumptionMeterID;
        document.getElementById('recoveryMeterIP').value = settings.recoveryMeterIP;
        document.getElementById('recoveryMeterPort').value = settings.recoveryMeterPort;
        document.getElementById('recoveryMeterID').value = settings.recoveryMeterID;
    }

    getStoredEngineeringSettings() {
        const defaultSettings = {
            consumptionMeterIP: '192.168.1.9',
            consumptionMeterPort: '502',
            consumptionMeterID: '2',
            recoveryMeterIP: '192.168.1.10',
            recoveryMeterPort: '502',
            recoveryMeterID: '3'
        };
        
        try {
            const stored = localStorage.getItem('engineeringSettings');
            return stored ? { ...defaultSettings, ...JSON.parse(stored) } : defaultSettings;
        } catch (error) {
            console.warn('讀取工程設定失敗:', error);
            return defaultSettings;
        }
    }

    saveEngineeringSettings() {
        const settings = {
            consumptionMeterIP: document.getElementById('consumptionMeterIP').value,
            consumptionMeterPort: document.getElementById('consumptionMeterPort').value,
            consumptionMeterID: document.getElementById('consumptionMeterID').value,
            recoveryMeterIP: document.getElementById('recoveryMeterIP').value,
            recoveryMeterPort: document.getElementById('recoveryMeterPort').value,
            recoveryMeterID: document.getElementById('recoveryMeterID').value,
            savedAt: new Date().toISOString()
        };
        
        try {
            localStorage.setItem('engineeringSettings', JSON.stringify(settings));
            this.showNotification('工程設定已儲存', 'success');
        } catch (error) {
            console.error('儲存工程設定失敗:', error);
            this.showNotification('儲存失敗', 'error');
        }
    }

    // 更新工程模式時間顯示
    updateEngineeringTime() {
        const now = new Date();
        
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        const hours = String(now.getHours()).padStart(2, '0');
        const minutes = String(now.getMinutes()).padStart(2, '0');
        const seconds = String(now.getSeconds()).padStart(2, '0');
        
        const timeStr = `${year}/${month}/${day} ${hours}:${minutes}:${seconds}`;
        const systemTimeElement = document.getElementById('systemTime');
        
        if (systemTimeElement) {
            systemTimeElement.textContent = timeStr;
        }

        // 模擬電表時間 (可能有些微差異)
        const meterTime = new Date(now.getTime() + Math.random() * 2000 - 1000);
        const meterTimeStr = `${meterTime.getFullYear()}/${String(meterTime.getMonth() + 1).padStart(2, '0')}/${String(meterTime.getDate()).padStart(2, '0')} ${String(meterTime.getHours()).padStart(2, '0')}:${String(meterTime.getMinutes()).padStart(2, '0')}:${String(meterTime.getSeconds()).padStart(2, '0')}`;
        
        const meterTimeElement = document.getElementById('meterTime');
        if (meterTimeElement) {
            meterTimeElement.textContent = meterTimeStr;
        }
    }

    // 同步電表時間
    async syncMeterTime() {
        this.showNotification('正在同步電表時間...', 'info');
        
        try {
            await this.simulateAsyncOperation(2000);
            
            const now = new Date();
            const timeStr = `${now.getFullYear()}/${String(now.getMonth() + 1).padStart(2, '0')}/${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`;
            document.getElementById('meterTime').textContent = timeStr;
            
            this.showNotification('電表時間同步成功', 'success');
        } catch (error) {
            console.error('電表時間同步失敗:', error);
            this.showNotification('電表時間同步失敗', 'error');
        }
    }

    // 重新讀取電表時間
    async refreshMeterTime() {
        this.showNotification('正在讀取電表時間...', 'info');
        
        try {
            await this.simulateAsyncOperation(1000);
            this.updateEngineeringTime();
            this.showNotification('電表時間已更新', 'success');
        } catch (error) {
            console.error('讀取電表時間失敗:', error);
            this.showNotification('讀取電表時間失敗', 'error');
        }
    }

    // 測試連接
    async testConnection(type) {
        const statusElement = document.getElementById(`${type}Status`);
        const ipInput = document.getElementById(`${type}MeterIP`);
        const portInput = document.getElementById(`${type}MeterPort`);
        
        if (!statusElement || !ipInput || !portInput) return;
        
        const ip = ipInput.value;
        const port = portInput.value;
        
        this.showNotification(`正在測試${type === 'consumption' ? '耗能' : '回收'}電表連接...`, 'info');
        
        try {
            const success = await this.simulateConnectionTest(ip, port);
            
            if (success) {
                statusElement.textContent = '在線';
                statusElement.className = 'status-indicator online';
                this.connectionStatus[type] = true;
                this.showNotification(`${type === 'consumption' ? '耗能' : '回收'}電表連接成功`, 'success');
            } else {
                throw new Error('連接失敗');
            }
        } catch (error) {
            statusElement.textContent = '離線';
            statusElement.className = 'status-indicator offline';
            this.connectionStatus[type] = false;
            this.showNotification(`${type === 'consumption' ? '耗能' : '回收'}電表連接失敗`, 'error');
        }
    }

    // 模擬連接測試
    async simulateConnectionTest(ip, port) {
        await this.simulateAsyncOperation(2000);
        return Math.random() > 0.2; // 80% 成功率
    }

    // 檢查初始連接狀態
    async checkInitialConnections() {
        setTimeout(() => {
            this.testConnection('consumption');
        }, 1000);
        
        setTimeout(() => {
            this.testConnection('recovery');
        }, 2000);
    }

    // ==================== 鍵盤快捷鍵 ====================

    setupKeyboardShortcuts() {
        let ctrlPressed = false;
        let shiftPressed = false;
        
        document.addEventListener('keydown', (event) => {
            if (event.ctrlKey) ctrlPressed = true;
            if (event.shiftKey) shiftPressed = true;
            
            // Ctrl + Shift + F10 (僅在設定彈窗開啟時生效)
            if (ctrlPressed && shiftPressed && event.key === 'F10') {
                const settingsModal = document.getElementById('settingsModal');
                if (settingsModal && settingsModal.classList.contains('show')) {
                    event.preventDefault();
                    this.toggleEngineeringMode();
                }
            }
        });
        
        document.addEventListener('keyup', (event) => {
            if (!event.ctrlKey) ctrlPressed = false;
            if (!event.shiftKey) shiftPressed = false;
        });
    }

    toggleEngineeringMode() {
        const engineeringBtn = document.getElementById('engineeringModeBtn');
        
        if (engineeringBtn) {
            this.engineeringModeVisible = !this.engineeringModeVisible;
            engineeringBtn.style.display = this.engineeringModeVisible ? 'flex' : 'none';
            
            if (this.engineeringModeVisible) {
                this.showNotification('🔧 工程模式已啟用', 'info');
                engineeringBtn.style.animation = 'engineeringAppear 0.5s ease-out';
            }
        }
    }

    // ==================== 通用功能 ====================

    simulateAsyncOperation(delay) {
        return new Promise((resolve) => {
            setTimeout(resolve, delay);
        });
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 600;
            z-index: 10001;
            transition: all 0.3s ease;
            transform: translateX(100%);
        `;
        
        const colors = {
            success: '#27ae60',
            error: '#e74c3c',
            info: '#3498db',
            warning: '#f39c12'
        };
        
        notification.style.background = colors[type] || colors.info;
        notification.style.boxShadow = `0 4px 15px ${colors[type]}40`;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }
}

// 全域彈窗控制器實例
let modalController;

// 全域函數
function openSettingsModal() {
    if (modalController) {
        modalController.openSettingsModal();
    }
}

function closeSettingsModal() {
    if (modalController) {
        modalController.closeSettingsModal();
    }
}

function openEngineeringModal() {
    if (modalController) {
        modalController.openEngineeringModal();
    }
}

function closeEngineeringModal() {
    if (modalController) {
        modalController.closeEngineeringModal();
    }
}

function saveSettings() {
    if (modalController) {
        modalController.saveSettings();
    }
}

function saveEngineeringSettings() {
    if (modalController) {
        modalController.saveEngineeringSettings();
    }
}

function syncMeterTime() {
    if (modalController) {
        modalController.syncMeterTime();
    }
}

function refreshMeterTime() {
    if (modalController) {
        modalController.refreshMeterTime();
    }
}

function testConnection(type) {
    if (modalController) {
        modalController.testConnection(type);
    }
}

// 初始化彈窗控制器
document.addEventListener('DOMContentLoaded', function() {
    modalController = new ModalController();
    
    // 啟動工程模式時間更新（僅在工程模式彈窗開啟時）
    setInterval(() => {
        const engineeringModal = document.getElementById('engineeringModal');
        if (engineeringModal && engineeringModal.classList.contains('show')) {
            modalController.updateEngineeringTime();
        }
    }, 1000);
    
    console.log('彈窗控制器已載入完成');
}); 