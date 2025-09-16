package controllers

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
        "html"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jordan-wright/unindexed"
)

// ErrInvalidRequest is thrown when a request with an invalid structure is
// received
var ErrInvalidRequest = errors.New("Invalid request")

// ErrCampaignComplete is thrown when an event is received for a campaign that
// has already been marked as complete.
var ErrCampaignComplete = errors.New("Event received on completed campaign")

// TransparencyResponse is the JSON response provided when a third-party
// makes a request to the transparency handler.
type TransparencyResponse struct {
	Server         string    `json:"server"`
	ContactAddress string    `json:"contact_address"`
	SendDate       time.Time `json:"send_date"`
}

// TransparencySuffix (when appended to a valid result ID), will cause Gophish
// to return a transparency response.
const TransparencySuffix = "+"

// PhishingServerOption is a functional option that is used to configure the
// the phishing server
type PhishingServerOption func(*PhishingServer)

// PhishingServer is an HTTP server that implements the campaign event
// handlers, such as email open tracking, click tracking, and more.
type PhishingServer struct {
	server         *http.Server
	config         config.PhishServer
	contactAddress string
}

// NewPhishingServer returns a new instance of the phishing server with
// provided options applied.
func NewPhishingServer(config config.PhishServer, options ...PhishingServerOption) *PhishingServer {
	defaultServer := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         config.ListenURL,
	}
	ps := &PhishingServer{
		server: defaultServer,
		config: config,
	}
	for _, opt := range options {
		opt(ps)
	}
	ps.registerRoutes()
	return ps
}

// WithContactAddress sets the contact address used by the transparency
// handlers
func WithContactAddress(addr string) PhishingServerOption {
	return func(ps *PhishingServer) {
		ps.contactAddress = addr
	}
}

