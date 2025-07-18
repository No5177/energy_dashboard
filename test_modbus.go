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

func testModbusMain() {
	fmt.Println("==================================================")
	fmt.Println("台達電表 Modbus TCP 通訊測試")
	fmt.Println("==================================================")
	fmt.Printf("電表 IP: 192.168.1.9\n")
	fmt.Printf("通訊位址: 2\n")
	fmt.Printf("連接埠: 502\n")
	fmt.Println("==================================================")

	// 建立 Modbus TCP 客戶端
	handler := modbus.NewTCPClientHandler("192.168.1.9:502")
	handler.Timeout = 10 * time.Second
	handler.IdleTimeout = 60 * time.Second
	handler.SlaveId = 2 // 設定通訊位址為 2
	handler.Logger = log.New(os.Stdout, "modbus: ", log.LstdFlags)

	// 連接到電表
	fmt.Println("正在連接到電表...")
	err := handler.Connect()
	if err != nil {
		log.Fatalf("❌ 無法連接到電表 192.168.1.9: %v", err)
	}
	defer handler.Close()

	fmt.Println("✅ 成功連接到電表")

	client := modbus.NewClient(handler)

	// 連續讀取電表資料
	for i := 0; i < 5; i++ {
		fmt.Printf("\n--- 第 %d 次讀取 ---\n", i+1)

		// 讀取相電壓平均值 (Vavg)
		// Modbus Address (Dec): 400263, Device Address (Hex): 0106
		vavgAddress := uint16(0x0106)
		fmt.Printf("讀取相電壓平均值 (地址: 0x%04X)...\n", vavgAddress)

		resultsVavg, err := client.ReadHoldingRegisters(vavgAddress, 2)
		if err != nil {
			log.Printf("❌ 讀取相電壓平均值失敗: %v", err)
		} else {
			if len(resultsVavg) >= 4 {
				// 嘗試不同的字節順序解析
				vavgFloat := binary.BigEndian.Uint32(resultsVavg)
				voltage := math.Float32frombits(vavgFloat)
				fmt.Printf("✅ 相電壓平均值 (Big-Endian): %.3f V\n", voltage)

				// 嘗試 Little-Endian 解析
				vavgFloatLE := binary.LittleEndian.Uint32(resultsVavg)
				voltageLE := math.Float32frombits(vavgFloatLE)
				fmt.Printf("   相電壓平均值 (Little-Endian): %.3f V\n", voltageLE)

				// 顯示原始資料
				fmt.Printf("   原始資料: %02X %02X %02X %02X\n",
					resultsVavg[0], resultsVavg[1], resultsVavg[2], resultsVavg[3])
			} else {
				fmt.Printf("❌ 接收到的資料長度不足: %d bytes\n", len(resultsVavg))
			}
		}

		// 讀取三相平均電流 (Iavg)
		// Modbus Address (Dec): 400295, Device Address (Hex): 0126
		iavgAddress := uint16(0x0126)
		fmt.Printf("讀取三相平均電流 (地址: 0x%04X)...\n", iavgAddress)

		resultsIavg, err := client.ReadHoldingRegisters(iavgAddress, 2)
		if err != nil {
			log.Printf("❌ 讀取三相平均電流失敗: %v", err)
		} else {
			if len(resultsIavg) >= 4 {
				// 嘗試不同的字節順序解析
				iavgFloat := binary.BigEndian.Uint32(resultsIavg)
				current := math.Float32frombits(iavgFloat)
				fmt.Printf("✅ 三相平均電流 (Big-Endian): %.3f A\n", current)

				// 嘗試 Little-Endian 解析
				iavgFloatLE := binary.LittleEndian.Uint32(resultsIavg)
				currentLE := math.Float32frombits(iavgFloatLE)
				fmt.Printf("   三相平均電流 (Little-Endian): %.3f A\n", currentLE)

				// 顯示原始資料
				fmt.Printf("   原始資料: %02X %02X %02X %02X\n",
					resultsIavg[0], resultsIavg[1], resultsIavg[2], resultsIavg[3])
			} else {
				fmt.Printf("❌ 接收到的資料長度不足: %d bytes\n", len(resultsIavg))
			}
		}

		// 嘗試讀取更多電表參數
		fmt.Println("嘗試讀取其他電表參數...")

		// 讀取功率因子 (假設地址)
		pfAddress := uint16(0x0150) // 這個地址需要根據實際電表手冊確認
		resultsPF, err := client.ReadHoldingRegisters(pfAddress, 2)
		if err != nil {
			fmt.Printf("讀取功率因子失敗 (地址: 0x%04X): %v\n", pfAddress, err)
		} else {
			if len(resultsPF) >= 4 {
				pfFloat := binary.BigEndian.Uint32(resultsPF)
				pf := math.Float32frombits(pfFloat)
				fmt.Printf("功率因子: %.3f\n", pf)
			}
		}

		// 等待 2 秒後進行下一次讀取
		if i < 4 {
			fmt.Println("等待 2 秒...")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("測試完成")
	fmt.Println("==================================================")
}
