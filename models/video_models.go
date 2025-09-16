package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// ======================================================
// Models
// ======================================================

type Video struct {
	ID              int       `gorm:"primary_key" json:"id"`
	Title           string    `json:"title"`
	Slug            string    `gorm:"unique_index" json:"slug"`
	Description     string    `json:"description"`
	Filename        string    `json:"filename"`          // /static/videos/xxx.mp4 또는 외부 URL
	DurationSeconds int       `json:"duration_seconds"`  // 초 단위
	Thumbnail       string    `json:"thumbnail"`         // 포스터 이미지 경로(선택)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Video) TableName() string { return "videos" }

type Enrollment struct {
	ID              int        `gorm:"primary_key" json:"id"`
	VideoID         int        `json:"video_id" gorm:"index"`
	ResultID        string     `json:"result_id" gorm:"index"` // gophish results.rid(or id)와 연동
	UserEmail       string     `json:"user_email"`
	StartedAt       *time.Time `json:"started_at"`
	LastPositionSec int        `json:"last_position_sec"`
	PlayedSeconds   int        `json:"played_seconds"`
	PercentPlayed   float64    `json:"percent_played"`
	Completed       bool       `json:"completed"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Enrollment) TableName() string { return "enrollments" }

// ======================================================
// Auto-migration (선택): 서버 시작 시 한번만 호출되면 끝.
// gophish의 초기화 시점(예: models.Setup() 또는 db 초기화 이후)에 호출하세요.
// ======================================================

func AutoMigrateTraining() error {
	if db == nil {
		return errors.New("db is nil")
	}
	// GORM이 있으면 자동 생성/스키마 갱신 (SQLite에서도 잘 동작)
	return db.AutoMigrate(&Video{}, &Enrollment{}).Error
}

// ======================================================
// Video CRUD-ish
// ======================================================

func GetAllVideos() ([]Video, error) {
	var list []Video
	if err := db.Order("created_at desc").Find(&list).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return list, nil
}

func GetVideoByID(id int) (*Video, error) {
	var v Video
	if err := db.Where("id = ?", id).First(&v).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("video not found")
		}
		return nil, err
	}
	return &v, nil
}

func GetVideoByIDString(idStr string) (*Video, error) {
	var id int
	_, err := fmtSscanf(idStr, "%d", &id)
	if err != nil {
		return nil, err
	}
	return GetVideoByID(id)
}

func CreateVideo(v *Video) error {
	if v == nil {
		return errors.New("nil video")
	}
	now := time.Now()
	v.CreatedAt, v.UpdatedAt = now, now
	return db.Create(v).Error
}

func UpdateVideo(v *Video) error {
	if v == nil || v.ID == 0 {
		return errors.New("invalid video")
	}
	v.UpdatedAt = time.Now()
	return db.Save(v).Error
}

func DeleteVideo(id int) error {
	return db.Where("id = ?", id).Delete(&Video{}).Error
}

// ======================================================
// Enrollment helpers
// ======================================================

func GetOrCreateEnrollment(videoID int, resultID string, userEmail string) (*Enrollment, error) {
	var e Enrollment
	err := db.Where("video_id = ? AND result_id = ?", videoID, resultID).First(&e).Error
	if err == nil {
		return &e, nil
	}
	if !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	// create
	now := time.Now()
	e = Enrollment{
		VideoID:         videoID,
		ResultID:        resultID,
		UserEmail:       userEmail,
		StartedAt:       nil,
		LastPositionSec: 0,
		PlayedSeconds:   0,
		PercentPlayed:   0.0,
		Completed:       false,
		CompletedAt:     nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func GetEnrollmentByID(id int) (*Enrollment, error) {
	var e Enrollment
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func SaveEnrollment(e *Enrollment) error {
	if e == nil || e.ID == 0 {
		return errors.New("invalid enrollment")
	}
	e.UpdatedAt = time.Now()
	return db.Save(e).Error
}

// ======================================================
// 작은 유틸 (fmt.Sscanf 대체: fmt를 import하지 않기위해)
// ======================================================

func fmtSscanf(s, format string, a ...interface{}) (int, error) {
	// 최소 의존으로 숫자만 파싱
	var n int
	_, err := fmtSscanfInt(s, &n)
	if err != nil {
		return 0, err
	}
	if len(a) != 1 {
		return 0, errors.New("invalid args")
	}
	switch p := a[0].(type) {
	case *int:
		*p = n
	default:
		return 0, errors.New("unsupported type")
	}
	return 1, nil
}

func fmtSscanfInt(s string, out *int) (int, error) {
	// 아주 단순한 정수 파서
	sign := 1
	i := 0
	if len(s) == 0 {
		return 0, errors.New("empty")
	}
	if s[0] == '-' {
		sign = -1
		i++
	}
	val := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, errors.New("not a number")
		}
		val = val*10 + int(c-'0')
	}
	*out = sign * val
	return 1, nil
}

