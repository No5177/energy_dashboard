# 能源監控系統 v2.0

完整的電表資料收集、儲存與視覺化系統，整合 Modbus TCP、SQLite3 和 Chart.js。

## 🎯 系統功能

### 核心功能
- **🔄 自動資料收集**: 每 5 秒透過 Modbus TCP 向台達電表收集資料
- **💾 資料庫儲存**: 使用 SQLite3 儲存歷史資料，支援高效查詢
- **📊 即時監控**: 網頁介面即時顯示電表參數
- **📈 趨勢分析**: 多時間軸曲線圖分析 (每日/每月/每季/每年)
- **🌐 HTTP API**: RESTful API 提供資料查詢服務

### 監控參數
1. **相電壓平均值** (V)
2. **三相平均電流** (A)
3. **頻率** (Hz)
4. **三相正向實功率** (kW)
5. **三相反向實功率** (kW)
6. **線實功率因數**
7. **電流諧波失真率** (%)

## 🏗️ 系統架構

```
台達電表 (192.168.1.9:502)
    ↓ Modbus TCP
Go 後端服務 (energy_backend.go)
    ↓ 每5秒收集
SQLite3 資料庫 (energy_data.db)
    ↓ HTTP API
網頁前端 (energy_dashboard.html)
    ↓ Chart.js
用戶界面 (localhost:8080)
```

## 📋 環境需求

### 開發環境
- **操作系統**: Windows 10/11
- **Go 語言**: 1.19 或更新版本
- **GCC 編譯器**: 支援 CGO (用於 SQLite3)
- **網路連線**: 下載相依套件

### 運行環境
- **Windows 10/11**
- **現代瀏覽器**: Chrome/Firefox/Edge
- **網路連線**: 與電表 192.168.1.9 的連線

## 🚀 快速開始

### 1. 安裝依賴

#### 安裝 Go 語言
1. 下載: https://golang.org/dl/
2. 安裝並設定環境變數
3. 驗證: `go version`

#### 安裝 GCC (如果需要)
- Windows: 安裝 TDM-GCC 或 MinGW-w64
- 或使用 Chocolatey: `choco install mingw`

### 2. 啟動系統

```batch
# 直接執行啟動腳本
start_energy_system.bat
```

系統會自動:
1. ✅ 檢查環境依賴
2. 📦 下載 Go 套件
3. 🔨 編譯程式
4. 🚀 啟動服務
5. 🌐 開啟瀏覽器

### 3. 訪問界面

- **主儀表板**: http://localhost:8080/energy_dashboard.html
- **API 端點**: 
  - 最新資料: http://localhost:8080/api/latest
  - 聚合資料: http://localhost:8080/api/aggregated

## 📊 使用說明

### 即時監控
- 自動顯示最新的電表參數
- 每 5 秒更新數據
- 圓餅圖顯示功率因子等百分比資料

### 趨勢分析
1. **選擇時間範圍**: 每日/每月/每季/每年
2. **選擇監控參數**: 從下拉選單選擇要分析的參數
3. **設定日期範圍**: 根據選擇的時間範圍設定具體日期
4. **更新圖表**: 點擊"更新圖表"按鈕

#### 時間軸說明
- **每日視圖**: 顯示 24 小時資料，橫軸標示每小時
- **每月視圖**: 顯示整月資料，橫軸標示每日
- **每季視圖**: 顯示 3 個月資料，橫軸標示每月
- **每年視圖**: 顯示 12 個月資料，橫軸標示每月

## 🔧 技術規格

### 後端 (Go)
- **Framework**: 原生 `net/http`
- **資料庫**: SQLite3 (`github.com/mattn/go-sqlite3`)
- **Modbus**: `github.com/goburrow/modbus`
- **CORS**: `github.com/rs/cors`

### 前端 (HTML/CSS/JS)
- **圖表庫**: Chart.js
- **樣式**: 自定義 CSS，響應式設計
- **資料格式**: JSON

### 資料庫結構
```sql
CREATE TABLE meter_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    device_id TEXT NOT NULL,
    json_data TEXT NOT NULL
);

-- 索引優化
CREATE INDEX idx_timestamp ON meter_data(timestamp);
CREATE INDEX idx_device_id ON meter_data(device_id);
CREATE INDEX idx_timestamp_device ON meter_data(timestamp, device_id);
```

