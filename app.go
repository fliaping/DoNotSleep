package main

import (
	"encoding/hex"
	wapi "github.com/iamacarpet/go-win64api"
	"log"
	"net"
	"strings"
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
	log.Print("service start")

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
		log.Printf("received host %s sent this: %s\n", addr, packageHex)

		isWol := isWolPackage(packageHex)
		if isWol {
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
