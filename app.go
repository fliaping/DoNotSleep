package main

import (
	"encoding/hex"
	wapi "github.com/iamacarpet/go-win64api"
	"log"
	"net"
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

func main() {
	lastRecWolTime := time.Now()
	isCanSleep := true
	noWolToSleepTime := time.Second * 120
	checkDuration := time.Second * 5

	go func() {
		t := time.NewTicker(checkDuration)
		defer t.Stop()

		for {
			<-t.C
			//log.Println("start check still receive wol ...")
			if time.Now().Sub(lastRecWolTime) > noWolToSleepTime && !isCanSleep {
				log.Printf("no receive wol in %s, reset can be sleep\n", noWolToSleepTime)
				res, err := wapi.SetThreadExecutionState(wapi.ES_CONTINUOUS)
				if res != 0 {
					log.Printf("[ERROR] reset to sleep fail,%s\n", err.Error())
				} else {
					isCanSleep = true
				}
				log.Printf("reset to sleep, result:%d\n", res)
			}
		}
	}()

	pc, err := net.ListenPacket("udp4", ":9")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		packageHex := hex.EncodeToString(buf[:n])
		log.Printf("%s sent this: %s\n", addr, packageHex)

		isWol := isWolPackage(packageHex)
		if isWol {
			lastRecWolTime = time.Now()

			if isCanSleep {
				res, err := wapi.SetThreadExecutionState(wapi.ES_CONTINUOUS | wapi.ES_SYSTEM_REQUIRED)
				if res != 0 {
					log.Printf("[ERROR] set not sleep fail,%s\n", err.Error())
				} else {
					isCanSleep = false
				}
				log.Printf("set not sleep, result:%d\n", res)
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
		log.Printf("wolStart:%s, packageStart:%s\n", wolStart, hexStr[:len(wolStart)])
		if strings.ToLower(hexStr[:len(wolStart)]) == strings.ToLower(wolStart) {
			return true
		}
	}
	return false
}
