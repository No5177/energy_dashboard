package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// MeterData 電表資料結構
type MeterData struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// EnergyDataClient 負責與 LabVIEW TCP 伺服器通訊的客戶端
type EnergyDataClient struct {
	LabviewHost string
	LabviewPort int
	JsonFile    string
	Running     bool
	StopChan    chan bool
}

// NewEnergyDataClient 建立新的資料客戶端
func NewEnergyDataClient() *EnergyDataClient {
	return &EnergyDataClient{
		LabviewHost: "localhost",
		LabviewPort: 8888,
		JsonFile:    "final.json",
		Running:     false,
		StopChan:    make(chan bool),
	}
}

// CalculateChecksum 計算資料的 CheckSum
func (client *EnergyDataClient) CalculateChecksum(data string) string {
	total := 0
	for _, char := range data {
		total += int(char)
	}
	return fmt.Sprintf("%02X", total&0xFF)
}

// CreateCommand 建立命令訊息
func (client *EnergyDataClient) CreateCommand(data string) string {
	// 計算 CheckSum
	checksum := client.CalculateChecksum(data)

	// 計算長度 (資料長度 + CheckSum長度2)
	dataLength := len(data) + 2
	lengthStr := fmt.Sprintf("%06d", dataLength)

	// 組合完整命令
	command := lengthStr + data + checksum

	return command
}

// ParseResponse 解析伺服器回應 (改進版)
func (client *EnergyDataClient) ParseResponse(rawData string) ([]MeterData, error) {
	// 記錄接收到的原始資料用於除錯
	log.Printf("接收到原始資料 (%d bytes): %q", len(rawData), rawData)

	if len(rawData) < 6 {
		return nil, fmt.Errorf("回應資料太短: %d bytes", len(rawData))
	}

	// 提取長度 (前6個字元)
	lengthStr := rawData[:6]
	log.Printf("長度欄位: '%s'", lengthStr)

	// 檢查長度欄位是否包含非數字字元
	for i, char := range lengthStr {
		if char < '0' || char > '9' {
			return nil, fmt.Errorf("長度欄位位置 %d 包含非數字字元: '%c' (ASCII: %d)", i, char, int(char))
		}
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, fmt.Errorf("長度格式錯誤: %v, 長度欄位: '%s'", err, lengthStr)
	}

	log.Printf("解析出的長度: %d", length)

	if len(rawData) < 6+length {
		return nil, fmt.Errorf("回應資料長度不足: 需要 %d bytes，實際 %d bytes", 6+length, len(rawData))
	}

	// 提取資料部分 (去除長度和最後2位checksum)
	dataPart := rawData[6 : 6+length-2]

	// 提取 checksum (最後2位)
	receivedChecksum := rawData[6+length-2 : 6+length]

	log.Printf("資料部分: %s", dataPart)
	log.Printf("接收的 CheckSum: %s", receivedChecksum)

	// 驗證 checksum
	calculatedChecksum := client.CalculateChecksum(dataPart)
	log.Printf("計算的 CheckSum: %s", calculatedChecksum)

	if strings.ToUpper(receivedChecksum) != strings.ToUpper(calculatedChecksum) {
		return nil, fmt.Errorf("CheckSum 錯誤: 接收=%s, 計算=%s", receivedChecksum, calculatedChecksum)
	}

	// 解析 JSON 資料
	var jsonData []MeterData
	err = json.Unmarshal([]byte(dataPart), &jsonData)
	if err != nil {
		return nil, fmt.Errorf("JSON 解析錯誤: %v, JSON 內容: %s", err, dataPart)
	}

	return jsonData, nil
}

// QueryMeterData 查詢電表資料 (修正版)
func (client *EnergyDataClient) QueryMeterData() bool {
	// 建立連線
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", client.LabviewHost, client.LabviewPort), 5*time.Second)
	if err != nil {
		log.Printf("連線錯誤: %v", err)
		return false
	}
	defer conn.Close()

	// 設定讀取超時
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 建立查詢命令
	command := client.CreateCommand("query")

	// 發送命令 (確保使用 ASCII 編碼)
	_, err = conn.Write([]byte(command))
	if err != nil {
		log.Printf("發送命令錯誤: %v", err)
		return false
	}

	// 接收回應 (修正：使用 bytes 而不是 string)
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("接收回應錯誤: %v", err)
		return false
	}

	// 記錄原始 bytes 用於除錯
	log.Printf("接收到 %d bytes 原始資料", n)

	// 確保資料是 ASCII 編碼
	rawBytes := buffer[:n]

	// 檢查是否有非 ASCII 字元並清理
	cleanBytes := make([]byte, 0, len(rawBytes))
	for _, b := range rawBytes {
		if b >= 32 && b <= 126 || b >= 48 && b <= 57 { // 可顯示 ASCII 或數字
			cleanBytes = append(cleanBytes, b)
		}
	}

	response := string(cleanBytes)
	log.Printf("清理後的資料: %s", response)

	// 解析回應
	jsonData, err := client.ParseResponse(response)
	if err != nil {
		log.Printf("解析回應錯誤: %v", err)
		return false
	}

	// 更新 JSON 檔案
	err = client.UpdateJsonFile(jsonData)
	if err != nil {
		log.Printf("更新 JSON 檔案錯誤: %v", err)
		return false
	}

	log.Printf("成功更新電表資料: %s", time.Now().Format("15:04:05"))
	return true
}

