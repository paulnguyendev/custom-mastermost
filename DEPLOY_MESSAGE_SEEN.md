# Deploy Message Seen Feature to Production

Hướng dẫn deploy thủ công tính năng Message Seen lên server production.

## 1. Build trên máy local (Windows)

### Build Server (Linux binary)
```powershell
cd D:\Workspaces\projects\mattermost\server

# Build cho Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mattermost ./cmd/mattermost
```

### Build Webapp
```powershell
cd D:\Workspaces\projects\mattermost\webapp
npm install
npm run build --workspace=channels
```

## 2. Backup trên Server Production

```bash
# SSH vào server
ssh user@your-server

# Backup database
pg_dump -U mmuser -d mattermost > backup_$(date +%Y%m%d).sql

# Backup binary và client hiện tại
cp /opt/mattermost/bin/mattermost /opt/mattermost/bin/mattermost.bak
cp -r /opt/mattermost/client /opt/mattermost/client.bak
```

## 3. Chạy Migration Database

Chạy SQL trực tiếp trên PostgreSQL:

```bash
psql -U mmuser -d mattermost
```

```sql
CREATE TABLE IF NOT EXISTS ReadReceipts (
    PostId VARCHAR(26) NOT NULL,
    UserId VARCHAR(26) NOT NULL,
    ChannelId VARCHAR(26) NOT NULL,
    SeenAt BIGINT NOT NULL DEFAULT 0,
    ExpireAt BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (PostId, UserId)
);

CREATE INDEX IF NOT EXISTS idx_readreceipts_post_id ON ReadReceipts(PostId);
CREATE INDEX IF NOT EXISTS idx_readreceipts_user_id ON ReadReceipts(UserId);
CREATE INDEX IF NOT EXISTS idx_readreceipts_channel_id ON ReadReceipts(ChannelId);
```

## 4. Upload Files lên Server

### Từ Windows (dùng SCP hoặc WinSCP)

**Server binary:**
```
Local:  D:\Workspaces\projects\mattermost\server\bin\mattermost
Remote: /opt/mattermost/bin/mattermost
```

**Webapp client:**
```
Local:  D:\Workspaces\projects\mattermost\webapp\channels\dist\*
Remote: /opt/mattermost/client/
```

### Hoặc dùng SCP command:
```powershell
scp server/bin/mattermost user@server:/opt/mattermost/bin/
scp -r webapp/channels/dist/* user@server:/opt/mattermost/client/
```

## 5. Restart Server

```bash
# Stop server
sudo systemctl stop mattermost

# Set permission
chmod +x /opt/mattermost/bin/mattermost

# Start server
sudo systemctl start mattermost

# Kiểm tra logs
tail -f /opt/mattermost/logs/mattermost.log
```

## 6. Verify

```bash
# Kiểm tra bảng đã tạo
psql -U mmuser -d mattermost -c "\d ReadReceipts"
```

Test: Mở 2 browser, User A gửi tin → User B xem → User A thấy "Seen"

## Rollback

```bash
sudo systemctl stop mattermost
cp /opt/mattermost/bin/mattermost.bak /opt/mattermost/bin/mattermost
cp -r /opt/mattermost/client.bak/* /opt/mattermost/client/
psql -U mmuser -d mattermost -c "DROP TABLE IF EXISTS ReadReceipts;"
sudo systemctl start mattermost
```

