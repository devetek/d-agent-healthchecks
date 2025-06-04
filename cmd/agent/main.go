package main

import (
	"d-agent-healthchecks/internal"
	"flag"
	"fmt"
	"log"
	"os"
)

var version = "0.1.0"

func main() {
	fmt.Printf("🌀 Agent Version: %s\n", version)
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
		go func(t internal.Task) {
			checkID, err := internal.EnsureCheckExists(t, config.Global, hostname)
			if err != nil {
				fmt.Printf("⚠️ Gagal sync check %s: %v\n", t.Name, err)
				return
			}
			log.Printf("▶️ Starting task loop for: %s", t.Name)
			internal.RunTaskLoop(t, checkID, config.Global)
		}(task) // ✅ penting: lempar task ke parameter func
	}

	select {} // block selamanya
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
