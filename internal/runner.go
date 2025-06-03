package internal

import (
	"fmt"
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
	start := time.Now()
	cmd := exec.Command("bash", "-c", task.Shell)

	// Output gabungan, supaya EPIPE tidak muncul
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ %s gagal: %v\n", task.Name, err)
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		EnqueuePing(uuid, "fail", global, duration)
		return
	}

	fmt.Printf("✅ %s OK\n", task.Name)
	EnqueuePing(uuid, "success", global, duration)
}
