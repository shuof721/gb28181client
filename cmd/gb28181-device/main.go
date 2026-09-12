package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
	"github.com/local/gb28181-device/internal/storage"
	"github.com/local/gb28181-device/internal/ui"
)

type multiWriter struct {
	mgr *device.Manager
	std io.Writer
}

func (m multiWriter) Write(p []byte) (int, error) {
	n, err := m.std.Write(p)
	line := string(p)
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	if line != "" {
		m.mgr.AppendLog(line)
	}
	return n, err
}

func main() {
	var cfgPath string
	var dataDir string
	var uiListen string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "初始/回退配置文件路径")
	flag.StringVar(&cfgPath, "c", "configs/config.yaml", "配置文件路径 (简写)")
	flag.StringVar(&dataDir, "data", "data/devices", "设备数据 JSON 存储目录")
	flag.StringVar(&uiListen, "listen", "", "Web 控制台监听地址 (默认读取 config.yaml 或 127.0.0.1:7080)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// 1. 初始化持久化存储
	store := storage.New(dataDir)
	if err := store.EnsureDir(); err != nil {
		fmt.Fprintf(os.Stderr, "create data dir failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 自动迁移现有 config.yaml（首次启动若无 json 设备，自动导入为第一台设备）
	if imported, err := store.ImportLegacyYAML(cfgPath); err != nil {
		log.Printf("[init] import legacy config failed: %v", err)
	} else if imported != nil {
		log.Printf("[init] imported default device %s from %s", imported.Device.ID, cfgPath)
	}

	// 3. 创建多设备生命周期管理器
	mgr := device.NewManager(store)
	log.SetOutput(multiWriter{mgr: mgr, std: os.Stderr})

	log.Printf("GB28181 multi-device simulator platform starting...")
	log.Printf("  storage dir: %s", dataDir)

	if err := mgr.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "initialize devices failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Web 控制台
	if uiListen == "" {
		if cfg, err := config.Load(cfgPath); err == nil && cfg.UI.Listen != "" {
			uiListen = cfg.UI.Listen
		} else {
			uiListen = "127.0.0.1:7080"
		}
	}

	srv := ui.New(mgr)
	go func() {
		if err := srv.ListenAndServe(uiListen); err != nil {
			log.Printf("[ui] server error: %v", err)
		}
	}()

	// 5. 监听退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("shutting down all devices...")
	mgr.StopAll()
	log.Printf("bye")
}
