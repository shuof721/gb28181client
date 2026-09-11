package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/local/gb28181-device/internal/config"
	"github.com/local/gb28181-device/internal/device"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("GB28181 device simulator starting")
	log.Printf("  device id : %s", cfg.Device.ID)
	log.Printf("  sip server: %s:%d (%s)", cfg.SIP.ServerIP, cfg.SIP.ServerPort, cfg.SIP.Transport)
	log.Printf("  sip local : %s:%d", cfg.SIP.LocalIP, cfg.SIP.LocalPort)
	log.Printf("  channels  : %d", len(cfg.Device.Channels))
	log.Printf("  media     : %s %dx%d@%dfps", cfg.Media.Source, cfg.Media.Width, cfg.Media.Height, cfg.Media.FPS)

	d := device.New(cfg)
	if err := d.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start device failed: %v\n", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("shutting down...")
	d.Stop()
	log.Printf("bye")
}
