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

// 台達電表參數定義 (根據通訊格式表格)
type MeterParameter struct {
	Name    string
	Address uint16
	Unit    string
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

func readMeterData() {
	fmt.Println("==================================================")
	fmt.Println("台達電表 Modbus TCP 通訊測試")
	fmt.Println("==================================================")
	fmt.Printf("電表 IP: 192.168.1.9\n")
	fmt.Printf("子網路遮罩: 255.255.255.0\n")
	fmt.Printf("閘道: 192.168.1.1\n")
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
		log.Fatalf("❌ 無法連接到電表 192.168.1.9: %v\n連接失敗可能原因:\n1. 電表未開機或網路未連接\n2. IP 地址設定錯誤\n3. 防火牆阻擋連線\n4. Modbus TCP 服務未啟用", err)
	}
	defer handler.Close()

	fmt.Println("✅ 成功連接到電表")

	client := modbus.NewClient(handler)

	// 連續讀取電表資料
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- 第 %d 次讀取 (%s) ---\n", i+1, time.Now().Format("15:04:05"))

		// 讀取所有定義的參數
		for _, param := range meterParameters {
			fmt.Printf("讀取 %s (暫存器地址: 0x%04X)...\n", param.Name, param.Address)

			results, err := client.ReadHoldingRegisters(param.Address, 2)
			if err != nil {
				log.Printf("❌ 讀取 %s 失敗: %v", param.Name, err)
				continue
			}

			if len(results) >= 4 {
				// 嘗試不同的字節順序解析

				// Big-Endian 解析 (ABCD)
				valueFloat := binary.BigEndian.Uint32(results)
				value := math.Float32frombits(valueFloat)

				// Word Swap 解析 (CDAB)
				swapped := []byte{results[2], results[3], results[0], results[1]}
				valueFloatSwap := binary.BigEndian.Uint32(swapped)
				valueSwap := math.Float32frombits(valueFloatSwap)

				// 直接作為 16 位整數解析
				value16 := binary.BigEndian.Uint16(results[0:2])

				fmt.Printf("✅ %s:\n", param.Name)
				fmt.Printf("   Big-Endian (32位浮點): %.3f %s\n", value, param.Unit)
				fmt.Printf("   Word-Swap (32位浮點): %.3f %s\n", valueSwap, param.Unit)
				fmt.Printf("   16位整數: %d %s\n", value16, param.Unit)

				// 顯示原始資料
				fmt.Printf("   原始暫存器: [0x%04X, 0x%04X]\n",
					binary.BigEndian.Uint16(results[0:2]),
					binary.BigEndian.Uint16(results[2:4]))
				fmt.Printf("   原始位元組: %02X %02X %02X %02X\n",
					results[0], results[1], results[2], results[3])
			} else {
				fmt.Printf("❌ 接收到的資料長度不足: %d bytes\n", len(results))
			}
			fmt.Println()
		}

		// 等待 5 秒後進行下一次讀取
		if i < 2 {
			fmt.Println("等待 5 秒...")
			time.Sleep(5 * time.Second)
		}
	}

	fmt.Println("\n==================================================")
	fmt.Println("測試完成")
	fmt.Println("==================================================")
	fmt.Println("注意事項:")
	fmt.Println("1. 如果數值看起來不正確，可能需要調整字節順序")
	fmt.Println("2. 請參考電表手冊確認正確的資料格式")
	fmt.Println("3. 某些參數可能需要不同的資料類型 (整數/浮點數)")
	fmt.Println("4. 根據實際測試結果選擇正確的解析方式")
}

func main() {
	readMeterData()
}
