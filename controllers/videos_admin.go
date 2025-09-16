package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

const adminVideosDir = "static/videos"

// 리스트
func (as *AdminServer) Videos(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Videos"

	list, err := models.GetAllVideos()
	if err != nil {
		log.Error(err)
		Flash(w, r, "danger", "Failed to load videos")
	}
	// 템플릿에 바인딩할 데이터
	type pageData struct {
		templateParams
		Videos []models.Video
	}
	getTemplate(w, "videos").ExecuteTemplate(w, "base", pageData{params, list})
}

// 등록 폼
func (as *AdminServer) VideoNew(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "New Video"
	getTemplate(w, "video_form").ExecuteTemplate(w, "base", params)
}

// 등록 처리
func (as *AdminServer) VideoCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		Flash(w, r, "danger", "Invalid form"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	desc := strings.TrimSpace(r.FormValue("description"))
	durStr := strings.TrimSpace(r.FormValue("duration_seconds"))
	if title == "" {
		Flash(w, r, "danger", "Title is required"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	duration := 0
	if durStr != "" {
		if d, err := strconv.Atoi(durStr); err == nil && d >= 0 { duration = d }
	}
	// 업로드
	f, header, err := r.FormFile("file")
	if err != nil {
		Flash(w, r, "danger", "Video file required"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	defer f.Close()
	if err := os.MkdirAll(adminVideosDir, 0755); err != nil {
		Flash(w, r, "danger", "Failed to prepare storage"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" {
		Flash(w, r, "danger", "Only .mp4 allowed"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	base := sanitizeFilename(strings.TrimSuffix(header.Filename, ext))
	dst := fmt.Sprintf("%d_%s%s", time.Now().Unix(), base, ext)
	dstPath := filepath.Join(adminVideosDir, dst)
	out, err := os.Create(dstPath)
	if err != nil { Flash(w, r, "danger", "Save error"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return }
	defer out.Close()
	if _, err := io.Copy(out, f); err != nil {
		Flash(w, r, "danger", "Save error"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}

	v := &models.Video{
		Title:           title,
		Slug:            slug,
		Description:     desc,
		Filename:        "/" + filepath.ToSlash(dstPath),
		DurationSeconds: duration,
	}
	if err := models.CreateVideo(v); err != nil {
		log.Error(err)
		Flash(w, r, "danger", "DB create error"); http.Redirect(w, r, "/videos/new", http.StatusSeeOther); return
	}
	Flash(w, r, "success", "Video created")
	http.Redirect(w, r, "/videos", http.StatusSeeOther)
}

// 수정 폼
func (as *AdminServer) VideoEdit(w http.ResponseWriter, r *http.Request) {
	id := muxVarInt(r, "id")
	if id == 0 { http.NotFound(w, r); return }
	v, err := models.GetVideoByID(id)
	if err != nil { http.NotFound(w, r); return }

	type pageData struct {
		templateParams
		Video *models.Video
	}
	params := newTemplateParams(r)
	params.Title = "Edit Video"
	getTemplate(w, "video_form").ExecuteTemplate(w, "base", pageData{params, v})
}

// 수정 처리
func (as *AdminServer) VideoUpdate(w http.ResponseWriter, r *http.Request) {
	id := muxVarInt(r, "id")
	if id == 0 { http.NotFound(w, r); return }
	v, err := models.GetVideoByID(id)
	if err != nil { http.NotFound(w, r); return }
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		Flash(w, r, "danger", "Invalid form"); http.Redirect(w, r, fmt.Sprintf("/videos/%d/edit", id), http.StatusSeeOther); return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	desc := strings.TrimSpace(r.FormValue("description"))
	durStr := strings.TrimSpace(r.FormValue("duration_seconds"))
	if title != "" { v.Title = title }
	v.Slug = slug
	v.Description = desc
	if d, err := strconv.Atoi(durStr); err == nil && d >= 0 { v.DurationSeconds = d }

	// 파일 교체 (선택)
	if f, header, err := r.FormFile("file"); err == nil {
		defer f.Close()
		if err := os.MkdirAll(adminVideosDir, 0755); err == nil {
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if ext == ".mp4" {
				base := sanitizeFilename(strings.TrimSuffix(header.Filename, ext))
				dst := fmt.Sprintf("%d_%s%s", time.Now().Unix(), base, ext)
				dstPath := filepath.Join(adminVideosDir, dst)
				if out, err := os.Create(dstPath); err == nil {
					defer out.Close()
					if _, err := io.Copy(out, f); err == nil {
						v.Filename = "/" + filepath.ToSlash(dstPath)
					}
				}
			}
		}
	}

	if err := models.UpdateVideo(v); err != nil {
		log.Error(err)
		Flash(w, r, "danger", "DB update error")
	} else {
		Flash(w, r, "success", "Video updated")
	}
	http.Redirect(w, r, "/videos", http.StatusSeeOther)
}

// 삭제
func (as *AdminServer) VideoDelete(w http.ResponseWriter, r *http.Request) {
	id := muxVarInt(r, "id")
	if id == 0 { http.NotFound(w, r); return }
	if err := models.DeleteVideo(id); err != nil {
		log.Error(err)
		Flash(w, r, "danger", "DB delete error")
	} else {
		Flash(w, r, "success", "Video deleted")
	}
	http.Redirect(w, r, "/videos", http.StatusSeeOther)
}

// 미리보기
func (as *AdminServer) VideoPreview(w http.ResponseWriter, r *http.Request) {
	id := muxVarInt(r, "id")
	if id == 0 { http.NotFound(w, r); return }
	v, err := models.GetVideoByID(id)
	if err != nil { http.NotFound(w, r); return }

	type pageData struct {
		templateParams
		Video *models.Video
	}
	params := newTemplateParams(r)
	params.Title = "Preview Video"
	getTemplate(w, "video_preview").ExecuteTemplate(w, "base", pageData{params, v})
}

// helpers
func muxVarInt(r *http.Request, key string) int {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars[key])
	return id
}

func sanitizeFilename(name string) string {
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-' || r == '_' || r == '.':
			return r
		default:
			return '-'
		}
	}, name)
	if name == "" { return "video" }
	return name
}

