package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const (
	BASE_URL         = "http://9.135.85.144:8080/api/generate"
	TOTAL_REQUESTS   = 20000
	DELAY_IN_MINUTES = 3
)

type JobData struct {
	JobName  string `json:"jobName"`
	JobType  int    `json:"jobType"`
	CronExpr string `json:"cronExpr"`
	Format   int    `json:"format"`
	Script   string `json:"script"`
	Retries  int    `json:"retries"`
}

func sendRequest(jobName string, wg *sync.WaitGroup) {
	defer wg.Done()

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	client := &http.Client{}

	// Calculate cron expression for DELAY_IN_MINUTES later
	//futureTime := time.Now().Add(time.Duration(DELAY_IN_MINUTES) * time.Minute)
	//cronExpr := fmt.Sprintf("%d %d * * *", futureTime.Minute(), futureTime.Hour())

	data := JobData{
		JobName:  jobName,
		JobType:  1,
		CronExpr: "* * * * * *",
		Format:   1,
		Script:   `print("hello world")`,
		Retries:  3,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Failed to marshal data:", err)
		return
	}

	req, err := http.NewRequest("POST", BASE_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return
	}
	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed for job: %s\n", jobName)
		fmt.Println("Server response:", resp.Status)
	}
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < TOTAL_REQUESTS; i++ {
		jobName := fmt.Sprintf("New-Job%d", i+1)
		wg.Add(1)
		go sendRequest(jobName, &wg)
	}

	wg.Wait()
	fmt.Printf("%d requests sent!\n", TOTAL_REQUESTS)
}
