# 台達電表 Modbus TCP 通訊分析報告

## 電表資訊
- **IP 地址**: 192.168.1.9
- **子網路遮罩**: 255.255.255.0  
- **閘道**: 192.168.1.1
- **通訊位址 (Slave ID)**: 2
- **Modbus TCP 埠號**: 502

## 測試結果

### ✅ 連線狀態
程式成功連接到電表，Modbus TCP 通訊正常。

### 📊 資料讀取結果

#### 1. 相電壓平均值 (暫存器地址: 0x0106)
- **狀態**: ✅ 成功讀取
- **資料格式**: IEEE754 32-bit 浮點數 (佔用2個暫存器)
- **字節順序**: Word-Swap (CDAB) 格式
- **讀取值**: 約 117V (正常電壓範圍)
- **原始資料範例**: `[0x199A, 0x42EA]` → 117.050 V

#### 2. 三相平均電流 (暫存器地址: 0x0126)  
- **狀態**: ✅ 成功讀取
- **資料格式**: IEEE754 32-bit 浮點數
- **讀取值**: 0.000 A (目前無負載)
- **原始資料**: `[0x0000, 0x0000]`

#### 3. 其他參數測試
| 參數 | 地址 | 狀態 | 說明 |
|------|------|------|------|
| 總有效功率 | 0x0140 | ❌ | 回傳 0xFFFFFFFF (無效值) |
| 功率因子 | 0x0150 | ❌ | 回傳 0xFFFFFFFF (無效值) |
| 頻率 | 0x0160 | ❌ | 回傳 0x00000000 (可能地址錯誤) |
| 總電能 | 0x0200 | ❌ | 回傳 0xFFFFFFFF (無效值) |

## 🔍 技術分析

### 浮點數格式解析
電表使用 **Word-Swap (CDAB)** 字節順序：
```
原始暫存器: [0x199A, 0x42EA]
重新排列: [0x42EA, 0x199A] 
轉換為浮點數: 117.050 V
```

### Modbus 通訊日誌
```
發送: 00 01 00 00 00 06 02 03 01 06 00 02
接收: 00 01 00 00 00 07 02 03 04 19 9A 42 EA
```

**封包解析**:
- `00 01`: 事務 ID
- `00 00`: 協議 ID
- `00 06`: 資料長度
- `02`: 單元 ID (通訊位址)
- `03`: 功能碼 (讀取保持暫存器)
- `01 06`: 起始地址 (0x0106)
- `00 02`: 暫存器數量 (2個)

## 💡 程式功能說明

### ModBus_request_API.go
- **功能**: 提供 Modbus 客戶端類別和通訊方法
- **主要結構**: `ModbusClient` 封裝連線設定和資料讀取
- **改進**: 已修正原始程式的錯誤，支援實際電表參數

### modbus_client.go  
- **功能**: 完整的測試程式，直接執行 Modbus 通訊
- **特色**: 
  - 自動嘗試不同字節順序解析
  - 顯示原始資料便於除錯
  - 連續讀取多次數據
  - 測試多個電表參數

### test_modbus.bat
- **功能**: 一鍵編譯並執行測試
- **自動化**: 編譯 → 執行 → 清理

## 🚀 使用方式

### 方法1: 直接執行 Go 程式
```bash
go run modbus_client.go
```

### 方法2: 使用批次檔
```bash
test_modbus.bat
```

### 方法3: 編譯後執行
```bash
go build -o modbus_test.exe modbus_client.go
modbus_test.exe
```

## 📋 建議與改進

### 1. 確認正確的暫存器地址
目前部分參數回傳無效值 (0xFFFFFFFF)，建議：
- 查閱電表詳細手冊確認正確地址
- 測試不同的地址範圍
- 確認參數是否需要特殊權限

### 2. 資料類型優化
某些參數可能不是浮點數格式：
- 整數參數 (16-bit, 32-bit)
- 定點小數
- 狀態位元

### 3. 錯誤處理增強
- 增加連線重試機制
- 加入資料驗證
- 記錄詳細錯誤訊息

### 4. 整合到主系統
可將 Modbus 客戶端整合到現有的 `main.go` 能源看板系統：
- 替代或補充現有的 TCP 通訊
- 提供更可靠的電表資料來源
- 支援多台電表同時監控

## 🔧 疑難排解

### 連線失敗
- 檢查電表 IP 和網路設定
- 確認防火牆設定
- 驗證 Modbus TCP 服務是否啟用

### 資料異常
- 確認暫存器地址正確性
- 檢查字節順序設定
- 驗證資料類型假設

### 效能問題
- 調整讀取間隔
- 使用批次讀取減少通訊次數
- 實施連線池管理

---

## 總結

✅ **Modbus TCP 通訊成功建立**  
✅ **電壓資料正常讀取 (~117V)**  
✅ **程式架構完整且可擴展**  
⚠️ **部分參數地址需要確認**  
⚠️ **需要參考電表手冊完善功能**

此 Modbus 客戶端為能源監控系統提供了可靠的電表通訊基礎，可作為現有系統的重要補充或替代方案。 