package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Note struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	Tags      string `json:"tags"`
	CreatedAt string `json:"created_at"`
}

const pageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Ech0 - 极简发布平台</title>
<style>
:root{--bg:#101216;--card:#1a1d26;--text:#e6e8ee;--muted:#8b90a3;--accent:#4ade80}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--text);font-family:system-ui,-apple-system,"PingFang SC","Microsoft YaHei",sans-serif;line-height:1.7}
.container{max-width:760px;margin:0 auto;padding:24px 16px}
header{display:flex;align-items:center;justify-content:space-between;padding:16px 0;border-bottom:1px solid #232838;margin-bottom:24px}
header h1{font-size:20px;color:var(--accent)}
.toolbar{display:flex;gap:10px;margin-bottom:24px}
.toolbar input{flex:1;background:var(--card);border:1px solid #2a3040;color:var(--text);padding:10px 14px;border-radius:8px}
.toolbar button{background:var(--accent);border:none;color:#0b0d10;font-weight:600;padding:10px 18px;border-radius:8px;cursor:pointer}
.note{background:var(--card);border-radius:10px;padding:16px 18px;margin-bottom:12px}
.note .text{font-size:15px;margin-bottom:6px}
.note .meta{font-size:12px;color:var(--muted)}
.note .tags{margin-top:6px}.note .tags span{font-size:12px;color:var(--accent);margin-right:8px}
.empty{color:var(--muted);text-align:center;padding:40px 0}
</style>
</head>
<body><div class="container">
<header><h1>Ech0</h1><span style="font-size:12px;color:var(--muted)">极简 · 开源 · 自托管</span></header>
<div class="toolbar"><input id="q" placeholder="搜索碎片..."><button onclick="load()">搜索</button></div>
<div id="app"><div class="empty">加载中...</div></div>
</div>
<script>
async function load(){
  const q=document.getElementById('q').value;
  const url='/api/notes'+(q?'?q='+encodeURIComponent(q):'');
  const r=await fetch(url);const d=await r.json();
  const app=document.getElementById('app');
  if(!d.length){app.innerHTML='<div class="empty">暂无内容</div>';return}
  app.innerHTML=d.map(n=>'<div class="note"><div class="text">'+n.content.replace(/</g,'&lt;')+'</div><div class="tags">'+(n.tags?n.tags.split(',').map(t=>'<span>#'+t+'</span>').join(''):'')+'</div><div class="meta">'+new Date(n.created_at).toLocaleString()+'</div></div>').join('');
}
document.getElementById('q').addEventListener('keydown',e=>{if(e.key==='Enter')load()});
load();
</script></body></html>`

func main() {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initDB()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/notes", notesHandler)
	http.HandleFunc("/api/notes/", noteHandler)
	http.HandleFunc("/api/rss", rssHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Ech0 running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDB() {
	sqlStmt := `CREATE TABLE IF NOT EXISTS notes (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  content TEXT NOT NULL,
	  tags TEXT NOT NULL DEFAULT '',
	  created_at TEXT NOT NULL
	)`
	if _, err := db.Exec(sqlStmt); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl, err := template.New("page").Parse(pageHTML)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.Execute(w, nil)
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		tag := strings.TrimSpace(r.URL.Query().Get("tag"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		size, _ := strconv.Atoi(r.URL.Query().Get("size"))
		if size < 1 || size > 50 {
			size = 20
		}
		where := "1=1"
		var args []interface{}
		if q != "" {
			where += " AND content LIKE ?"
			args = append(args, "%"+q+"%")
		}
		if tag != "" {
			where += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
		args = append(args, size, (page-1)*size)
		rows, err := db.Query("SELECT id, content, tags, created_at FROM notes WHERE "+where+" ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?", args...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		notes := []Note{}
		for rows.Next() {
			var n Note
			if err := rows.Scan(&n.ID, &n.Content, &n.Tags, &n.CreatedAt); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			notes = append(notes, n)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, 200, notes)
	case "POST":
		var n Note
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		n.Content = strings.TrimSpace(n.Content)
		if n.Content == "" {
			http.Error(w, "content is required", 400)
			return
		}
		n.Tags = strings.TrimSpace(strings.ReplaceAll(n.Tags, "，", ","))
		n.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		res, err := db.Exec("INSERT INTO notes (content, tags, created_at) VALUES (?, ?, ?)", n.Content, n.Tags, n.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		id, _ := res.LastInsertId()
		n.ID = int(id)
		writeJSON(w, 201, n)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", 405)
	}
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	switch r.Method {
	case "GET":
		var n Note
		err := db.QueryRow("SELECT id, content, tags, created_at FROM notes WHERE id = ?", id).
			Scan(&n.ID, &n.Content, &n.Tags, &n.CreatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "not found", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, 200, n)
	case "DELETE":
		res, err := db.Exec("DELETE FROM notes WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			http.Error(w, "not found", 404)
			return
		}
		writeJSON(w, 200, map[string]bool{"success": true})
	default:
		w.Header().Set("Allow", "GET, DELETE")
		http.Error(w, "method not allowed", 405)
	}
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, content, tags, created_at FROM notes ORDER BY created_at DESC LIMIT 20")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var items strings.Builder
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Content, &n.Tags, &n.CreatedAt); err != nil {
			continue
		}
		esc := strings.ReplaceAll(n.Content, "&", "&amp;")
		esc = strings.ReplaceAll(esc, "<", "&lt;")
		esc = strings.ReplaceAll(esc, ">", "&gt;")
		items.WriteString(fmt.Sprintf("<item><title>%s</title><description>%s</description><pubDate>%s</pubDate></item>", esc, esc, n.CreatedAt))
	}
	w.Header().Set("Content-Type", "application/rss+xml")
	fmt.Fprintf(w, `<rss version="2.0"><channel><title>Ech0</title><description>Micro blog</description>%s</channel></rss>`, items.String())
}
