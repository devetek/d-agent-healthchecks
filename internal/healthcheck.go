package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	// 1. Cek dari cache lokal
	if uuid, err := getUUIDFromCache(task.Slug); err == nil && uuid != "" {
		log.Printf("📁 UUID ditemukan dari cache: %s", uuid)
		return uuid, nil
	}

	// 2. Cek dari config (tidak update, hanya gunakan untuk ping)
	if task.UUID != "" {
		log.Printf("🧾 UUID ditemukan di config: %s (tidak di-update)", task.UUID)
		_ = saveUUIDToCache(task.Slug, task.UUID)
		return task.UUID, nil
	}

	// 3. Lookup dari slug
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
			_ = saveUUIDToCache(task.Slug, uuid)
		}
		return uuid, err
	}

	var resp struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}

	_ = saveUUIDToCache(task.Slug, resp.UUID)
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

// getUUIDFromCache membaca UUID dari file lokal
func getUUIDFromCache(slug string) (string, error) {
	path := fmt.Sprintf("/etc/d-agent-healthchecks/.check_id/%s.txt", slug)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
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
