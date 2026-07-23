package audit

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var logPath = "gophish_audit.log"

// SetLogPath sets the audit log path from config
func SetLogPath(path string) {
	logPath = path
}

func LogEvent(r *http.Request, username, event, status string) {
	// Verzeichnis anlegen falls nicht vorhanden
	dir := filepath.Dir(logPath)
	if dir != "." {
		os.MkdirAll(dir, 0750)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	defer f.Close()

	ip := r.RemoteAddr
	line := fmt.Sprintf("%s | %-10s | %-40s | IP: %-20s | %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		event,
		username,
		ip,
		status,
	)
	f.WriteString(line)
}
