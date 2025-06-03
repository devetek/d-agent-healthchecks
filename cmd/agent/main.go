package main

import (
	"d-agent-healthchecks/internal"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
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
		go func(t internal.Task) {
			for {
				fmt.Printf("🌐 BaseURL = %s\n", config.Global.BaseURL)
				fmt.Printf("🔑 API Key = %s\n", config.Global.APIKey)

				checkID, err := internal.EnsureCheckExists(t, config.Global, hostname)
				if err != nil {
					log.Printf("⚠️ Gagal sync check %s: %v", t.Name, err)
				} else {
					internal.RunTaskLoop(t, checkID, config.Global)
				}
				time.Sleep(60 * time.Second)
			}
		}(task)
	}

	select {} // blok selamanya
}
