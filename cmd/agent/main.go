package main

import (
	"d-agent-healthchecks/internal"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	clearCheckIDCache()

	configPath := flag.String("config", "configs/agent.yml", "Path to config file")
	flag.Parse()

	fmt.Printf("📂 Using config: %s\n", *configPath)

	config, err := internal.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	hostname := internal.GetHostname()

	for _, task := range config.Tasks {
		t := task
		go func() {
			checkID, err := internal.EnsureCheckExists(t, config.Global, hostname)
			if err != nil {
				fmt.Printf("⚠️ Gagal sync check %s: %v\n", t.Name, err)
				return
			}
			// 🚀 Hanya jalankan loop di sini, tidak perlu retry ulang ulang
			internal.RunTaskLoop(t, checkID, config.Global)
		}()
	}

	select {} // block forever
}

func clearCheckIDCache() {
	cacheDir := ".check_id"
	err := os.RemoveAll(cacheDir)
	if err != nil {
		log.Printf("⚠️ Gagal hapus cache folder %s: %v", cacheDir, err)
	} else {
		log.Printf("🧹 Cache %s dibersihkan", cacheDir)
	}
}
