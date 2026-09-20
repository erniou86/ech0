package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type User struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

type Room struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type Message struct {
	ID        int    `json:"id"`
	RoomID    int    `json:"room_id"`
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

const pageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Ech0 Chat</title>
  <style>
    :root {
      --bg: #0b1020;
      --panel: #121a2b;
      --panel-2: #1a2439;
      --border: #26324a;
      --text: #edf3ff;
      --muted: #9aa9c7;
      --accent: #7c8cff;
      --accent-2: #5ad1b7;
      --danger: #ff6d7a;
      --shadow: 0 18px 45px rgba(4, 10, 22, 0.55);
    }
    * { box-sizing: border-box; }
    html, body { margin: 0; height: 100%; background: linear-gradient(180deg, #0b1020, #101827 25%, #0b1020); color: var(--text); font-family: Inter, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif; }
    body { display: flex; align-items: center; justify-content: center; padding: 24px; }
    .app { width: min(1200px, 100%); min-height: 760px; display: grid; grid-template-columns: 280px 1fr; background: rgba(9, 15, 27, 0.92); border: 1px solid var(--border); border-radius: 22px; box-shadow: var(--shadow); overflow: hidden; }
    .sidebar { background: rgba(17, 24, 38, 0.86); border-right: 1px solid var(--border); padding: 22px 18px; }
    .brand { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; }
    .brand .name { font-size: 1.25rem; font-weight: 700; letter-spacing: 0.04em; }
    .brand .badge { font-size: 0.7rem; border: 1px solid var(--border); padding: 4px 8px; border-radius: 999px; color: var(--muted); }
    .section-title { color: var(--muted); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.12em; margin: 18px 8px 10px; }
    .user-box, .room-form, .message-form { background: rgba(255,255,255,0.02); border: 1px solid var(--border); border-radius: 14px; padding: 12px; }
    .user-box input, .room-form input, .room-form textarea, .message-form input { width: 100%; border: 1px solid var(--border); background: rgba(11,16,32,0.7); color: var(--text); padding: 11px 12px; border-radius: 10px; outline: none; }
    .user-box input:focus, .room-form input:focus, .room-form textarea:focus, .message-form input:focus { border-color: rgba(124,140,255,0.85); box-shadow: 0 0 0 3px rgba(124,140,255,0.14); }
    .primary-btn, .secondary-btn { border: none; cursor: pointer; padding: 10px 14px; border-radius: 10px; font-weight: 600; transition: 0.2s ease; }
    .primary-btn { background: linear-gradient(135deg, var(--accent), #8ea1ff); color: #0a1122; }
    .secondary-btn { background: var(--panel-2); color: var(--text); border: 1px solid var(--border); }
    .primary-btn:hover, .secondary-btn:hover { transform: translateY(-1px); }
    .user-box { display: grid; gap: 8px; }
    .user-meta { display: flex; justify-content: space-between; align-items: center; margin-top: 8px; }
    .status { font-size: 0.75rem; color: var(--muted); }
    .pill { display: inline-flex; align-items: center; gap: 6px; padding: 5px 8px; border-radius: 999px; background: rgba(90, 209, 183, 0.12); color: var(--accent-2); font-size: 0.72rem; }
    .room-list { list-style: none; margin: 0; padding: 0; display: grid; gap: 8px; }
    .room-item { padding: 12px 12px; border-radius: 12px; border: 1px solid transparent; background: rgba(255,255,255,0.02); cursor: pointer; transition: background 0.2s ease, border 0.2s ease; }
    .room-item:hover { background: rgba(255,255,255,0.04); border-color: var(--border); }
    .room-item.selected { background: rgba(124,140,255,0.12); border-color: rgba(124,140,255,0.45); }
    .room-name { font-weight: 600; margin-bottom: 4px; }
    .room-desc { color: var(--muted); font-size: 0.75rem; line-height: 1.5; }
    .main { display: flex; flex-direction: column; min-height: 0; }
    .topbar { height: 76px; display: flex; align-items: center; justify-content: space-between; padding: 18px 22px; border-bottom: 1px solid var(--border); background: rgba(17, 24, 38, 0.35); }
    .topbar h2 { margin: 0; font-size: 1.3rem; }
    .topbar .meta { color: var(--muted); font-size: 0.8rem; }
    .chat-area { flex: 1; min-height: 0; display: flex; flex-direction: column; }
    .messages { flex: 1; overflow: auto; padding: 18px 22px 0; display: flex; flex-direction: column; gap: 12px; }
    .message { max-width: 70%; display: flex; flex-direction: column; }
    .message.self { align-self: flex-end; }
    .meta-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; font-size: 0.72rem; color: var(--muted); }
    .bubble { padding: 12px 14px; border-radius: 14px; line-height: 1.6; background: rgba(255,255,255,0.03); border: 1px solid var(--border); }
    .message.self .bubble { background: linear-gradient(135deg, rgba(124,140,255,0.2), rgba(90,209,183,0.14)); border-color: rgba(124,140,255,0.48); }
    .composer { border-top: 1px solid var(--border); padding: 16px 22px 22px; }
    .message-form { display: flex; gap: 10px; align-items: center; }
    .message-form input { flex: 1; }
    .empty { color: var(--muted); padding: 20px; text-align: center; }
    @media (max-width: 780px) { .app { grid-template-columns: 1fr; } .sidebar { border-right: none; border-bottom: 1px solid var(--border); } .message { max-width: 90%; } }
  </style>
</head>
<body>
  <div class="app">
    <aside class="sidebar">
      <div class="brand">
        <div class="name">Ech0 Chat</div>
        <div class="badge">全球聊天</div>
      </div>

      <div class="section-title">账号</div>
      <div class="user-box">
        <input id="displayName" placeholder="输入昵称" maxlength="32" />
        <button class="primary-btn" id="registerBtn">进入聊天</button>
        <div class="user-meta">
          <span class="status" id="userStatus">未登录</span>
          <span class="pill" id="userPill">游客</span>
        </div>
      </div>

      <div class="section-title">新房间</div>
      <div class="room-form">
        <input id="roomName" placeholder="房间名" maxlength="40" />
        <textarea id="roomDesc" placeholder="房间简介" rows="3" style="margin-top:10px; resize:none;" maxlength="200"></textarea>
        <button class="secondary-btn" id="createRoomBtn" style="margin-top:10px; width:100%;">创建房间</button>
      </div>

      <div class="section-title">房间列表</div>
      <ul class="room-list" id="roomList"></ul>
    </aside>

    <main class="main">
      <div class="topbar">
        <h2 id="currentRoomTitle">选择房间</h2>
        <div class="meta" id="currentRoomMeta">等待连接</div>
      </div>
      <div class="chat-area">
        <div class="messages" id="messages"></div>
        <div class="composer">
          <div class="message-form">
            <input id="messageInput" placeholder="输入消息，按回车发送..." maxlength="2000" />
            <button class="primary-btn" id="sendBtn">发送</button>
          </div>
        </div>
      </div>
    </main>
  </div>

  <script>
    let state = {
      user: null,
      rooms: [],
      activeRoomId: null,
      messages: []
    };

    const roomList = document.getElementById('roomList');
    const messagesEl = document.getElementById('messages');
    const currentRoomTitle = document.getElementById('currentRoomTitle');
    const currentRoomMeta = document.getElementById('currentRoomMeta');
    const userStatus = document.getElementById('userStatus');
    const userPill = document.getElementById('userPill');

    function setUser(user) {
      state.user = user;
      localStorage.setItem('ech0-user', JSON.stringify(user));
      userStatus.textContent = '已登录：' + user.display_name;
      userPill.textContent = user.display_name;
    }

    function getStoredUser() {
      try {
        const raw = localStorage.getItem('ech0-user');
        return raw ? JSON.parse(raw) : null;
      } catch (error) {
        return null;
      }
    }

    function renderRooms() {
      roomList.innerHTML = '';
      if (!state.rooms.length) {
        const li = document.createElement('li');
        li.className = 'empty';
        li.textContent = '暂无房间';
        roomList.appendChild(li);
        return;
      }

      state.rooms.forEach((room) => {
        const li = document.createElement('li');
        li.className = 'room-item' + (room.id === state.activeRoomId ? ' selected' : '');
        li.innerHTML = '<div class="room-name"># ' + escapeHtml(room.name) + '</div><div class="room-desc">' + escapeHtml(room.description || '暂无简介') + '</div>';
        li.addEventListener('click', function() {
          state.activeRoomId = room.id;
          renderRooms();
          loadMessages();
        });
        roomList.appendChild(li);
      });
    }

    function escapeHtml(value) {
      return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/\"/g, '&quot;')
        .replace(/'/g, '&#039;');
    }

    function renderMessages() {
      messagesEl.innerHTML = '';
      if (!state.messages.length) {
        const empty = document.createElement('div');
        empty.className = 'empty';
        empty.textContent = '暂无消息，发出第一条消息吧';
        messagesEl.appendChild(empty);
        return;
      }

      state.messages.forEach((msg) => {
        const wrapper = document.createElement('div');
        wrapper.className = 'message ' + (state.user && state.user.id === msg.user_id ? 'self' : '');

        const meta = document.createElement('div');
        meta.className = 'meta-row';
        meta.innerHTML = '<span>' + escapeHtml(msg.user_name || '匿名用户') + '</span><span>·</span><span>' + escapeHtml(new Date(msg.created_at).toLocaleString()) + '</span>';

        const bubble = document.createElement('div');
        bubble.className = 'bubble';
        bubble.textContent = msg.content;

        wrapper.appendChild(meta);
        wrapper.appendChild(bubble);
        messagesEl.appendChild(wrapper);
      });

      messagesEl.scrollTop = messagesEl.scrollHeight;
    }

    async function api(path, method, payload) {
      const options = {
        method: method || 'GET',
        headers: { 'Content-Type': 'application/json' }
      };
      if (payload) options.body = JSON.stringify(payload);

      const response = await fetch(path, options);
      const text = await response.text();
      let data = null;
      try { data = text ? JSON.parse(text) : null; } catch (error) { data = null; }

      if (!response.ok) {
        const message = data && data.error ? data.error : '请求失败';
        throw new Error(message);
      }
      return data;
    }

    async function registerUser() {
      const name = document.getElementById('displayName').value.trim();
      if (!name) {
        alert('请输入昵称');
        return;
      }
      try {
        const result = await api('/api/users/register', 'POST', { display_name: name });
        setUser(result);
      } catch (error) {
        alert(error.message);
      }
    }

    async function loadRooms() {
      try {
        const rooms = await api('/api/rooms', 'GET');
        state.rooms = rooms || [];
        if (!state.activeRoomId && state.rooms.length) {
          state.activeRoomId = state.rooms[0].id;
        }
        renderRooms();
        if (state.activeRoomId) {
          loadMessages();
        }
      } catch (error) {
        console.error(error);
      }
    }

    async function createRoom() {
      const name = document.getElementById('roomName').value.trim();
      const description = document.getElementById('roomDesc').value.trim();
      if (!name) {
        alert('房间名不能为空');
        return;
      }
      try {
        const room = await api('/api/rooms', 'POST', { name, description });
        document.getElementById('roomName').value = '';
        document.getElementById('roomDesc').value = '';
        state.activeRoomId = room.id;
        await loadRooms();
      } catch (error) {
        alert(error.message);
      }
    }

    async function loadMessages() {
      if (!state.activeRoomId) {
        currentRoomTitle.textContent = '选择房间';
        currentRoomMeta.textContent = '等待连接';
        messagesEl.innerHTML = '<div class="empty">请选择一个房间</div>';
        return;
      }

      const room = state.rooms.find((item) => item.id === state.activeRoomId);
      if (room) {
        currentRoomTitle.textContent = '#' + room.name;
        currentRoomMeta.textContent = room.description || '开放讨论';
      }

      try {
        const list = await api('/api/rooms/' + state.activeRoomId + '/messages', 'GET');
        state.messages = list || [];
        renderMessages();
      } catch (error) {
        console.error(error);
      }
    }

    async function sendMessage() {
      if (!state.user) {
        alert('请先登录');
        return;
      }
      if (!state.activeRoomId) {
        alert('请先选择房间');
        return;
      }

      const input = document.getElementById('messageInput');
      const content = input.value.trim();
      if (!content) return;

      try {
        await api('/api/rooms/' + state.activeRoomId + '/messages', 'POST', {
          user_id: state.user.id,
          content: content
        });
        input.value = '';
        await loadMessages();
      } catch (error) {
        alert(error.message);
      }
    }

    document.getElementById('registerBtn').addEventListener('click', registerUser);
document.getElementById('createRoomBtn').addEventListener('click', createRoom);
document.getElementById('sendBtn').addEventListener('click', sendMessage);
document.getElementById('messageInput').addEventListener('keydown', function (event) {
  if (event.key === 'Enter') {
    sendMessage();
  }
});

const storedUser = getStoredUser();
if (storedUser) {
  setUser(storedUser);
}
loadRooms();
setInterval(function() {
  if (state.activeRoomId) {
    loadMessages();
  }
}, 5000);
  </script>
</body>
</html>`

func main() {
	var err error
	db, err = openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/users/register", registerUserHandler)
	http.HandleFunc("/api/rooms", roomsHandler)
	http.HandleFunc("/api/rooms/", roomMessagesHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Ech0 Chat running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func openDB() (*sql.DB, error) {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "./chat.db"
	}
	return sql.Open("sqlite3", path)
}

func initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			display_name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(room_id) REFERENCES rooms(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_name ON rooms(name)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room_id, created_at)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(pageHTML))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	displayName := strings.TrimSpace(payload.DisplayName)
	if displayName == "" {
		writeError(w, http.StatusBadRequest, "display_name is required")
		return
	}
	if len(displayName) > 32 {
		displayName = displayName[:32]
	}

	result, err := db.Exec("INSERT INTO users (display_name, created_at) VALUES (?, ?)", displayName, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, User{ID: int(id), DisplayName: displayName, CreatedAt: time.Now().UTC().Format(time.RFC3339)})
}

func roomsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, name, description, created_at FROM rooms ORDER BY created_at DESC, id DESC")
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		rooms := []Room{}
		for rows.Next() {
			var room Room
			if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			rooms = append(rooms, room)
		}
		writeJSON(w, http.StatusOK, rooms)
	case http.MethodPost:
		var payload struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		name := strings.TrimSpace(payload.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "room name is required")
			return
		}
		if len(name) > 40 {
			name = name[:40]
		}
		description := strings.TrimSpace(payload.Description)
		if len(description) > 200 {
			description = description[:200]
		}

		result, err := db.Exec("INSERT INTO rooms (name, description, created_at) VALUES (?, ?, ?)", name, description, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		id, _ := result.LastInsertId()
		writeJSON(w, http.StatusCreated, Room{ID: int(id), Name: name, Description: description, CreatedAt: time.Now().UTC().Format(time.RFC3339)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func roomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		http.NotFound(w, r)
		return
	}

	roomID, err := strconv.Atoi(trimmed)
	if err != nil || roomID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	if r.Method == http.MethodGet {
		rows, err := db.Query(`
			SELECT m.id, m.room_id, m.user_id, u.display_name, m.content, m.created_at
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.room_id = ?
			ORDER BY m.created_at ASC, m.id ASC LIMIT 200
		`, roomID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		msgs := []Message{}
		for rows.Next() {
			var msg Message
			if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.UserName, &msg.Content, &msg.CreatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			msgs = append(msgs, msg)
		}
		writeJSON(w, http.StatusOK, msgs)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		UserID  int    `json:"user_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.UserID <= 0 {
		writeError(w, http.StatusBadRequest, "valid user_id is required")
		return
	}
	content := strings.TrimSpace(payload.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, "message content is required")
		return
	}
	if len(content) > 2000 {
		content = content[:2000]
	}

	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", payload.UserID).Scan(&exists); err != nil || !exists {
		writeError(w, http.StatusBadRequest, "user not found")
		return
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec("INSERT INTO messages (room_id, user_id, content, created_at) VALUES (?, ?, ?, ?)", roomID, payload.UserID, content, createdAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, _ := result.LastInsertId()

	var userName string
	if err := db.QueryRow("SELECT display_name FROM users WHERE id = ?", payload.UserID).Scan(&userName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, Message{ID: int(id), RoomID: roomID, UserID: payload.UserID, UserName: userName, Content: content, CreatedAt: createdAt})
}

func init() {
	if os.Getenv("DB_PATH") == "" {
		_ = os.Setenv("DB_PATH", "./chat.db")
	}
}

func main2() {}

func main3() {}

func route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		indexHandler(w, r)
	case r.URL.Path == "/health":
		healthHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}
