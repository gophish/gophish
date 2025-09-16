package controllers

import (
    "encoding/json"
    "net/http"
    "html/template"
    "path/filepath"
    "strconv"
    "time"

    "github.com/gophish/gophish/models"
    // import your DB package accordingly
)

// Admin: list videos
func VideosListHandler(w http.ResponseWriter, r *http.Request) {
    videos, err := models.GetAllVideos() // implement in models
    if err != nil { http.Error(w, "internal", 500); return }
    json.NewEncoder(w).Encode(videos)
}

// Admin: create/update/delete video endpoints: implement file upload separately
// Public: landing page for training (GET): /training/landing?rid=...&vid=...
func TrainingLandingHandler(w http.ResponseWriter, r *http.Request) {
    rid := r.URL.Query().Get("rid")
    vid := r.URL.Query().Get("vid")

    // validate rid -> lookup result/user
    result, err := models.GetResultByRID(rid)
    if err != nil { http.Error(w, "invalid rid", http.StatusBadRequest); return }

    video, err := models.GetVideoByIDString(vid)
    if err != nil { http.Error(w, "invalid vid", http.StatusNotFound); return }

    // create or get enrollment
    resIDStr := strconv.Itoa(int(result.Id)) // <-- Id(소문자 d) 사용
    enroll, _ := models.GetOrCreateEnrollment(video.ID, resIDStr, result.Email)

    // Render landing template and pass video metadata + enrollment ID
    data := map[string]interface{}{
        "Video":      video,
        "Enrollment": enroll,
        "Result":     result,
    }
    renderTemplate(w, "training_landing.html", data)
}

// API: progress heartbeat (POST) => update enrollment
// Endpoint: POST /api/training/progress
// Body: { enrollment_id: int, current_time: float (seconds), duration: float (seconds), event: "play"|"pause"|"progress"|"ended" }
func TrainingProgressAPI(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        EnrollmentID int     `json:"enrollment_id"`
        CurrentTime  float64 `json:"current_time"`
        Duration     float64 `json:"duration"`
        Event        string  `json:"event"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "bad request", 400); return
    }
    enroll, err := models.GetEnrollmentByID(payload.EnrollmentID)
    if err != nil { http.Error(w, "enrollment not found", 404); return }

    now := time.Now()
    // update started time if not set
    if enroll.StartedAt == nil {
        enroll.StartedAt = &now
    }

    currentSec := int(payload.CurrentTime)
    enroll.LastPositionSec = currentSec
    // increment played seconds heuristically (server-side calculation could be more complex)
    enroll.PlayedSeconds = currentSec
    if payload.Duration > 0 {
        percent := (float64(currentSec) / payload.Duration) * 100.0
        enroll.PercentPlayed = percent
    }

    // Mark completed if event == "ended" OR percent >= 95.0
    if payload.Event == "ended" || enroll.PercentPlayed >= 95.0 {
        if !enroll.Completed {
            enroll.Completed = true
            enroll.CompletedAt = &now
            // optionally: create certificate row, issue serial
        }
    }

    if err := models.SaveEnrollment(enroll); err != nil {
        http.Error(w, "unable to save", 500); return
    }
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(enroll)
}

// API: request certificate -> generate PDF and return path or PDF content
// GET /training/certificate?enrollment_id=123
func TrainingCertificateHandler(w http.ResponseWriter, r *http.Request) {
    eidStr := r.URL.Query().Get("enrollment_id")
    eid, _ := strconv.Atoi(eidStr)
    enroll, err := models.GetEnrollmentByID(eid)
    if err != nil || !enroll.Completed {
        http.Error(w, "not eligible", 403); return
    }
    // Render certificate HTML and convert to PDF with gofpdf/wkhtmltopdf
    pdfPath, err := generateCertificatePDF(enroll)
    if err != nil {
        http.Error(w, "pdf error", 500); return
    }
    // respond with link or download
    http.Redirect(w, r, pdfPath, http.StatusFound)
}

// 간단 템플릿 렌더러 (templates/<name> 로딩)
func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	path := filepath.Join("templates", name)
	t, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// controllers/training.go
func generateCertificatePDF(e *models.Enrollment) (string, error) {
	// TODO: 실제 PDF 생성으로 교체 (gofpdf 또는 wkhtmltopdf 등)
	// 일단은 dummy 경로 반환해 리디렉트 에러를 방지
	return "/static/certs/dummy.pdf", nil
}
