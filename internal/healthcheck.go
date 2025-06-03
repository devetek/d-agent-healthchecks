package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// EnsureCheckExists memastikan UUID check tersedia (dari config, cache lokal, lookup, atau create baru)
func EnsureCheckExists(task Task, global GlobalConfig, hostname string) (string, error) {
	// Jika UUID ada → verifikasi dulu ke server
	if task.UUID != "" {
		log.Printf("🧾 UUID ditemukan di config: %s", task.UUID)
		url := fmt.Sprintf("%s/api/v3/checks/%s", global.BaseURL, task.UUID)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("X-Api-Key", global.APIKey)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("❌ Gagal verifikasi UUID %s: %v", task.UUID, err)
		}
		defer res.Body.Close()

		if res.StatusCode == 200 {
			log.Printf("✅ Check dengan UUID %s valid, digunakan langsung", task.UUID)
			return task.UUID, nil
		}

		log.Printf("⚠️ UUID %s tidak ditemukan, buat check baru...", task.UUID)
	}

	// Jika tidak ada UUID (atau tidak valid) → cek via slug
	checkIDFile := filepath.Join(".check_id", fmt.Sprintf("%s.txt", task.Slug))
	if data, err := os.ReadFile(checkIDFile); err == nil {
		cachedUUID := strings.TrimSpace(string(data))
		log.Printf("📦 Menggunakan UUID dari cache lokal: %s", cachedUUID)
		return cachedUUID, nil
	}

	// Cek slug ke server
	url := fmt.Sprintf("%s/api/v3/checks/%s", global.BaseURL, task.Slug)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", global.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.Printf("➕ Check belum ada: %s → buat baru", task.Slug)
		uuid, err := CreateCheck(task, global, hostname)
		if err == nil {
			os.MkdirAll(".check_id", 0755)
			_ = os.WriteFile(checkIDFile, []byte(uuid), 0644)
		}
		return uuid, err
	}

	var resp struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}

	log.Printf("📡 UUID ditemukan dari slug: %s", resp.UUID)
	os.MkdirAll(".check_id", 0755)
	_ = os.WriteFile(checkIDFile, []byte(resp.UUID), 0644)
	return resp.UUID, nil
}

// CreateCheck membuat check baru di server Healthchecks.io
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

	body := CreateCheckPayload{
		Name:    fmt.Sprintf("[%s] %s", hostname, task.Name),
		Slug:    task.Slug,
		Tags:    strings.Join(task.Tags, " "),
		Timeout: timeout,
		Grace:   grace,
	}

	jsonData, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", global.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return "", fmt.Errorf("❌ gagal create check: status %d", res.StatusCode)
	}

	var resp struct {
		UUID string `json:"uuid"`
	}
	err = json.NewDecoder(res.Body).Decode(&resp)
	return resp.UUID, err
}

// saveUUIDToCache menyimpan UUID ke file lokal
func saveUUIDToCache(slug, uuid string) error {
	dir := "/etc/d-agent-healthchecks/.check_id"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := fmt.Sprintf("%s/%s.txt", dir, slug)
	return os.WriteFile(path, []byte(uuid), 0600)
}
