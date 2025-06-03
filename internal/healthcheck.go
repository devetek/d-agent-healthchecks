package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type CreateCheckPayload struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Tags    string `json:"tags"`
	Timeout int    `json:"timeout"`
	Grace   int    `json:"grace"`
}

// EnsureCheckExists returns UUID, whether from static UUID, or by querying API
func EnsureCheckExists(task Task, global GlobalConfig, hostname string) (string, error) {
	// Jika sudah punya UUID, update check dulu
	if task.UUID != "" {
		log.Printf("🔄 UUID ditemukan di config, update: %s", task.UUID)

		go scheduleCheckUpdate(task, global, task.UUID) // background update tiap 10 menit

		err := updateCheck(task, global, task.UUID) // update sekali di awal
		if err != nil {
			log.Printf("⚠️ Gagal update check %s: %v", task.Name, err)
		}

		return task.UUID, nil
	}

	// Kalau tidak ada UUID, lookup via slug
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
		return CreateCheck(task, global, hostname)
	}

	var resp struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}

	go scheduleCheckUpdate(task, global, resp.UUID)
	return resp.UUID, nil
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

func updateCheck(task Task, global GlobalConfig, uuid string) error {
	url := fmt.Sprintf("%s/api/v3/checks/%s", global.BaseURL, uuid)

	body := CreateCheckPayload{
		Name:    task.Name,
		Slug:    task.Slug,
		Tags:    strings.Join(task.Tags, " "),
		Timeout: task.Interval,
		Grace:   task.Grace,
	}

	jsonData, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", global.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("❌ gagal update check: status %d", res.StatusCode)
	}
	return nil
}

func scheduleCheckUpdate(task Task, global GlobalConfig, uuid string) {
	ticker := time.NewTicker(10 * time.Minute)
	for {
		err := updateCheck(task, global, uuid)
		if err != nil {
			log.Printf("⚠️ Gagal update check background %s: %v", task.Slug, err)
		}
		<-ticker.C
	}
}
