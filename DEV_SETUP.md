# Mattermost Development Setup Guide

## Yêu cầu hệ thống

- **Go**: Đã cài tại `C:\Program Files\Go`
- **Node.js**: v18+ với npm
- **Docker Desktop**: Đang chạy (cho PostgreSQL, Redis, etc.)
- **MSYS2** (optional): Cho `make` command

## Quick Start

### Cách 1: Chạy script tự động (Khuyến nghị)

```powershell
# Từ thư mục gốc mattermost
.\start-dev.ps1
```

Script này sẽ tự động:
- Kiểm tra Docker đang chạy
- Start backend server (port 8065)
- Start webpack dev server (port 9005)

### Cách 2: Chạy thủ công

**Terminal 1 - Backend Server:**
```powershell
cd server
.\run-server.bat
```

**Terminal 2 - Webpack Dev Server:**
```powershell
cd webapp\channels
npm run dev-server
```

## URLs

| Service | URL |
|---------|-----|
| Frontend (Dev) | http://localhost:9005 |
| Backend API | http://localhost:8065 |
| System Console | http://localhost:9005/admin_console |

## Cấu hình đã thiết lập

### CORS (server/config/config.json)
```json
"AllowCorsFrom": "http://localhost:9005",
"CorsAllowCredentials": true
```

### Plugin Directories
```json
"Directory": "D:/Workspaces/projects/mattermost/server/plugins",
"ClientDirectory": "D:/Workspaces/projects/mattermost/server/client/plugins"
```

## Troubleshooting

### Port 8065 đã bị chiếm
```powershell
# Tìm process
netstat -ano | findstr :8065

# Kill process (thay PID)
taskkill /F /PID <PID>
```

### Port 9005 đã bị chiếm
```powershell
netstat -ano | findstr :9005
taskkill /F /PID <PID>
```

### WebSocket errors
- Đảm bảo backend server đang chạy trước
- Kiểm tra CORS config trong `server/config/config.json`

### Plugins disabled
- Đảm bảo thư mục `server/client/plugins` và `server/plugins` tồn tại
- Restart server sau khi tạo thư mục

## Dừng Development

```powershell
# Dừng tất cả
.\stop-dev.ps1

# Hoặc thủ công
taskkill /F /IM go.exe
taskkill /F /IM node.exe
```

## Cấu trúc thư mục quan trọng

```
mattermost/
├── server/
│   ├── config/config.json    # Server config
│   ├── run-server.bat        # Script start server
│   ├── plugins/              # Server plugins
│   └── client/plugins/       # Client plugins
├── webapp/
│   └── channels/             # Main webapp
│       └── webpack.config.js # Webpack config
├── start-dev.ps1             # Script start all
└── stop-dev.ps1              # Script stop all
```