## 🔌 API 接口

### 1. 獲取最新資料
```http
GET /api/latest
```

**回應範例**:
```json
[
  {
    "index": 0,
    "name": "相電壓平均值",
    "value": 220.5,
    "unit": "V"
  },
  {
    "index": 1,
    "name": "三相平均電流",
    "value": 5.2,
    "unit": "A"
  }
]
```

### 2. 獲取聚合資料
```http
GET /api/aggregated?range=daily&date=2025-01-15&parameter=相電壓平均值
```

**參數說明**:
- `range`: 時間範圍 (`daily`/`monthly`/`quarterly`/`yearly`)
- `date`: 日期參數 (格式依 range 而異)
- `parameter`: 參數名稱

**回應範例**:
```json
[
  {
    "timestamp": "2025-01-15T00:00:00Z",
    "parameter": "相電壓平均值",
    "avg_value": 220.5,
    "min_value": 218.2,
    "max_value": 222.8
  }
]
```

## 🛠️ 故障排除

### 常見問題

**Q: 編譯時出現 CGO 錯誤**
```
A: 確保安裝了 GCC 編譯器
   Windows: 安裝 TDM-GCC 或 MinGW-w64
   設定環境變數 CGO_ENABLED=1
```

**Q: 無法連接到電表**
```
A: 檢查網路連線和電表設定
   1. ping 192.168.1.9
   2. 確認電表 Modbus TCP 服務啟用
   3. 檢查防火牆設定
```

**Q: 網頁無法載入資料**
```
A: 檢查後端服務狀態
   1. 確認 http://localhost:8080 可訪問
   2. 檢查控制台錯誤訊息
   3. 驗證 API 端點回應
```

**Q: 圖表顯示異常**
```
A: 檢查瀏覽器相容性和網路連線
   1. 確保 Chart.js CDN 可訪問
   2. 檢查瀏覽器控制台錯誤
   3. 清除瀏覽器快取
```

## 📁 檔案結構

```
專案目錄/
├── energy_backend.go              # Go 後端主程式
├── energy_dashboard.html          # 網頁前端
├── css/
│   └── energy_dashboard.css       # 樣式檔案
├── go.mod                         # Go 模組管理
├── go.sum                         # 依賴版本鎖定
├── start_energy_system.bat        # 啟動腳本
├── README_ENERGY_MONITORING.md    # 本文件
└── energy_data.db                 # SQLite 資料庫 (自動生成)
```

## 🔄 資料流程

1. **資料收集**: Go 後端每 5 秒透過 Modbus TCP 讀取電表資料
2. **資料解析**: 解析 IEEE754 浮點數格式 (Word-Swap)
3. **資料儲存**: JSON 格式儲存至 SQLite3 資料庫
4. **API 服務**: HTTP API 提供即時和歷史資料查詢
5. **前端顯示**: 網頁透過 AJAX 獲取資料並用 Chart.js 繪製圖表

## 🎨 客製化

### 修改監控參數
編輯 `energy_backend.go` 中的 `meterParameters` 變數:
```go
var meterParameters = []MeterParameter{
    {"自定義參數", 0x0XXX, "單位"},
    // 添加更多參數...
}
```

### 調整收集頻率
修改 `StartDataCollection()` 中的 ticker:
```go
ticker := time.NewTicker(10 * time.Second) // 改為 10 秒
```

### 自定義圖表樣式
編輯 `css/energy_dashboard.css` 中的圖表相關樣式。

## 📈 效能優化

### 資料庫優化
- ✅ 已建立時間戳和設備 ID 索引
- ✅ 使用參數化查詢防止 SQL Injection
- ✅ 資料壓縮和聚合查詢

### 前端優化
- ✅ 資料快取機制
- ✅ 圖表重用和銷毀
- ✅ 響應式設計

### 網路優化
- ✅ CORS 支援
- ✅ 靜態檔案服務
- ✅ API 錯誤處理

---

## 📞 技術支援

如有問題，請檢查:
1. 控制台錯誤訊息
2. 網路連線狀態
3. 電表設定參數
4. Go 環境配置

系統設計充分考慮了可擴展性和維護性，支援多設備監控和客製化需求。 