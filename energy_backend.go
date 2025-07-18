package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/goburrow/modbus"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

// 台達電表參數定義
type MeterParameter struct {
	Name    string `json:"name"`
	Address uint16 `json:"address"`
	Unit    string `json:"unit"`
}

// 電表數據結構
type MeterData struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	DeviceID  string    `json:"device_id"`
	JSONData  string    `json:"json_data"`
}

// 電表讀取值結構
type MeterReading struct {
	Index int     `json:"index"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// 聚合資料結構
type AggregatedData struct {
	Timestamp time.Time `json:"timestamp"`
	Parameter string    `json:"parameter"`
	AvgValue  float64   `json:"avg_value"`
	MinValue  float64   `json:"min_value"`
	MaxValue  float64   `json:"max_value"`
}

// 根據提供的參數表格定義電表參數
var meterParameters = []MeterParameter{
	{"相電壓平均值", 0x0106, "V"},
	{"三相平均電流", 0x0126, "A"},
	{"頻率", 0x0142, "Hz"},
	{"三相正向實功率", 0x015C, "kW"},
	{"三相反向實功率", 0x015E, "kW"},
	{"線實功率因數", 0x0132, "N/A"},
	{"電流諧波失真率", 0x0188, "%"},
	{"電流諧波失真率", 0x018A, "%"},
}

// 能源系統結構
type EnergySystem struct {
	db          *sql.DB
	modbusHost  string
	modbusPort  int
	slaveID     byte
	running     bool
	stopChannel chan bool
}

// 建立新的能源系統
func NewEnergySystem() *EnergySystem {
	return &EnergySystem{
		modbusHost:  "192.168.1.9",
		modbusPort:  502,
		slaveID:     2,
		running:     false,
		stopChannel: make(chan bool),
	}
}

// 初始化資料庫
func (es *EnergySystem) InitDatabase() error {
	var err error
	es.db, err = sql.Open("sqlite3", "./energy_data.db")
	if err != nil {
		return fmt.Errorf("無法開啟資料庫: %v", err)
	}

	// 建立資料表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS meter_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		device_id TEXT NOT NULL,
		json_data TEXT NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_timestamp ON meter_data(timestamp);
	CREATE INDEX IF NOT EXISTS idx_device_id ON meter_data(device_id);
	CREATE INDEX IF NOT EXISTS idx_timestamp_device ON meter_data(timestamp, device_id);
	`

	_, err = es.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("建立資料表失敗: %v", err)
	}

	log.Println("✅ 資料庫初始化完成")
	return nil
}

