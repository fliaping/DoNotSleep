package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger service.Logger

type program struct {
}

var stopChan = make(chan struct{})

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	// Do work here
	go func() {
		start(stopChan)
	}()
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	// Any long running operations should be done in Start() - Stop().
	// 发送停止信号
	log.Printf("Stop start.......")
	close(stopChan)
	log.Printf("Stop end.......")
	return nil
}

func main() {
	// 获取当前执行的可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("获取可执行文件路径失败:", err)
		return
	}

	// 将路径转换为绝对路径
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		fmt.Println("转换为绝对路径失败:", err)
		return
	}
	// 使用filepath.Dir获取程序所在的文件夹路径
	dirPath := filepath.Dir(absPath)
	// 构造日志文件的完整路径
	logFilePath := filepath.Join(dirPath, "do_not_sleep.log")

	fmt.Println("程序日志所在路径:", logFilePath)

	// 将日志输出重定向到文件
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath, // 日志文件路径
		MaxSize:    10,                 // 日志文件最大大小（MB）
		MaxBackups: 2,                  // 保留旧文件的最大个数
		MaxAge:     28,                 // 保留旧文件的最大天数
		Compress:   true,              // 是否压缩/归档旧文件
	})
	svcConfig := &service.Config{
		Name:        "DoNotSleep",
		DisplayName: "DoNotSleep",
		Description: "Do Not Sleep, when received wol package continuously.",
		Arguments:   []string{"run"},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "install":
			err = s.Install()
			if err != nil {
				fmt.Println("Failed to install:", err)
				return
			}
			fmt.Println("Service installed")
		case "start":
			err = s.Start()
			if err != nil {
				fmt.Println("Failed to start:", err)
				return
			}
			fmt.Println("Service started")
		case "stop":
			err = s.Stop()
			if err != nil {
				fmt.Println("Failed to stop:", err)
				return
			}
			fmt.Println("Service stopped")
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				fmt.Println("Failed to uninstall:", err)
				return
			}
			fmt.Println("Service uninstalled")
		case "run":
			err = s.Run()
			if err != nil {
				fmt.Println("Failed to run:", err)
				return
			}
		default:
			fmt.Println("Invalid command. Available commands are: install, start, stop, uninstall, run")
		}
		return
	} else {
		fmt.Println("Invalid command. Available commands are: install, start, stop, uninstall, run")
	}
}
