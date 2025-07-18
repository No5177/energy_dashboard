package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/goburrow/modbus"
)

// ModbusClient 用於 Modbus 通訊的結構體
type ModbusClient struct {
	IPAddress string
	Port      string
	SlaveID   byte
	handler   *modbus.TCPClientHandler
	client    modbus.Client
}

// NewModbusClient 建立新的 Modbus 客戶端
func NewModbusClient(ipAddress, port string, slaveID byte) *ModbusClient {
	return &ModbusClient{
		IPAddress: ipAddress,
		Port:      port,
		SlaveID:   slaveID,
	}
}

// Connect 連接到 Modbus 裝置
func (mc *ModbusClient) Connect() error {
	// 設定 Modbus TCP 客戶端
	address := fmt.Sprintf("%s:%s", mc.IPAddress, mc.Port)
	mc.handler = modbus.NewTCPClientHandler(address)
	mc.handler.Timeout = 10 * time.Second
	mc.handler.IdleTimeout = 60 * time.Second
	mc.handler.SlaveId = mc.SlaveID // 設定通訊位址
	mc.handler.Logger = log.New(os.Stdout, "modbus: ", log.LstdFlags)

	// 連接到電表
	err := mc.handler.Connect()
	if err != nil {
		return fmt.Errorf("無法連接到電表 %s: %v", address, err)
	}

	mc.client = modbus.NewClient(mc.handler)
	return nil
}

// Close 關閉連接
func (mc *ModbusClient) Close() {
	if mc.handler != nil {
		mc.handler.Close()
	}
}

// ReadMeterData 讀取電表資料
func (mc *ModbusClient) ReadMeterData() error {
	log.Printf("開始讀取電表資料 (IP: %s, 通訊位址: %d)", mc.IPAddress, mc.SlaveID)

	// 讀取相電壓平均值 (Vavg)
	// Modbus Address (Dec): 400263, Device Address (Hex): 0106
	vavgAddress := uint16(0x0106) // 相電壓平均值的 Modbus 寄存器地址
	resultsVavg, err := mc.client.ReadHoldingRegisters(vavgAddress, 2)
	if err != nil {
		log.Printf("讀取相電壓平均值失敗: %v", err)
	} else {
		// 將 4 個位元組轉換為浮點數 (IEEE754 標準)
		// 假設是 ABCD (Big-Endian) 順序
		if len(resultsVavg) >= 4 {
			vavgFloat := binary.BigEndian.Uint32(resultsVavg)
			voltage := math.Float32frombits(vavgFloat)
			fmt.Printf("相電壓平均值 (Vavg): %.3f V\n", voltage)
		}
	}

	// 讀取三相平均電流 (Iavg)
	// Modbus Address (Dec): 400295, Device Address (Hex): 0126
	iavgAddress := uint16(0x0126) // 三相平均電流的 Modbus 寄存器地址
	resultsIavg, err := mc.client.ReadHoldingRegisters(iavgAddress, 2)
	if err != nil {
		log.Printf("讀取三相平均電流失敗: %v", err)
	} else {
		if len(resultsIavg) >= 4 {
			iavgFloat := binary.BigEndian.Uint32(resultsIavg)
			current := math.Float32frombits(iavgFloat)
			fmt.Printf("三相平均電流 (Iavg): %.3f A\n", current)
		}
	}

	return nil
}

// TestModbusConnection 測試 Modbus 連接的函數
func TestModbusConnection() {
	// 根據提供的電表資訊建立客戶端
	// IP: 192.168.1.9, 通訊位址: 2
	client := NewModbusClient("192.168.1.9", "502", 2)

	// 連接到電表
	err := client.Connect()
	if err != nil {
		log.Fatalf("連接失敗: %v", err)
	}
	defer client.Close()

	log.Println("成功連接到電表")

	// 讀取電表資料
	err = client.ReadMeterData()
	if err != nil {
		log.Printf("讀取資料失敗: %v", err)
	}
}

// 如果這個檔案被直接執行，則運行測試
func init() {
	// 可以在這裡放置初始化代碼
}
