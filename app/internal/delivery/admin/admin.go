package admin

import (
	"crypto/subtle"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"gorm.io/gorm"

	model "langs/internal/domain"
)

// DefaultPort is an intentionally uncommon high port so the admin panel does
// not clash with typical local services (3000/5000/8080/8081/5432 ...).
const DefaultPort = "49281"

// Server is a tiny self-contained HTTP admin panel that reports how many users
// the bot has and what each of them is currently doing. It reuses the bot's
// GORM connection and is protected with HTTP Basic Auth read from the env.
type Server struct {
	db       *gorm.DB
	username string
	password string
	port     string
	tmpl     *template.Template
}

// NewServer builds the admin server. Credentials come from ADMIN_USER /
// ADMIN_PASSWORD and the listen port from ADMIN_PORT (falling back to
// DefaultPort). If credentials are missing the server refuses to start.
func NewServer(db *gorm.DB) *Server {
	port := os.Getenv("ADMIN_PORT")
	if port == "" {
		port = DefaultPort
	}

	return &Server{
		db:       db,
		username: os.Getenv("ADMIN_USER"),
		password: os.Getenv("ADMIN_PASSWORD"),
		port:     port,
		tmpl:     template.Must(template.New("dashboard").Parse(dashboardTemplate)),
	}
}

// Start launches the admin HTTP server in a background goroutine. It is a
// no-op (with a warning) when credentials are not configured.
func (s *Server) Start() {
	if s.username == "" || s.password == "" {
		log.Println("Admin panel disabled: set ADMIN_USER and ADMIN_PASSWORD to enable it")
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.withAuth(s.handleDashboard))

	go func() {
		addr := ":" + s.port
		log.Printf("Starting admin panel on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("Admin panel server stopped: %v", err)
		}
	}()
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(s.username)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(s.password)) == 1

		if !ok || !userMatch || !passMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="Lang Bot Admin", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// userRow is the view model for a single row in the users table.
type userRow struct {
	ChatID     int64
	NativeLang string
	TargetLang string
	WordCount  int64
	BookCount  int64
	Activity   string
	LastTest   string
	Timezone   string
	QuietHours bool
}

