package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type CreateCheckPayload struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Tags    string `json:"tags"`
	Timeout int    `json:"timeout"`
	Grace   int    `json:"grace"`
}

func EnsureCheckExists(task Task, global GlobalConfig, hostname string) (string, error) {
	// 1. Coba gunakan UUID dari task
	if task.UUID != "" {
		log.Printf("🧾 UUID ditemukan di config: %s", task.UUID)
		if checkExists(global, task.UUID) {
			log.Printf("✅ UUID %s valid, digunakan langsung", task.UUID)
			return task.UUID, nil
		}
		log.Printf("⚠️ UUID %s tidak valid, lanjut cari cache...", task.UUID)
	}

	// 2. Coba gunakan cache lokal berdasarkan slug
	cacheFile := filepath.Join("/etc/d-agent-healthchecks/.check_id", fmt.Sprintf("%s.txt", task.Slug))
	if data, err := os.ReadFile(cacheFile); err == nil {
		cachedUUID := strings.TrimSpace(string(data))
		log.Printf("📦 Menggunakan UUID dari cache lokal: %s", cachedUUID)
		if checkExists(global, cachedUUID) {
			return cachedUUID, nil
		}
		log.Printf("⚠️ UUID dari cache juga tidak valid, buat check baru...")
	}

	// 3. UUID tidak tersedia atau tidak valid → buat baru
	log.Printf("➕ Membuat check baru: %s", task.Slug)
	uuid, err := CreateCheck(task, global, hostname)
	if err != nil {
		return "", err
	}

	_ = saveUUIDToCache(task.Slug, uuid)
	return uuid, nil
}

func checkExists(global GlobalConfig, uuid string) bool {
	url := fmt.Sprintf("%s/api/v3/checks/%s", global.BaseURL, uuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", global.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ Gagal request UUID %s: %v", uuid, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ UUID %s tidak ditemukan (status %d)", uuid, resp.StatusCode)
		return false
	}
	return true
}

func CreateCheck(task Task, global GlobalConfig, hostname string) (string, error) {
	url := fmt.Sprintf("%s/api/v3/checks/", global.BaseURL)

	timeout := global.DefaultCheckInterval
	grace := global.DefaultGrace
	if task.Interval > 0 {
		timeout = task.Interval
	}
	if task.Grace > 0 {
		grace = task.Grace
	}

	payload := CreateCheckPayload{
		Name:    fmt.Sprintf("[%s] %s", hostname, task.Name),
		Slug:    task.Slug,
		Tags:    strings.Join(task.Tags, " "),
		Timeout: timeout,
		Grace:   grace,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", global.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("❌ gagal create check: status %d, body: %s", resp.StatusCode, body)
	}

	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("❌ gagal parse response UUID: %v", err)
	}

	log.Printf("✅ Check berhasil dibuat: UUID %s", result.UUID)
	return result.UUID, nil
}

func saveUUIDToCache(slug, uuid string) error {
	dir := "/etc/d-agent-healthchecks/.check_id"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.txt", slug))
	return os.WriteFile(path, []byte(uuid), 0600)
}