// Start launches the phishing server, listening on the configured address.
func (ps *PhishingServer) Start() {
	if ps.config.UseTLS {
		// Only support TLS 1.2 and above - ref #1691, #1689
		ps.server.TLSConfig = defaultTLSConfig
		err := util.CheckAndCreateSSL(ps.config.CertPath, ps.config.KeyPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Starting phishing server at https://%s", ps.config.ListenURL)
		log.Fatal(ps.server.ListenAndServeTLS(ps.config.CertPath, ps.config.KeyPath))
	}
	// If TLS isn't configured, just listen on HTTP
	log.Infof("Starting phishing server at http://%s", ps.config.ListenURL)
	log.Fatal(ps.server.ListenAndServe())
}

// Shutdown attempts to gracefully shutdown the server.
func (ps *PhishingServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return ps.server.Shutdown(ctx)
}

// CreatePhishingRouter creates the router that handles phishing connections.
func (ps *PhishingServer) registerRoutes() {
	router := mux.NewRouter()
	fileServer := http.FileServer(unindexed.Dir("./static/endpoint/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	router.HandleFunc("/track", ps.TrackHandler)
	router.HandleFunc("/robots.txt", ps.RobotsHandler)
	router.HandleFunc("/{path:.*}/track", ps.TrackHandler)
	router.HandleFunc("/{path:.*}/report", ps.ReportHandler)
	router.HandleFunc("/report", ps.ReportHandler)

	// 첨부파일 열람 경로 추가
	router.HandleFunc("/fileopen", ps.FileOpenHandler).Methods("GET")

	// 신고 폼(메모 입력) 경로 추가
	router.HandleFunc("/report-form", ps.ReportFormGet).Methods("GET")
	router.HandleFunc("/report-form", ps.ReportFormPost).Methods("POST")

	// 교육(수신자) 라우트 추가 — 같은 패키지이므로 접두사 없이 함수명만 씀
	router.HandleFunc("/training/landing", TrainingLandingHandler).Methods("GET")
	router.HandleFunc("/api/training/progress", TrainingProgressAPI).Methods("POST")
	router.HandleFunc("/training/certificate", TrainingCertificateHandler).Methods("GET")

	// 루트("/") 처리
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    rid := r.URL.Query().Get(models.RecipientParameter)
	    if rid != "" {
	        // rid 파라미터가 있으면 원래 피싱 랜딩 페이지 핸들러로 넘김
	        ps.PhishHandler(w, r)
	        return
	    }
	    // rid가 없으면 신고 폼으로 리다이렉트
	    http.Redirect(w, r, "/report-form", http.StatusFound)
	}).Methods("GET")

	router.HandleFunc("/{path:.*}", ps.PhishHandler)

	// Setup GZIP compression
	gzipWrapper, _ := gziphandler.NewGzipLevelHandler(gzip.BestCompression)
	phishHandler := gzipWrapper(router)

	// Respect X-Forwarded-For and X-Real-IP headers in case we're behind a
	// reverse proxy.
	phishHandler = handlers.ProxyHeaders(phishHandler)

	// Setup logging
	phishHandler = handlers.CombinedLoggingHandler(log.Writer(), phishHandler)
	ps.server.Handler = phishHandler
}

// /fileopen FileOpenHandler 추가
func (ps *PhishingServer) FileOpenHandler(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()

    rid := strings.TrimSpace(r.Form.Get(models.RecipientParameter)) // "rid"
    rawURL := strings.TrimSpace(r.Form.Get("url"))

    // 공백으로 끝나면 transparency "+" 보정
    if strings.HasSuffix(rid, " ") {
        rid = strings.TrimRight(rid, " ") + TransparencySuffix
    }

    // 실제 id
    id := strings.TrimSuffix(rid, TransparencySuffix)

    target := ""
    //if rawURL != "" && isSafeInternalPath(rawURL) {
    if rawURL != "" {
        target = rawURL
    } else {
        target = "/static/warning.html?rid=" + url.QueryEscape(id)
    }

    // '유효한 rid' + '프리뷰가 아닐 때'만 이벤트/업데이트 시도
    if rid != "" && !strings.HasPrefix(id, models.PreviewPrefix) {
        if rs, err := models.GetResult(id); err == nil {
            // EventDetails 구성(간단 버전)
            d := models.EventDetails{
                Payload: r.Form,
                Browser: map[string]string{
                    "user-agent": r.Header.Get("User-Agent"),
                    "address":    r.RemoteAddr, // 필요하면 SplitHostPort 적용
                },
            }
            if err := rs.HandleAttachmentExecuted(d); err != nil {
                log.Errorf("fileopen: handle attachment open failed (rid=%s): %v", id, err)
            }
        } else {
            log.Infof("fileopen: no result for rid=%s; redirecting to %s", id, target)
        }
    }

    http.Redirect(w, r, target, http.StatusFound)
}

// 내부 경로만 허용하는 간단한 필터
func isSafeInternalPath(u string) bool {
    lu := strings.ToLower(u)
    if strings.HasPrefix(lu, "http://") || strings.HasPrefix(lu, "https://") {
        return false
    }
    // 데이터/자바스크립트 스킴 방지
    if strings.HasPrefix(lu, "data:") || strings.HasPrefix(lu, "javascript:") {
        return false
    }
    // 공백/제어문자 제거
    u = strings.TrimSpace(u)
    // 상대/루트 경로 허용
    if strings.HasPrefix(u, "/") || strings.HasPrefix(u, "./") || strings.HasPrefix(u, "../") {
        return true
    }
    // warning.html 처럼 단순 파일명도 허용 (static 하위 등)
    if !strings.Contains(u, "://") && !strings.HasPrefix(u, "//") {
        return true
    }
    return false
}

// GET /report-form?rid=...
func (ps *PhishingServer) ReportFormGet(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get(models.RecipientParameter) // "rid"
	// rid 없어도 폼은 열 수 있도록 허용 (POST에서 fallback 매칭 처리)
	// if rid == "" { ... }  <-- 차단하지 않음

	fmt.Fprintf(w, `<!doctype html>
<html lang="ko">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>해킹 메일 신고</title>
<style>
  :root{
    --header-top:#253445; --header-bot:#0f1a24;
    --bg:#f5f7fa; --panel:#ffffff; --border:#e1e5ea; --heading:#2c3e50;
    --text:#2b2f33; --muted:#6c7a89; --primary:#337ab7; --primary-dark:#286090;
    --control-bg:#fff; --control-border:#ccd4dd; --chip-bg:#f8fafc;
  }
  *{box-sizing:border-box} html,body{height:100%%}
  body{
    margin:0;
    font:14px/1.45 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,"Apple SD Gothic Neo","Noto Sans KR",sans-serif;
    color:var(--text);
    background:
      linear-gradient(180deg, var(--header-top), var(--header-bot) 280px),
      var(--bg);
    background-repeat:no-repeat;
  }
  .container{max-width:900px;margin:40px auto;padding:0 20px}
  .panel{background:var(--panel);border:1px solid var(--border);border-radius:4px;box-shadow:0 1px 2px rgba(0,0,0,.03)}
  .panel-heading{padding:14px 18px;border-bottom:1px solid var(--border);text-align:center}
  .panel-heading h1{margin:0;font-size:32px;color:var(--heading);font-weight:800}
  .panel-body{padding:18px}
  .form-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
  .row-2{grid-column:1 / -1}
  .form-group{margin-bottom:14px}
  label{display:block;margin-bottom:6px;color:#44525f;font-weight:600;font-size:13px}
  .form-control{
    width:100%%;display:block;padding:8px 10px;border:1px solid var(--control-border);
    border-radius:4px;background:var(--control-bg);color:var(--text);outline:none;transition:border .15s ease
  }
  .form-control:focus{border-color:#99b5d1}
  .form-control::placeholder {color:#aeb6bf;opacity:1;}
  .help{font-size:12px;color:var(--muted);margin-top:4px}
  .chips{display:flex;flex-wrap:wrap;gap:10px}
  .chip{display:inline-flex;align-items:center;gap:6px;padding:7px 10px;border:1px solid var(--control-border);border-radius:999px;background:var(--chip-bg)}
  .panel-footer{padding:14px 18px;border-top:1px solid var(--border);display:flex;gap:10px;justify-content:flex-end;background:#fafbfd}
  .btn{display:inline-block;padding:8px 14px;border-radius:4px;text-decoration:none;cursor:pointer;border:1px solid transparent;font-weight:600}
  .btn-default{background:#fff;border-color:var(--control-border);color:var(--heading)}
  .btn-default:hover{background:#f6f8fb}
  .btn-primary{background:var(--primary);border-color:var(--primary-dark);color:#fff}
  .btn-primary:hover{background:var(--primary-dark)}
  .alert{display:none;margin:0 18px 14px;border:1px solid #f1c4c0;background:#fdecea;color:#6b1b12;border-radius:4px;padding:10px 12px}
  .alert.show{display:block}
  @media (max-width:720px){ .form-grid{grid-template-columns:1fr} }
</style>
</head>
<body>
  <div class="container">
    <div class="panel">
      <div class="panel-heading"><h1>해킹 메일 신고</h1></div>

      <div id="formAlert" class="alert" role="alert" aria-live="polite"></div>

      <form id="reportForm" method="POST" action="/report-form" novalidate>
        <div class="panel-body">
          <!-- RID: 있을 수도/없을 수도 있음 -->
          <input type="hidden" name="rid" value="%s"/>
          <!-- 숨김: 메일 열람 일시(자동 설정) -->
          <input type="hidden" name="viewed_at" id="viewed_at"/>

          <div class="form-grid">
            <div class="form-group">
              <label for="reporter_name">신고자</label>
              <input class="form-control" id="reporter_name" type="text" name="reporter_name" placeholder="홍길동" required minlength="2" maxlength="50">
            </div>

            <div class="form-group">
              <label for="reporter_email">신고자 메일</label>
              <input class="form-control" id="reporter_email" type="email" name="reporter_email" placeholder="user@example.com" required maxlength="120">
            </div>

            <div class="form-group">
              <label for="mail_subject">해킹메일 제목</label>
              <input class="form-control" id="mail_subject" type="text" name="mail_subject" placeholder="[인사팀] 급여 정정 안내" required minlength="2" maxlength="200">
            </div>

            <div class="form-group">
              <label for="mail_from">해킹메일 주소</label>
              <input class="form-control" id="mail_from" type="email" name="mail_from" placeholder="hr@example.com" required maxlength="120">
            </div>

            <!-- div class="form-group row-2">
              <label>행위 여부 <span class="help">(최소 1개 이상 선택)</span></label>
              <div class="chips" id="behaviors">
                <label class="chip"><input type="checkbox" name="opened" value="yes"> 메일 열람</label>
                <label class="chip"><input type="checkbox" name="clicked" value="yes"> 링크 클릭</label>
                <label class="chip"><input type="checkbox" name="submitted" value="yes"> 정보 입력</label>
                <label class="chip"><input type="checkbox" name="downloaded" value="yes"> 파일 다운</label>
                <label class="chip"><input type="checkbox" name="executed" value="yes"> 파일 실행</label>
              </div>
            </div -->
            <!-- div class="form-group row-2">
              <div class="chips" id="behaviors">
                <input type="hidden" name="opened" />
                <input type="hidden" name="clicked"/ >
                <input type="hidden" name="submitted" />
                <input type="hidden" name="downloaded" />
                <input type="hidden" name="executed" />
              </div>
            </div -->
          </div>
        </div>

        <div class="panel-footer">
          <button type="reset" class="btn btn-default">초기화</button>
          <button type="submit" class="btn btn-primary">해킹 메일 신고</button>
        </div>
      </form>
    </div>
  </div>

<script>
  // 숨김 viewed_at 자동 값
  (function(){
    var el = document.getElementById('viewed_at');
    if (el && !el.value) {
      const d = new Date();
      d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
      el.value = d.toISOString();
    }
  })();

  // 유효성 + 체크 최소 1개
  (function(){
    const form = document.getElementById('reportForm');
    const alertBox = document.getElementById('formAlert');
    form.addEventListener('submit', function(e){
      alertBox.classList.remove('show');
      alertBox.textContent = '';
      if (!form.checkValidity()) {
        e.preventDefault();
        alertBox.textContent = '필수 항목을 올바르게 입력해 주세요.';
        alertBox.classList.add('show');
        form.reportValidity();
        return;
      }
      //const anyChecked = Array.from(form.querySelectorAll('#behaviors input[type=checkbox]')).some(ch => ch.checked);
      //if (!anyChecked) {
        //e.preventDefault();
        //alertBox.textContent = '행위 여부는 최소 1개 이상 선택해야 합니다.';
        //alertBox.classList.add('show');
      //}
    }, false);
  })();
</script>
</body></html>`, html.EscapeString(rid))
}

// POST /report-form
func (ps *PhishingServer) ReportFormPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	// 입력값
	get := func(k string) string { return strings.TrimSpace(r.Form.Get(k)) }
	rid := get(models.RecipientParameter)
	viewedAt      := get("viewed_at")
	reporterName  := get("reporter_name")
	reporterEmail := get("reporter_email") // 수신자 이메일
	mailFrom      := get("mail_from")      // 발신자 이메일
	mailSubject   := get("mail_subject")
	yn  := func(k string) string {
		if strings.ToLower(get(k)) == "yes" { return "Yes" }
		return "No"
	}
        opened     := yn("opened")
	clicked    := yn("clicked")
	submitted  := yn("submitted")
	downloaded := yn("downloaded")
	executed   := yn("executed")

	// --- 서버측 검증 ---
	if len(reporterName) < 2 || len(mailSubject) < 2 {
		http.Error(w, "invalid input", http.StatusBadRequest); return
	}
	if !strings.Contains(reporterEmail, "@") || !strings.Contains(mailFrom, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest); return
	}
	//if clicked=="No" && submitted=="No" && downloaded=="No" && executed=="No" {
		//http.Error(w, "at least one behavior must be selected", http.StatusBadRequest); return
	//}
	// -------------------

	var res *models.Result
	// 1) rid 우선 시도
	if rid != "" {
		id := strings.TrimSuffix(rid, TransparencySuffix)
		if rs, err := models.GetResult(id); err == nil {
			res = &rs
		}
	}
	// 2) rid 실패 시 → 이메일+제목으로 매칭
	if res == nil {
		if rs, err := models.FindResultByEmailAndRenderedSubject(reporterEmail, mailSubject); err == nil {
			res = rs
		} else {
			// 매칭 실패: 기록만 남기고 감사 페이지로 이동
			log.Infof("report-form: no match by email+subject (email=%s, subject=%s) → thank-you", reporterEmail, mailSubject)
                        http.Redirect(w, r, "/static/thank-you.html", http.StatusSeeOther)
			return
		}
	}

	// report_note 요약 저장
        res.ReportNote = fmt.Sprintf(
                `[신고 일시: %s] [신고자: %s (%s)] [발신자: %s] [메일 제목: %s] [메일 열람: %s] [링크 클릭: %s] [정보 입력: %s] [파일 다운로드: %s] [파일 실행: %s]`,
                viewedAt, reporterName, reporterEmail, mailFrom, mailSubject, opened, clicked, submitted, downloaded, executed,
        )

	// 이벤트 Details 에 원본 폼 포함
	d := models.EventDetails{
		Payload: r.Form,
		Browser: map[string]string{},
	}

	// Email Reported 처리
	if err := res.HandleEmailReport(d); err != nil {
		log.Error(err)
		// 처리 실패여도 사용자 경험은 동일하게 감사 페이지로
		http.Redirect(w, r, "/static/thank-you.html", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/static/thank-you.html", http.StatusSeeOther)
}

// TrackHandler tracks emails as they are opened, updating the status for the given Result
func (ps *PhishingServer) TrackHandler(w http.ResponseWriter, r *http.Request) {
	r, err := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	// Check for a preview
	if _, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		http.ServeFile(w, r, "static/images/pixel.png")
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	d := ctx.Get(r, "details").(models.EventDetails)

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		ps.TransparencyHandler(w, r)
		return
	}

	err = rs.HandleEmailOpened(d)
	if err != nil {
		log.Error(err)
	}
	http.ServeFile(w, r, "static/images/pixel.png")
}

