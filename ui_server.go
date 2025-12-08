package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type Status struct {
	LastTriggered string `json:"lastTriggered"`
	Message       string `json:"message"`
	Logs          []string `json:"logs"`
}

var (
	lastTriggered string
	logs          []string
	mu            sync.Mutex
)

func triggerChaos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	lastTriggered = time.Now().Format(time.RFC1123)
	cmd := exec.Command("./chaosmonkey", "intest")
	out, _ := cmd.CombinedOutput()
	entry := fmt.Sprintf("[%s] Chaos Triggered! Output: %s", lastTriggered, string(out))
	logs = append(logs, entry)
	resp := Status{LastTriggered: lastTriggered, Message: "Chaos Triggered!", Logs: logs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	resp := Status{
		LastTriggered: lastTriggered,
		Message:       "System running",
		Logs:          logs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func submitSuggestion(w http.ResponseWriter, r *http.Request) {
	type Suggestion struct {
		Text string `json:"text"`
	}
	var s Suggestion
	json.NewDecoder(r.Body).Decode(&s)
	entry := fmt.Sprintf("[%s] 💡 Suggestion: %s", time.Now().Format(time.RFC1123), s.Text)
	mu.Lock()
	logs = append(logs, entry)
	mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func main() {
	fs := http.FileServer(http.Dir("ui"))
	http.Handle("/", fs)
	http.HandleFunc("/api/status", getStatus)
	http.HandleFunc("/api/trigger", triggerChaos)
	http.HandleFunc("/api/suggest", submitSuggestion)

	fmt.Println("🚀 Fancy UI running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
