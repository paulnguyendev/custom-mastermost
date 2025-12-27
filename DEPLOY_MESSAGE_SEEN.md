# Deploy Message Seen Feature to Production

Hướng dẫn deploy tính năng Message Seen lên server production đang chạy.

## Prerequisites

- Server production đang chạy Mattermost
- Quyền truy cập SSH vào server
- Quyền truy cập database PostgreSQL

## 1. Backup Database

```bash
# SSH vào server production
ssh user@your-production-server

# Backup database trước khi deploy
pg_dump -U mmuser -d mattermost > backup_$(date +%Y%m%d_%H%M%S).sql
```

## 2. Pull Code Mới

```bash
cd /path/to/mattermost

# Pull code từ branch custom
git fetch custom
git checkout custom
git pull custom custom
```

## 3. Chạy Database Migration

Migration sẽ tạo bảng `ReadReceipts`:

```bash
# Chạy migration tự động khi start server
# Hoặc chạy manual:
cd server
./bin/mattermost db migrate
```

**Migration tạo bảng:**
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

## 4. Build Server

```bash
cd server

# Build server
make build-linux  # Linux
# hoặc
make build        # Current OS
```

## 5. Build Webapp

```bash
cd webapp

# Install dependencies (nếu cần)
npm install

# Build production
make dist
```

## 6. Restart Server

```bash
# Stop server hiện tại
sudo systemctl stop mattermost
# hoặc
pkill mattermost

# Copy files mới (nếu build ở máy khác)
cp -r server/bin/* /opt/mattermost/bin/
cp -r webapp/channels/dist/* /opt/mattermost/client/

# Start server
sudo systemctl start mattermost
# hoặc
cd /opt/mattermost
./bin/mattermost server
```

## 7. Verify Deployment

1. Kiểm tra server logs:
```bash
tail -f /opt/mattermost/logs/mattermost.log
```

2. Kiểm tra migration đã chạy:
```bash
psql -U mmuser -d mattermost -c "SELECT * FROM db_migrations WHERE version = 149;"
```

3. Kiểm tra bảng ReadReceipts:
```bash
psql -U mmuser -d mattermost -c "\d ReadReceipts"
```

4. Test tính năng:
   - Mở 2 browser với 2 user khác nhau
   - User A gửi tin nhắn
   - User B xem tin nhắn
   - User A sẽ thấy indicator "Seen" với avatar của User B

## Rollback (nếu cần)

```bash
# Stop server
sudo systemctl stop mattermost

# Restore database
psql -U mmuser -d mattermost < backup_YYYYMMDD_HHMMSS.sql

# Checkout code cũ
git checkout <previous-commit>

# Rebuild và restart
cd server && make build
sudo systemctl start mattermost
```

## Docker Deployment

Nếu dùng Docker:

```bash
# Pull image mới hoặc build
docker-compose build mattermost

# Restart container
docker-compose down
docker-compose up -d

# Kiểm tra logs
docker-compose logs -f mattermost
```

## Troubleshooting

### Lỗi migration
```bash
# Xóa migration record và chạy lại
psql -U mmuser -d mattermost -c "DELETE FROM db_migrations WHERE version = 149;"
./bin/mattermost db migrate
```

### Lỗi permission database
```bash
# Grant quyền cho user
psql -U postgres -d mattermost -c "GRANT ALL ON TABLE ReadReceipts TO mmuser;"
```

### Webapp không load
```bash
# Clear cache và rebuild
cd webapp
rm -rf node_modules/.cache
npm run build
```

