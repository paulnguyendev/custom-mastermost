# Deploy Message Seen Feature to Production

## Release Note
- Release date: 27/12/2024
- Version: 1.3

## Overview
- Thêm tính năng Message Seen (Read Receipts) để biết ai đã xem tin nhắn

## Features
- Hiển thị avatar người đã xem tin nhắn bên dưới mỗi tin
- Tự động đánh dấu đã xem khi tin nhắn xuất hiện trên màn hình
- Popover hiển thị danh sách đầy đủ người đã xem khi hover
- Real-time update qua WebSocket khi có người xem tin nhắn mới
- Persist data khi refresh trang

---

Hướng dẫn deploy thủ công tính năng Message Seen lên server production.

## 1. Build trên máy local (Windows)

### Build Server (Linux binary)
```powershell
cd D:\Workspaces\projects\mattermost\server
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/mattermost ./cmd/mattermost
```

### Build Webapp
```powershell
cd D:\Workspaces\projects\mattermost\webapp
npm install
npm run build --workspace=channels
```

### Đóng gói files
```powershell
# Zip webapp
Compress-Archive -Path D:\Workspaces\projects\mattermost\webapp\channels\dist\* -DestinationPath D:\Workspaces\projects\mattermost\webapp\protalk-client-YYYYMMDD.zip -Force

# Zip server binary
Compress-Archive -Path D:\Workspaces\projects\mattermost\server\bin\mattermost -DestinationPath D:\Workspaces\projects\mattermost\server\protalk-server-YYYYMMDD.zip -Force
```

## 2. Upload lên Server

```powershell
scp D:\Workspaces\projects\mattermost\webapp\protalk-client-YYYYMMDD.zip user@your-server:/tmp/
scp D:\Workspaces\projects\mattermost\server\protalk-server-YYYYMMDD.zip user@your-server:/tmp/
```

## 3. SSH vào Server và Backup

```bash
ssh user@your-server

# Backup database
pg_dump -U mmuser -d mattermost > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup webapp
cd /opt/mattermost
sudo mv client client.bak.$(date +%Y%m%d_%H%M%S)

# Backup server binary
sudo cp bin/mattermost bin/mattermost.bak.$(date +%Y%m%d_%H%M%S)
```

## 4. Chạy Migration Database

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

## 5. Deploy Files

```bash
# Deploy webapp
cd /opt/mattermost
sudo mkdir client
cd client
sudo unzip /tmp/protalk-client-YYYYMMDD.zip
sudo chown -R mattermost:mattermost /opt/mattermost/client

# Deploy server binary
cd /opt/mattermost/bin
sudo unzip /tmp/protalk-server-YYYYMMDD.zip
sudo chmod +x mattermost
sudo chown mattermost:mattermost mattermost
```

## 6. Restart Server

```bash
sudo systemctl restart mattermost
tail -f /opt/mattermost/logs/mattermost.log
```

## 7. Verify

```bash
psql -U mmuser -d mattermost -c "\d ReadReceipts"
```

Test: User A gửi tin → User B xem → User A thấy "Seen"

## Rollback

```bash
sudo systemctl stop mattermost
sudo rm -rf /opt/mattermost/client
sudo mv /opt/mattermost/client.bak.YYYYMMDD_HHMMSS /opt/mattermost/client
sudo cp /opt/mattermost/bin/mattermost.bak.YYYYMMDD_HHMMSS /opt/mattermost/bin/mattermost
psql -U mmuser -d mattermost -c "DROP TABLE IF EXISTS ReadReceipts;"
sudo systemctl start mattermost
```