// UpdateJsonFile 更新 JSON 檔案
func (client *EnergyDataClient) UpdateJsonFile(data []MeterData) error {
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("JSON 編碼錯誤: %v", err)
	}

	err = ioutil.WriteFile(client.JsonFile, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("寫入檔案錯誤: %v", err)
	}

	return nil
}

// StartPeriodicQuery 開始定期查詢
func (client *EnergyDataClient) StartPeriodicQuery() {
	client.Running = true
	log.Println("開始每 5 秒查詢電表資料...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for client.Running {
		select {
		case <-ticker.C:
			client.QueryMeterData()
		case <-client.StopChan:
			return
		}
	}
}

// Stop 停止查詢
func (client *EnergyDataClient) Stop() {
	client.Running = false
	close(client.StopChan)
}

// EnergyWebServer 能源看板 Web 伺服器
type EnergyWebServer struct {
	WebPort    int
	DataClient *EnergyDataClient
	Server     *http.Server
	Running    bool
}

// NewEnergyWebServer 建立新的 Web 伺服器
func NewEnergyWebServer() *EnergyWebServer {
	return &EnergyWebServer{
		WebPort:    5177,
		DataClient: NewEnergyDataClient(),
		Running:    false,
	}
}

// CustomHandler 自定義 HTTP 請求處理器
func (server *EnergyWebServer) CustomHandler(w http.ResponseWriter, r *http.Request) {
	// 添加 CORS 標頭
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 處理 OPTIONS 請求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 記錄請求
	log.Printf("[%s] %s %s", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	// 使用標準檔案伺服器處理
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}

// StartWebServer 啟動 Web 伺服器
func (server *EnergyWebServer) StartWebServer() error {
	// 確保在正確的目錄中
	if _, err := os.Stat("energy_dashboard.html"); os.IsNotExist(err) {
		return fmt.Errorf("錯誤: 找不到 energy_dashboard.html 檔案")
	}

	// 建立 HTTP 伺服器
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.CustomHandler)

	server.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", server.WebPort),
		Handler: mux,
	}

	log.Printf("Web 伺服器啟動於 http://localhost:%d", server.WebPort)

	// 在新 goroutine 中啟動伺服器
	go func() {
		if err := server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Web 伺服器錯誤: %v", err)
		}
	}()

	return nil
}

// StartDataClient 啟動資料客戶端
func (server *EnergyWebServer) StartDataClient() {
	// 在新 goroutine 中啟動定期查詢
	go server.DataClient.StartPeriodicQuery()
}

// OpenDashboard 開啟能源看板網頁
func (server *EnergyWebServer) OpenDashboard() {
	url := fmt.Sprintf("http://localhost:%d/energy_dashboard.html", server.WebPort)

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
		log.Printf("開啟網頁錯誤: %v", err)
	} else {
		log.Printf("已開啟能源看板: %s", url)
	}
}

// Start 啟動完整系統
func (server *EnergyWebServer) Start() error {
	server.Running = true

	fmt.Println("==================================================")
	fmt.Println("能源看板系統啟動中...")
	fmt.Println("==================================================")

	// 1. 啟動 Web 伺服器
	err := server.StartWebServer()
	if err != nil {
		return err
	}

	// 2. 等待一下讓伺服器完全啟動
	time.Sleep(1 * time.Second)

	// 3. 啟動資料客戶端
	server.StartDataClient()

	// 4. 等待一下讓第一次查詢完成
	time.Sleep(2 * time.Second)

	// 5. 開啟網頁
	server.OpenDashboard()

	fmt.Println("==================================================")
	fmt.Println("系統啟動完成！")
	fmt.Printf("Web 介面: http://localhost:%d/energy_dashboard.html\n", server.WebPort)
	fmt.Println("每 5 秒自動更新電表資料")
	fmt.Println("按 Ctrl+C 停止系統")
	fmt.Println("==================================================")

	return nil
}

// Stop 停止系統
func (server *EnergyWebServer) Stop() {
	server.Running = false

	// 停止資料客戶端
	server.DataClient.Stop()

	// 停止 Web 伺服器
	if server.Server != nil {
		server.Server.Close()
	}

	log.Println("系統已停止")
}

func main() {
	server := NewEnergyWebServer()

	// 設定信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 啟動系統
	err := server.Start()
	if err != nil {
		log.Fatalf("系統啟動失敗: %v", err)
	}

	// 等待中斷信號
	<-sigChan
	fmt.Println("\n接收到中斷信號，正在停止系統...")
	server.Stop()
}
