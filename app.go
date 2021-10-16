package main

import (
	"encoding/hex"
	"fmt"
	wapi "github.com/iamacarpet/go-win64api"
	"github.com/robfig/cron"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

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
func main() {
	log.Print("service start")
	var latestWolTime int64 = time.Now().Unix()
	//开启休眠定时
	c := cron.New() //精确到秒
	//定时任务
	spec := "* */1 * * * ?" //cron表达式，每秒一次
	err := c.AddFunc(spec, func() {
		mrdCon := mrdConnected()
		timeout := (time.Now().Unix() - latestWolTime) > 15*60 //15分钟没收到WOL即超时
		if !mrdCon && timeout {
			log.Printf("go sleep, mrdConnected%s,timeout:%s\n", mrdCon, timeout)
			//rundll32.exe powrprof.dll,SetSuspendState 0,1,0
			c := exec.Command("powershell", "rundll32.exe powrprof.dll,SetSuspendState 0,1,0")
			if err := c.Run(); err != nil {
				fmt.Println("Error: ", err)
			}
		}
	})
	if err != nil {
		return
	}
	c.Start()

	pc, err := net.ListenPacket("udp4", ":9")
	if err != nil {
		panic(err)
	}
	defer func(pc net.PacketConn) {
		err := pc.Close()
		if err != nil {
			log.Printf("[ERROR] ListenPacket error,%s\n", err.Error())
		}
	}(pc)

	buf := make([]byte, 1024)

	// 从udp循环读取数据
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		packageHex := hex.EncodeToString(buf[:n])
		log.Printf("received host %s sent this: %s\n", addr, packageHex)

		isWol := isWolPackage(packageHex)
		if isWol {
			latestWolTime = time.Now().Unix()
			res, err := wapi.SetThreadExecutionState(wapi.ES_SYSTEM_REQUIRED)
			if res != 0 {
				log.Printf("[ERROR] reset sleep timer fail,%s\n", err.Error())
			} else {
				log.Printf("reset sleep timer, result:%d\n", res)
			}
		}
	}
}

func isWolPackage(hexStr string) bool {
	as, err := getMacAddr()
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range as {
		macAddr := strings.ReplaceAll(a, ":", "")
		wolStart := "ffffffffffff" + macAddr
		//log.Printf("wolStart:%s, packageStart:%s\n", wolStart, hexStr[:len(wolStart)])
		if strings.ToLower(hexStr[:len(wolStart)]) == strings.ToLower(wolStart) {
			return true
		}
	}
	return false
}
