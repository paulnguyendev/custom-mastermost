# Deploy Message Seen Feature to Production

## Release Note
- Release date: 27/12/2024
- Version: 1.3

## Overview
- Thêm tính năng "Đã xem" giúp bạn biết ai đã đọc tin nhắn của mình

## Features
- Xem ai đã đọc tin nhắn: Avatar của người đã xem hiển thị ngay bên dưới tin nhắn
- Tự động cập nhật: Khi bạn cuộn đến tin nhắn, hệ thống tự động đánh dấu bạn đã xem
- Xem danh sách đầy đủ: Di chuột vào avatar để xem tất cả người đã đọc
- Cập nhật tức thì: Thấy ngay khi có người mới xem tin nhắn của bạn

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

# Backup database (sử dụng user postgres)
sudo -u postgres pg_dump -d mattermost > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup webapp
cd /opt/mattermost
sudo mv client client.bak.$(date +%Y%m%d_%H%M%S)

# Backup server binary
sudo cp bin/mattermost bin/mattermost.bak.$(date +%Y%m%d_%H%M%S)
```

## 4. Chạy Migration Database

```bash
# Kết nối vào database với user postgres
sudo -u postgres psql mattermost
```

```sql
-- Tạo bảng ReadReceipts
CREATE TABLE IF NOT EXISTS ReadReceipts (
    PostId VARCHAR(26) NOT NULL,
    UserId VARCHAR(26) NOT NULL,
    ChannelId VARCHAR(26) NOT NULL,
    SeenAt BIGINT NOT NULL DEFAULT 0,
    ExpireAt BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (PostId, UserId)
);

-- Tạo các indexes
CREATE INDEX IF NOT EXISTS idx_readreceipts_post_id ON ReadReceipts(PostId);
CREATE INDEX IF NOT EXISTS idx_readreceipts_user_id ON ReadReceipts(UserId);
CREATE INDEX IF NOT EXISTS idx_readreceipts_channel_id ON ReadReceipts(ChannelId);

-- Verify bảng đã tạo thành công
\d ReadReceipts

-- Thoát khỏi psql
\q
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
sudo unzip /tmp/protalk-server-20251227.zip
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
# Verify cấu trúc bảng ReadReceipts
sudo -u postgres psql -d mattermost -c "\d ReadReceipts"

# Kiểm tra các indexes
sudo -u postgres psql -d mattermost -c "\di *readreceipts*"

# Kiểm tra số dòng trong bảng (nên là 0 vì mới tạo)
sudo -u postgres psql -d mattermost -c "SELECT COUNT(*) FROM ReadReceipts;"
```

Test: User A gửi tin → User B xem → User A thấy "Seen"

## Rollback

```bash
sudo systemctl stop mattermost
sudo rm -rf /opt/mattermost/client
sudo mv /opt/mattermost/client.bak.YYYYMMDD_HHMMSS /opt/mattermost/client
sudo cp /opt/mattermost/bin/mattermost.bak.YYYYMMDD_HHMMSS /opt/mattermost/bin/mattermost
sudo -u postgres psql -d mattermost -c "DROP TABLE IF EXISTS ReadReceipts;"
sudo systemctl start mattermost
```



# Backup i18n hiện tại
sudo mv /opt/mattermost/i18n /opt/mattermost/i18n.bak.$(date +%Y%m%d_%H%M%S)

# Copy i18n từ client mới
sudo cp -r /opt/mattermost/client/i18n /opt/mattermost/

# Set quyền
sudo chown -R mattermost:mattermost /opt/mattermost/i18n

# Restart
sudo systemctl restart mattermost

# Kiểm tra status
sudo systemctl status mattermost



# Lấy token từ browser (F12 > Application > Local Storage > token)
# Hoặc tạo token mới
TOKEN="xcmywrsttpgyjci1aszujpybke"

# Test API seen với một post ID bất kỳ
curl -X POST https://mattermost.izi.ai.vn/api/v4/posts/o9ofd8188tnj8gx1xnr8ug4s1y/seen \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -v

# Hoặc test API get seen
curl -X GET http://localhost:8065/api/v4/posts/o9ofd8188tnj8gx1xnr8ug4s1y/seen \
  -H "Authorization: Bearer $TOKEN" \
  -v


# Kiểm tra xem binary có chứa string "InitMessageSeen" không
strings /opt/mattermost/bin/mattermost | grep -i "InitMessageSeen"

# Hoặc kiểm tra "message_seen"
strings /opt/mattermost/bin/mattermost | grep -i "message_seen"



# Search trong các file JS (bỏ qua plugins)
find . -name "*.js" -not -path "./plugins/*" -exec grep -l "markMessageAsSeen" {} \; | head -10

# Hoặc search cụ thể hơn
grep -r "markMessageAsSeen" --include="*.js" --exclude-dir=plugins . 2>/dev/null | head -20



cd D:\Workspaces\projects\mattermost\webapp

# Clean build cũ
Remove-Item -Recurse -Force channels\dist -ErrorAction SilentlyContinue

# Build lại
npm run build --workspace=channels

# Verify code có trong build
Select-String -Path "channels\dist\*.js" -Pattern "markMessageAsSeen" | Select-Object -First 3

# Nếu thấy kết quả, đóng gói
$date = Get-Date -Format "yyyyMMdd"
Compress-Archive -Path channels\dist\* -DestinationPath "protalk-client-$date.zip" -Force

Write-Host "Build completed: protalk-client-$date.zip"