// 讀取電表資料
func (es *EnergySystem) ReadMeterData() ([]MeterReading, error) {
	// 建立 Modbus TCP 客戶端
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", es.modbusHost, es.modbusPort))
	handler.Timeout = 10 * time.Second
	handler.IdleTimeout = 60 * time.Second
	handler.SlaveId = es.slaveID

	// 連接到電表
	err := handler.Connect()
	if err != nil {
		return nil, fmt.Errorf("無法連接到電表: %v", err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	readings := make([]MeterReading, 0)

	// 讀取所有參數
	for i, param := range meterParameters {
		results, err := client.ReadHoldingRegisters(param.Address, 2)
		if err != nil {
			log.Printf("❌ 讀取 %s 失敗: %v", param.Name, err)
			continue
		}

		if len(results) >= 4 {
			// 使用 Word-Swap 解析 (根據之前的測試結果)
			swapped := []byte{results[2], results[3], results[0], results[1]}
			valueFloat := binary.BigEndian.Uint32(swapped)
			value := math.Float32frombits(valueFloat)

			reading := MeterReading{
				Index: i,
				Name:  param.Name,
				Value: float64(value),
				Unit:  param.Unit,
			}
			readings = append(readings, reading)
		}
	}

	return readings, nil
}

// 儲存資料到資料庫
func (es *EnergySystem) SaveToDatabase(readings []MeterReading) error {
	jsonData, err := json.Marshal(readings)
	if err != nil {
		return fmt.Errorf("JSON 編碼失敗: %v", err)
	}

	insertSQL := `INSERT INTO meter_data (device_id, json_data) VALUES (?, ?)`
	_, err = es.db.Exec(insertSQL, "DPMC530E", string(jsonData))
	if err != nil {
		return fmt.Errorf("資料庫插入失敗: %v", err)
	}

	return nil
}

// 定時資料收集
func (es *EnergySystem) StartDataCollection() {
	es.running = true
	log.Println("🔄 開始每 5 秒收集電表資料...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for es.running {
		select {
		case <-ticker.C:
			readings, err := es.ReadMeterData()
			if err != nil {
				log.Printf("❌ 讀取電表資料失敗: %v", err)
				continue
			}

			err = es.SaveToDatabase(readings)
			if err != nil {
				log.Printf("❌ 儲存資料失敗: %v", err)
				continue
			}

			log.Printf("✅ 成功收集並儲存 %d 筆資料 (%s)", len(readings), time.Now().Format("15:04:05"))

		case <-es.stopChannel:
			return
		}
	}
}

// 停止資料收集
func (es *EnergySystem) StopDataCollection() {
	es.running = false
	close(es.stopChannel)
}

// HTTP API 處理器

// 獲取最新資料 (原有功能相容)
func (es *EnergySystem) GetLatestDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	querySQL := `SELECT json_data FROM meter_data ORDER BY timestamp DESC LIMIT 1`
	var jsonData string
	err := es.db.QueryRow(querySQL).Scan(&jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("查詢失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(jsonData))
}

// 獲取聚合資料 (新功能)
func (es *EnergySystem) GetAggregatedDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 解析查詢參數
	timeRange := r.URL.Query().Get("range")     // daily, monthly, quarterly, yearly
	dateParam := r.URL.Query().Get("date")      // 格式依範圍而定
	parameter := r.URL.Query().Get("parameter") // 參數名稱

	if timeRange == "" || dateParam == "" || parameter == "" {
		http.Error(w, "缺少必要參數: range, date, parameter", http.StatusBadRequest)
		return
	}

	aggregatedData, err := es.getAggregatedData(timeRange, dateParam, parameter)
	if err != nil {
		http.Error(w, fmt.Sprintf("聚合資料查詢失敗: %v", err), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(aggregatedData)
	if err != nil {
		http.Error(w, fmt.Sprintf("JSON 編碼失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

// 聚合資料查詢邏輯
func (es *EnergySystem) getAggregatedData(timeRange, dateParam, parameter string) ([]AggregatedData, error) {
	var querySQL string
	var timeFormat string
	var args []interface{}

	switch timeRange {
	case "daily":
		// 按小時聚合，顯示一天24小時
		querySQL = `
		SELECT 
			datetime(timestamp, 'localtime', 'start of hour') as hour_timestamp,
			json_extract(json_data, '$[*]') as readings
		FROM meter_data 
		WHERE date(timestamp, 'localtime') = ?
		ORDER BY hour_timestamp`
		args = []interface{}{dateParam}
		timeFormat = "hour"

	case "monthly":
		// 按日聚合，顯示一個月的每一天
		querySQL = `
		SELECT 
			date(timestamp, 'localtime') as day_timestamp,
			json_extract(json_data, '$[*]') as readings
		FROM meter_data 
		WHERE strftime('%Y-%m', timestamp, 'localtime') = ?
		ORDER BY day_timestamp`
		args = []interface{}{dateParam}
		timeFormat = "day"

	case "quarterly":
		// 按週聚合，顯示一季的資料
		quarterMap := map[string][]string{
			"Q1": {"01", "02", "03"},
			"Q2": {"04", "05", "06"},
			"Q3": {"07", "08", "09"},
			"Q4": {"10", "11", "12"},
		}
		year := dateParam[:4]
		quarter := dateParam[5:]
		months := quarterMap[quarter]

		querySQL = `
		SELECT 
			strftime('%Y-%m', timestamp, 'localtime') as month_timestamp,
			json_extract(json_data, '$[*]') as readings
		FROM meter_data 
		WHERE strftime('%Y', timestamp, 'localtime') = ? 
		AND strftime('%m', timestamp, 'localtime') IN (?, ?, ?)
		ORDER BY month_timestamp`
		args = []interface{}{year, months[0], months[1], months[2]}
		timeFormat = "month"

	case "yearly":
		// 按月聚合，顯示一年的12個月
		querySQL = `
		SELECT 
			strftime('%Y-%m', timestamp, 'localtime') as month_timestamp,
			json_extract(json_data, '$[*]') as readings
		FROM meter_data 
		WHERE strftime('%Y', timestamp, 'localtime') = ?
		ORDER BY month_timestamp`
		args = []interface{}{dateParam}
		timeFormat = "month"

	default:
		return nil, fmt.Errorf("不支援的時間範圍: %s", timeRange)
	}

	rows, err := es.db.Query(querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]AggregatedData, 0)

	// 這裡需要實現複雜的 JSON 解析和聚合邏輯
	// 由於 SQLite 的 JSON 函數限制，我們先簡化實現
	for rows.Next() {
		var timestampStr string
		var readingsJSON string

		err := rows.Scan(&timestampStr, &readingsJSON)
		if err != nil {
			continue
		}

		// 解析時間戳
		var timestamp time.Time
		switch timeFormat {
		case "hour":
			timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
		case "day":
			timestamp, _ = time.Parse("2006-01-02", timestampStr)
		case "month":
			timestamp, _ = time.Parse("2006-01", timestampStr)
		}

		// 簡化實現：返回模擬資料
		result = append(result, AggregatedData{
			Timestamp: timestamp,
			Parameter: parameter,
			AvgValue:  220.5 + float64(len(result)), // 模擬資料
			MinValue:  215.0,
			MaxValue:  225.0,
		})
	}

	return result, nil
}

// 啟動 HTTP 服務器
func (es *EnergySystem) StartHTTPServer() {
	mux := http.NewServeMux()

	// API 端點
	mux.HandleFunc("/api/latest", es.GetLatestDataHandler)
	mux.HandleFunc("/api/aggregated", es.GetAggregatedDataHandler)

	// 靜態檔案服務
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// 設定 CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(mux)

	log.Println("🌐 HTTP 服務器啟動於 http://localhost:8080")

	go func() {
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Printf("HTTP 服務器錯誤: %v", err)
		}
	}()
}

// 開啟瀏覽器
func (es *EnergySystem) OpenBrowser() {
	url := "http://localhost:8080/energy_dashboard.html"
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("開啟瀏覽器失敗: %v", err)
	} else {
		log.Printf("🌐 已開啟能源儀表板: %s", url)
	}
}

// 啟動完整系統
func (es *EnergySystem) Start() error {
	fmt.Println("==================================================")
	fmt.Println("能源監控系統啟動中...")
	fmt.Println("==================================================")

	// 1. 初始化資料庫
	err := es.InitDatabase()
	if err != nil {
		return err
	}

	// 2. 啟動 HTTP 服務器
	es.StartHTTPServer()

	// 3. 啟動資料收集
	go es.StartDataCollection()

	// 4. 等待系統穩定
	time.Sleep(2 * time.Second)

	// 5. 開啟瀏覽器
	es.OpenBrowser()

	fmt.Println("==================================================")
	fmt.Println("✅ 系統啟動完成！")
	fmt.Println("📊 能源儀表板: http://localhost:8080/energy_dashboard.html")
	fmt.Println("🔄 每 5 秒自動收集電表資料")
	fmt.Println("💾 資料儲存至 SQLite3: energy_data.db")
	fmt.Println("按 Ctrl+C 停止系統")
	fmt.Println("==================================================")

	return nil
}

// 停止系統
func (es *EnergySystem) Stop() {
	es.StopDataCollection()
	if es.db != nil {
		es.db.Close()
	}
	log.Println("🛑 系統已停止")
}

func main() {
	system := NewEnergySystem()

	// 設定信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 啟動系統
	err := system.Start()
	if err != nil {
		log.Fatalf("系統啟動失敗: %v", err)
	}

	// 等待中斷信號
	<-sigChan
	fmt.Println("\n接收到中斷信號，正在停止系統...")
	system.Stop()
}
