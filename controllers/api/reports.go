package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// ReportExportData represents the data structure for exporting campaign data to MJSET
type ReportExportData struct {
	Campaign    CampaignInfo      `json:"campaign"`
	Credentials []CredentialEntry `json:"credentials"`
	Clicks      []ClickEntry      `json:"clicks"`
	Targets     []string          `json:"targets"`
	Stats       ReportStats       `json:"stats"`
}

// CampaignInfo contains basic campaign information
type CampaignInfo struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	LaunchDate  time.Time `json:"launch_date"`
	CreatedDate time.Time `json:"created_date"`
	URL         string    `json:"url"`
	Status      string    `json:"status"`
}

// CredentialEntry represents a credential submission
type CredentialEntry struct {
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// ClickEntry represents a click/visit event
type ClickEntry struct {
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportStats contains campaign statistics
type ReportStats struct {
	TotalTargets    int64 `json:"total_targets"`
	EmailsSent      int64 `json:"emails_sent"`
	EmailsOpened    int64 `json:"emails_opened"`
	LinksClicked    int64 `json:"links_clicked"`
	CredSubmitted   int64 `json:"cred_submitted"`
	EmailsReported  int64 `json:"emails_reported"`
}

// CampaignExportData exports campaign data in a format suitable for MJSET report generation
func (as *Server) CampaignExportData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	
	// Get campaign with all related data
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}

	// Build export data structure
	exportData := ReportExportData{
		Campaign: CampaignInfo{
			ID:          c.Id,
			Name:        c.Name,
			LaunchDate:  c.LaunchDate,
			CreatedDate: c.CreatedDate,
			URL:         c.URL,
			Status:      c.Status,
		},
		Credentials: []CredentialEntry{},
		Clicks:      []ClickEntry{},
		Targets:     []string{},
		Stats: ReportStats{
			TotalTargets:   int64(len(c.Results)),
			EmailsSent:     0,
			EmailsOpened:   0,
			LinksClicked:   0,
			CredSubmitted:  0,
			EmailsReported: 0,
		},
	}

	// Process results to get target emails and reported count
	for _, result := range c.Results {
		// Add target email
		exportData.Targets = append(exportData.Targets, result.Email)
		
		if result.Reported {
			exportData.Stats.EmailsReported++
		}
	}

	// Process events for both detailed data AND statistics
	for _, event := range c.Events {
		// Count statistics based on events (not just final status)
		switch event.Message {
		case models.EventSent:
			exportData.Stats.EmailsSent++
		case models.EventOpened:
			exportData.Stats.EmailsOpened++
		case models.EventClicked:
			exportData.Stats.LinksClicked++
		case models.EventDataSubmit:
			exportData.Stats.CredSubmitted++
		}
		
		// Process detailed event data
		switch event.Message {
		case models.EventDataSubmit:
			// Parse credential submission
			var details models.EventDetails
			if err := json.Unmarshal([]byte(event.Details), &details); err == nil {
				if payload := details.Payload; payload != nil {
					username := ""
					password := ""
					if usernames, ok := payload["username"]; ok && len(usernames) > 0 {
						username = usernames[0]
					}
					if passwords, ok := payload["password"]; ok && len(passwords) > 0 {
						password = passwords[0]
					}
					
					// Find the corresponding result for IP address
					ip := ""
					for _, result := range c.Results {
						if result.Email == event.Email {
							ip = result.IP
							break
						}
					}
					
					exportData.Credentials = append(exportData.Credentials, CredentialEntry{
						Email:     event.Email,
						Username:  username,
						Password:  password,
						IP:        ip,
						Timestamp: event.Time,
					})
				}
			}
		case models.EventClicked:
			// Find the corresponding result for IP address
			ip := ""
			for _, result := range c.Results {
				if result.Email == event.Email {
					ip = result.IP
					break
				}
			}
			
			exportData.Clicks = append(exportData.Clicks, ClickEntry{
				Email:     event.Email,
				IP:        ip,
				Timestamp: event.Time,
			})
		}
	}

	JSONResponse(w, exportData, http.StatusOK)
}

// GenerateReports triggers the Python report generation process
func (as *Server) GenerateReports(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	
	// Verify campaign exists and user has access
	_, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}

	// Get the current working directory to find the reports script
	workDir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Unable to determine working directory"}, http.StatusInternalServerError)
		return
	}

	// Path to the Python bridge script
	scriptPath := filepath.Join(workDir, "reports", "gophish_bridge.py")
	
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		log.Error("Python bridge script not found at: " + scriptPath)
		JSONResponse(w, models.Response{Success: false, Message: "Report generation script not found"}, http.StatusInternalServerError)
		return
	}

	// Execute the Python script
	cmd := exec.Command("python3", scriptPath, strconv.FormatInt(id, 10))
	cmd.Dir = filepath.Join(workDir, "reports")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Report generation failed: %s, Output: %s", err.Error(), string(output)))
		JSONResponse(w, models.Response{Success: false, Message: "Report generation failed: " + err.Error()}, http.StatusInternalServerError)
		return
	}

	log.Info(fmt.Sprintf("Reports generated successfully for campaign %d", id))
	JSONResponse(w, models.Response{Success: true, Message: "Reports generated successfully"}, http.StatusOK)
}

// DownloadReport serves generated report files
func (as *Server) DownloadReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	reportType := vars["type"]
	
	// Verify campaign exists and user has access
	_, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}

	// Determine file path based on report type
	workDir, _ := os.Getwd()
	reportsDir := filepath.Join(workDir, "reports", "output", fmt.Sprintf("campaign_%d", id))
	
	var fileName string
	switch reportType {
	case "executive":
		fileName = "SocialEngExecReport.pdf"
	case "confidential":
		fileName = "Confidential_Report.pdf"
	case "internal":
		fileName = "Internal_Use.pdf"
	default:
		JSONResponse(w, models.Response{Success: false, Message: "Invalid report type"}, http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(reportsDir, fileName)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		JSONResponse(w, models.Response{Success: false, Message: "Report file not found"}, http.StatusNotFound)
		return
	}

	// Serve the file
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	http.ServeFile(w, r, filePath)
}