// ReportHandler tracks emails as they are reported, updating the status for the given Result
func (ps *PhishingServer) ReportHandler(w http.ResponseWriter, r *http.Request) {
	r, err := setupContext(r)
	w.Header().Set("Access-Control-Allow-Origin", "*") // To allow Chrome extensions (or other pages) to report a campaign without violating CORS
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	// Check for a preview
	if _, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	d := ctx.Get(r, "details").(models.EventDetails)

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		ps.TransparencyHandler(w, r)
		return
	}

	err = rs.HandleEmailReport(d)
	if err != nil {
		log.Error(err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// PhishHandler handles incoming client connections and registers the associated actions performed
// (such as clicked link, etc.)
func (ps *PhishingServer) PhishHandler(w http.ResponseWriter, r *http.Request) {
	r, err := setupContext(r)
	if err != nil {
		// Log the error if it wasn't something we can safely ignore
		if err != ErrInvalidRequest && err != ErrCampaignComplete {
			log.Error(err)
		}
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Server", config.ServerName) // Useful for checking if this is a GoPhish server (e.g. for campaign reporting plugins)
	var ptx models.PhishingTemplateContext
	// Check for a preview
	if preview, ok := ctx.Get(r, "result").(models.EmailRequest); ok {
		ptx, err = models.NewPhishingTemplateContext(&preview, preview.BaseRecipient, preview.RId)
		if err != nil {
			log.Error(err)
			http.NotFound(w, r)
			return
		}
		p, err := models.GetPage(preview.PageId, preview.UserId)
		if err != nil {
			log.Error(err)
			http.NotFound(w, r)
			return
		}
		renderPhishResponse(w, r, ptx, p)
		return
	}
	rs := ctx.Get(r, "result").(models.Result)
	rid := ctx.Get(r, "rid").(string)
	c := ctx.Get(r, "campaign").(models.Campaign)
	d := ctx.Get(r, "details").(models.EventDetails)

	// Check for a transparency request
	if strings.HasSuffix(rid, TransparencySuffix) {
		ps.TransparencyHandler(w, r)
		return
	}

	p, err := models.GetPage(c.PageId, c.UserId)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == "GET":
		err = rs.HandleClickedLink(d)
		if err != nil {
			log.Error(err)
		}
	case r.Method == "POST":
		err = rs.HandleFormSubmit(d)
		if err != nil {
			log.Error(err)
		}
	}
	ptx, err = models.NewPhishingTemplateContext(&c, rs.BaseRecipient, rs.RId)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
	}
	renderPhishResponse(w, r, ptx, p)
}

// renderPhishResponse handles rendering the correct response to the phishing
// connection. This usually involves writing out the page HTML or redirecting
// the user to the correct URL.
func renderPhishResponse(w http.ResponseWriter, r *http.Request, ptx models.PhishingTemplateContext, p models.Page) {
	// If the request was a form submit and a redirect URL was specified, we
	// should send the user to that URL
	if r.Method == "POST" {
		if p.RedirectURL != "" {
			redirectURL, err := models.ExecuteTemplate(p.RedirectURL, ptx)
			if err != nil {
				log.Error(err)
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
	}
	// Otherwise, we just need to write out the templated HTML
	html, err := models.ExecuteTemplate(p.HTML, ptx)
	if err != nil {
		log.Error(err)
		http.NotFound(w, r)
		return
	}
	w.Write([]byte(html))
}

// RobotsHandler prevents search engines, etc. from indexing phishing materials
func (ps *PhishingServer) RobotsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *\nDisallow: /")
}

// TransparencyHandler returns a TransparencyResponse for the provided result
// and campaign.
func (ps *PhishingServer) TransparencyHandler(w http.ResponseWriter, r *http.Request) {
	rs := ctx.Get(r, "result").(models.Result)
	tr := &TransparencyResponse{
		Server:         config.ServerName,
		SendDate:       rs.SendDate,
		ContactAddress: ps.contactAddress,
	}
	api.JSONResponse(w, tr, http.StatusOK)
}

// setupContext handles some of the administrative work around receiving a new
// request, such as checking the result ID, the campaign, etc.
func setupContext(r *http.Request) (*http.Request, error) {
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return r, err
	}
	rid := r.Form.Get(models.RecipientParameter)
	if rid == "" {
		return r, ErrInvalidRequest
	}
	// Since we want to support the common case of adding a "+" to indicate a
	// transparency request, we need to take care to handle the case where the
	// request ends with a space, since a "+" is technically reserved for use
	// as a URL encoding of a space.
	if strings.HasSuffix(rid, " ") {
		// We'll trim off the space
		rid = strings.TrimRight(rid, " ")
		// Then we'll add the transparency suffix
		rid = fmt.Sprintf("%s%s", rid, TransparencySuffix)
	}
	// Finally, if this is a transparency request, we'll need to verify that
	// a valid rid has been provided, so we'll look up the result with a
	// trimmed parameter.
	id := strings.TrimSuffix(rid, TransparencySuffix)
	// Check to see if this is a preview or a real result
	if strings.HasPrefix(id, models.PreviewPrefix) {
		rs, err := models.GetEmailRequestByResultId(id)
		if err != nil {
			return r, err
		}
		r = ctx.Set(r, "result", rs)
		return r, nil
	}
	rs, err := models.GetResult(id)
	if err != nil {
		return r, err
	}
	c, err := models.GetCampaign(rs.CampaignId, rs.UserId)
	if err != nil {
		log.Error(err)
		return r, err
	}
	// Don't process events for completed campaigns
	if c.Status == models.CampaignComplete {
		return r, ErrCampaignComplete
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	// Handle post processing such as GeoIP
	err = rs.UpdateGeo(ip)
	if err != nil {
		log.Error(err)
	}
	d := models.EventDetails{
		Payload: r.Form,
		Browser: make(map[string]string),
	}
	d.Browser["address"] = ip
	d.Browser["user-agent"] = r.Header.Get("User-Agent")

	r = ctx.Set(r, "rid", rid)
	r = ctx.Set(r, "result", rs)
	r = ctx.Set(r, "campaign", c)
	r = ctx.Set(r, "details", d)
	return r, nil
}
