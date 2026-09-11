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
	"github.com/local/gb28181-device/internal/ui"
)

type multiWriter struct {
	dev *device.Device
	std io.Writer
}

func (m multiWriter) Write(p []byte) (int, error) {
	n, err := m.std.Write(p)
	line := string(p)
	// 去掉末尾换行后再存
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	if line != "" {
		m.dev.AppendLog(line)
	}
	return n, err
}

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	d := device.New(cfg)
	log.SetOutput(multiWriter{dev: d, std: os.Stderr})

	log.Printf("GB28181 device simulator starting")
	log.Printf("  device id : %s", cfg.Device.ID)
	log.Printf("  sip server: %s:%d (%s)", cfg.SIP.ServerIP, cfg.SIP.ServerPort, cfg.SIP.Transport)
	log.Printf("  sip local : %s:%d", cfg.SIP.LocalIP, cfg.SIP.LocalPort)
	log.Printf("  channels  : %d", len(cfg.Device.Channels))
	log.Printf("  media     : %s %dx%d@%dfps", cfg.Media.Source, cfg.Media.Width, cfg.Media.Height, cfg.Media.FPS)

	if err := d.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start device failed: %v\n", err)
		os.Exit(1)
	}

	if cfg.UI.Enabled {
		srv := ui.New(d, cfg)
		go func() {
			if err := srv.ListenAndServe(cfg.UI.Listen); err != nil {
				log.Printf("[ui] server error: %v", err)
			}
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("shutting down...")
	d.Stop()
	log.Printf("bye")
}
