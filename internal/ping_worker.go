package internal

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Data ping yang dikirim ke queue
type PingRequest struct {
	UUID     string
	Status   string // "success" / "fail"
	Global   GlobalConfig
	Duration time.Duration
}

// Channel untuk antrian ping
var pingQueue = make(chan PingRequest, 100)

// Inisialisasi worker ping
func StartPingWorker() {
	go func() {
		for ping := range pingQueue {
			sendPing(ping)
			time.Sleep(5000 * time.Millisecond) // delay antar ping
		}
	}()
}

// Fungsi dipanggil dari runTaskOnce
func EnqueuePing(uuid, status string, global GlobalConfig, dur time.Duration) {
	pingQueue <- PingRequest{
		UUID:     uuid,
		Status:   status,
		Global:   global,
		Duration: dur,
	}
}

func sendPing(p PingRequest) {
	url := fmt.Sprintf("%s/ping/%s", p.Global.BaseURL, p.UUID)
	if p.Status == "fail" {
		url += "/fail"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("❌ Gagal buat request: %v", err)
		return
	}

	req.Header.Set("User-Agent", "d-agent-healthchecks")
	req.Header.Set("X-Api-Key", p.Global.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Gagal kirim ping ke %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("📡 Ping terkirim ke %s [%s], status: %d", url, p.Status, resp.StatusCode)
}
