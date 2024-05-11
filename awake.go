package main

import (
	"encoding/hex"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	// wapi "github.com/iamacarpet/go-win64api"
	// "github.com/robfig/cron"
)

var (
	kernel32                           = syscall.NewLazyDLL("kernel32.dll")
	procSetThreadExecutionState        = kernel32.NewProc("SetThreadExecutionState")
	ES_AWAYMODE_REQUIRED        uint32 = 0x00000040
	ES_CONTINUOUS               uint32 = 0x80000000
	ES_DISPLAY_REQUIRED         uint32 = 0x00000002
	ES_SYSTEM_REQUIRED          uint32 = 0x00000001
)

var latestWolTime time.Time
var mutex sync.Mutex

// SetThreadExecutionState 设置线程执行状态以阻止系统休眠
func SetThreadExecutionState(esFlags uint32) error {
	ret, _, err := procSetThreadExecutionState.Call(uintptr(esFlags))
	if ret == 0 {
		return err
	}
	return nil
}

// 获取本机的物理地址
func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
func mrdConnected() bool {
	out, err := exec.Command("powershell", "[void][System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms');[System.Windows.Forms.SystemInformation]::TerminalServerSession").Output()
	if err != nil {
		log.Fatal(err)
	}
	var output string = string(out)
	return strings.HasPrefix(output, "True")
}

func updateLatestWolTime() {
	mutex.Lock()
	latestWolTime = time.Now()
	mutex.Unlock()
	log.Printf("received wol request,time:", latestWolTime)
}

func checkTimeout() {
	for range time.Tick(5 * time.Second) {
		mrdCon := mrdConnected()

		mutex.Lock()
		timeout := time.Since(latestWolTime) > 3*time.Minute //3分钟没收到WOL即超时，可以休眠
		mutex.Unlock()

		if !mrdCon && timeout {
			log.Printf("go sleep, mrdConnected%s,timeout:%s\n", mrdCon, timeout)

			//rundll32.exe powrprof.dll,SetSuspendState 0,1,0
			// c := exec.Command("powershell", "rundll32.exe powrprof.dll,SetSuspendState 0,1,0")
			// if err := c.Run(); err != nil {
			// 	fmt.Println("Error: ", err)
			// }
			// 阻止系统休眠和关闭显示器
			err := SetThreadExecutionState(ES_CONTINUOUS)
			if err != nil {
				log.Printf("SetThreadExecutionState to resume sleep FAILED!!!!:", err)
			} else {
				log.Printf("SetThreadExecutionState to resume sleep", err)
			}
		} else {
			err := SetThreadExecutionState(ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_AWAYMODE_REQUIRED)
			if err != nil {
				log.Printf("SetThreadExecutionState to disable sleep FAILED!!!!:", err)
			} else {
				log.Printf("SetThreadExecutionState success, disable sleep")
			}
		}

	}
}
func start(stopChan <-chan struct{}) {
	log.Print("service started")
	// 创建一个定时器，每隔5秒触发一次
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // 程序结束时停止定时器

	// 设置初始值
	updateLatestWolTime()
	// 开启定时任务
	go checkTimeout()

	wolPackage, err := net.ListenPacket("udp4", ":9")
	if err != nil {
		panic(err)
	}
	defer func(pc net.PacketConn) {
		err := pc.Close()
		if err != nil {
			log.Printf("[ERROR] ListenPacket error,%s\n", err.Error())
		}
	}(wolPackage)

	buf := make([]byte, 1024)

	for {
		select {
		case <-stopChan: // 接收到停止信号
			log.Print("service stop")
			return // 退出函数，从而结束goroutine的执行
		default:
			// 从udp循环读取数据, 循环执行
			n, addr, err := wolPackage.ReadFrom(buf)
			if err != nil {
				panic(err)
			}
			packageHex := hex.EncodeToString(buf[:n])
			log.Printf("received host %s sent this: %s\n", addr, packageHex)

			isWol := isWolPackage(packageHex)
			if isWol {
				updateLatestWolTime()
			}
		}
	}
}

// 是否是本机的wol的包
func isWolPackage(hexStr string) bool {
	as, err := getMacAddr()
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range as {
		macAddr := strings.ReplaceAll(a, ":", "")
		wolStart := "ffffffffffff" + macAddr
		// log.Printf("wolStart:%s, packageStart:%s\n", wolStart, hexStr[:len(wolStart)])
		if strings.ToLower(hexStr[:len(wolStart)]) == strings.ToLower(wolStart) {
			return true
		}
	}
	return false
}
