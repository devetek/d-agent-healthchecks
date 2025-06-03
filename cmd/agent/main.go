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
	// Tambahkan flag config
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
			log.Printf("🕒 Inisialisasi task: %s", t.Name)
			checkID, err := internal.EnsureCheckExists(t, config.Global, hostname)
			if err != nil {
				log.Printf("⚠️ Gagal sync check %s: %v", t.Name, err)
				return
			}
			internal.RunTaskLoop(t, checkID, config.Global)
		}()
	}

	select {} // blok selamanya
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