type dashboardData struct {
	Generated  string
	TotalUsers int
	TotalWords int64
	TotalBooks int64
	ActiveNow  int
	Users      []userRow
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	var users []*model.User
	if err := s.db.Find(&users).Error; err != nil {
		http.Error(w, "Failed to load users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wordCounts := s.countByChat("words")
	bookCounts := s.countByChat("books")

	data := dashboardData{
		Generated:  time.Now().Format("2006-01-02 15:04:05 MST"),
		TotalUsers: len(users),
	}

	for _, u := range users {
		wc := wordCounts[u.ChatId]
		bc := bookCounts[u.ChatId]
		data.TotalWords += wc
		data.TotalBooks += bc

		if u.IsAwaitingInput() {
			data.ActiveNow++
		}

		data.Users = append(data.Users, userRow{
			ChatID:     u.ChatId,
			NativeLang: orDash(u.NativeLang),
			TargetLang: orDash(u.TargetLang),
			WordCount:  wc,
			BookCount:  bc,
			Activity:   activityLabel(u),
			LastTest:   lastTestLabel(u.LastTestSentAt),
			Timezone:   orDash(u.Timezone),
			QuietHours: u.QuietHours.Enabled,
		})
	}

	// Show the most engaged users first.
	sort.SliceStable(data.Users, func(i, j int) bool {
		return data.Users[i].WordCount > data.Users[j].WordCount
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("Admin dashboard render error: %v", err)
	}
}

// countByChat returns a chat_id -> row count map for the given table.
func (s *Server) countByChat(table string) map[int64]int64 {
	type row struct {
		ChatID int64
		Count  int64
	}

	var rows []row
	result := map[int64]int64{}

	if err := s.db.
		Table(table).
		Select("chat_id, COUNT(*) as count").
		Group("chat_id").
		Scan(&rows).Error; err != nil {
		log.Printf("Admin count query failed for %s: %v", table, err)
		return result
	}

	for _, rw := range rows {
		result[rw.ChatID] = rw.Count
	}
	return result
}

// activityLabel turns the user's current in-bot scenario into a human-readable
// description of what they are doing right now.
func activityLabel(u *model.User) string {
	switch u.StateData.Scenario {
	case "":
		return "Idle (menu)"
	case model.CustomTranslationScenario:
		return "Adding custom translation"
	case model.WritingExamScenario:
		return "Taking a writing exam"
	case model.WordSearchScenario:
		return "Searching words"
	case "timezone":
		return "Setting timezone"
	case "quiet_hours_start":
		return "Setting quiet hours (start)"
	case "quiet_hours_end":
		return "Setting quiet hours (end)"
	default:
		return u.StateData.Scenario
	}
}

func lastTestLabel(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02 15:04")
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Lang Bot · Admin</title>
<style>
  :root {
    --bg: #0f1420;
    --card: #1a2233;
    --card-2: #202b40;
    --text: #e8edf6;
    --muted: #8b98b0;
    --accent: #5b9dff;
    --border: #2a3550;
    --ok: #38d39f;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: var(--bg);
    color: var(--text);
    padding: 32px;
  }
  h1 { font-size: 22px; margin: 0 0 4px; }
  .sub { color: var(--muted); font-size: 13px; margin-bottom: 28px; }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 32px; }
  .card {
    background: var(--card);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 20px;
  }
  .card .label { color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
  .card .value { font-size: 32px; font-weight: 700; margin-top: 8px; }
  .card .value.accent { color: var(--accent); }
  .card .value.ok { color: var(--ok); }
  .table-wrap { background: var(--card); border: 1px solid var(--border); border-radius: 14px; overflow: hidden; }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th, td { padding: 12px 16px; text-align: left; }
  th { background: var(--card-2); color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .05em; }
  tr + tr td { border-top: 1px solid var(--border); }
  tbody tr:hover { background: rgba(91,157,255,.06); }
  .mono { font-variant-numeric: tabular-nums; }
  .pill {
    display: inline-block; padding: 3px 10px; border-radius: 999px;
    font-size: 12px; background: rgba(91,157,255,.14); color: var(--accent);
  }
  .pill.idle { background: rgba(139,152,176,.14); color: var(--muted); }
  .pill.active { background: rgba(56,211,159,.16); color: var(--ok); }
  .empty { padding: 40px; text-align: center; color: var(--muted); }
</style>
</head>
<body>
  <h1>Lang Bot · Admin</h1>
  <div class="sub">Generated {{.Generated}}</div>

  <div class="cards">
    <div class="card"><div class="label">Users</div><div class="value accent">{{.TotalUsers}}</div></div>
    <div class="card"><div class="label">Active now</div><div class="value ok">{{.ActiveNow}}</div></div>
    <div class="card"><div class="label">Saved words</div><div class="value">{{.TotalWords}}</div></div>
    <div class="card"><div class="label">Books</div><div class="value">{{.TotalBooks}}</div></div>
  </div>

  <div class="table-wrap">
    {{if .Users}}
    <table>
      <thead>
        <tr>
          <th>Chat ID</th>
          <th>Languages</th>
          <th>Words</th>
          <th>Books</th>
          <th>Doing now</th>
          <th>Last test</th>
          <th>Timezone</th>
          <th>Quiet hours</th>
        </tr>
      </thead>
      <tbody>
        {{range .Users}}
        <tr>
          <td class="mono">{{.ChatID}}</td>
          <td>{{.NativeLang}} → {{.TargetLang}}</td>
          <td class="mono">{{.WordCount}}</td>
          <td class="mono">{{.BookCount}}</td>
          <td>
            {{if eq .Activity "Idle (menu)"}}<span class="pill idle">{{.Activity}}</span>
            {{else}}<span class="pill active">{{.Activity}}</span>{{end}}
          </td>
          <td class="mono">{{.LastTest}}</td>
          <td>{{.Timezone}}</td>
          <td>{{if .QuietHours}}on{{else}}off{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
    {{else}}
    <div class="empty">No users yet.</div>
    {{end}}
  </div>
</body>
</html>`
