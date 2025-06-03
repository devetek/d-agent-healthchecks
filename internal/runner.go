package internal

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

func RunTaskLoop(task Task, checkUUID string, global GlobalConfig) {
	interval := time.Duration(task.Interval)
	if interval == 0 {
		interval = time.Duration(global.DefaultCheckInterval)
	}
	ticker := time.NewTicker(interval * time.Second)

	for {
		runTaskOnce(task, checkUUID, global)
		<-ticker.C
	}
}

func runTaskOnce(task Task, uuid string, global GlobalConfig) {
	cmd := exec.Command("bash", "-c", task.Shell)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ %s gagal: %v\n", task.Name, err)
		fmt.Println(string(output))
		http.Get(fmt.Sprintf("%s/ping/%s/fail", global.BaseURL, uuid))
		return
	}
	fmt.Printf("✅ %s OK\n", task.Name)
	http.Get(fmt.Sprintf("%s/ping/%s", global.BaseURL, uuid))
}
