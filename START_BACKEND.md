# Start Backend - เริ่ม Backend Server

## ปัญหา
```
curl: (7) Failed to connect to localhost port 8080
```

Backend ไม่ได้ running

## วิธีแก้ไข

### Step 1: ตรวจสอบ Docker Services

```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
docker-compose ps
```

**ต้องเห็น**:
- `postgres` container running
- `redis` container running (optional)

**ถ้าไม่ running**:
```bash
docker-compose up -d
```

### Step 2: ตรวจสอบ Database Migrations

```bash
make mu
```

**ควรเห็น**: "✅ Migrations completed!"

### Step 3: Start Backend

#### วิธีที่ 1: Start Backend แยก (แนะนำสำหรับ debug)

```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
make start-backend
```

**ควรเห็น**:
```
🚀 Starting backend...
Starting application...
Database initialized successfully
Price refresh worker started
Server starting at 0.0.0.0:8080
```

#### วิธีที่ 2: Start ทั้งหมดพร้อมกัน

```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
make start
```

### Step 4: ตรวจสอบว่า Backend Running

**ใน terminal อื่น**:
```bash
curl http://localhost:8080/health
```

**ควรได้**:
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

**หรือเปิดเบราว์เซอร์**:
```
http://localhost:8080/swagger/index.html
```

## Troubleshooting

### ปัญหา: Database connection failed

**Error message**:
```
Failed to initialize database
```

**แก้ไข**:
1. ตรวจสอบว่า Docker services running:
   ```bash
   docker-compose ps
   ```

2. ถ้าไม่ running:
   ```bash
   docker-compose up -d
   ```

3. ตรวจสอบ config.json:
   ```bash
   cat configs/config.json
   ```

4. รอ database ready:
   ```bash
   sleep 5
   make start-backend
   ```

### ปัญหา: Port 8080 already in use

**Error message**:
```
bind: address already in use
```

**แก้ไข**:
```bash
# หา process ที่ใช้ port 8080
lsof -i :8080

# Kill process (แทน PID ด้วย process ID)
kill -9 PID

# หรือเปลี่ยน port ใน config.json
```

### ปัญหา: Config file not found

**Error message**:
```
panic: config not found
```

**แก้ไข**:
```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
cp configs/config.example.json configs/config.json
```

## Quick Start Checklist

- [ ] Docker services running (`docker-compose ps`)
- [ ] Migrations completed (`make mu`)
- [ ] Config file exists (`configs/config.json`)
- [ ] Backend started (`make start-backend`)
- [ ] Health check works (`curl http://localhost:8080/health`)

## Common Commands

```bash
# Start Docker services
docker-compose up -d

# Run migrations
make mu

# Start backend
make start-backend

# Check backend health
curl http://localhost:8080/health

# View logs
docker-compose logs -f
```
