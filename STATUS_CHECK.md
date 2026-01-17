# Status Check - ตรวจสอบสถานะ Running Services

## ตรวจสอบ Frontend (Next.js)

Frontend กำลังทำงานถ้าเห็น:
```
✓ Ready in [time]ms
✓ Compiled /[page] in [time]ms
GET /[page] 200 in [time]ms
```

**Status**: ✅ Running at `http://localhost:3000`

## ตรวจสอบ Backend (Go API)

### วิธีที่ 1: ตรวจสอบ Health Check

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

### วิธีที่ 2: ตรวจสอบ Swagger UI

เปิดเบราว์เซอร์:
```
http://localhost:8080/swagger/index.html
```

ถ้าเปิดได้ → Backend running ✅
ถ้าเปิดไม่ได้ → Backend ไม่ได้ running ❌

### วิธีที่ 3: ตรวจสอบ Process

```bash
# macOS/Linux
lsof -i :8080  # Backend
lsof -i :3000  # Frontend
```

## วิธีแก้ไข "exit status 1"

`exit status 1` หมายถึง process มี error หรือ exit ด้วย error code

### ถ้า Backend exit status 1

**ตรวจสอบ**:
```bash
# ดู backend logs (ใน terminal ที่ run make start-backend)
# หรือตรวจสอบ error messages
```

**แก้ไข**:
1. ตรวจสอบว่า config.json มีอยู่และถูกต้อง
2. ตรวจสอบว่า database running (`docker-compose ps`)
3. ตรวจสอบว่า migrations รันแล้ว (`make mu`)
4. รัน backend แยกเพื่อดู error:
   ```bash
   cd /Users/golfz/Workspaces/golfz/jonosize/project
   make start-backend
   ```

### ถ้า Frontend มีปัญหา

Frontend ดูเหมือนจะทำงานปกติ (เห็น Compiled และ 200 status)

ถ้ามีปัญหา:
1. ตรวจสอบ Browser Console (F12)
2. ตรวจสอบ Network tab
3. ดู logs ใน terminal

## Quick Status Check

```bash
# Check Backend
curl http://localhost:8080/health && echo "✅ Backend OK" || echo "❌ Backend not running"

# Check Frontend
curl http://localhost:3000 > /dev/null 2>&1 && echo "✅ Frontend OK" || echo "❌ Frontend not running"
```

## ถ้า Backend ไม่ Running

1. **Start Backend**:
   ```bash
   cd /Users/golfz/Workspaces/golfz/jonosize/project
   make start-backend
   ```

2. **ตรวจสอบ Error Messages**:
   - ดู logs ใน terminal
   - ตรวจสอบว่า config.json มีอยู่
   - ตรวจสอบว่า database running

3. **ตรวจสอบ Dependencies**:
   ```bash
   # ตรวจสอบว่า Docker running
   docker-compose ps

   # ตรวจสอบว่า migrations รันแล้ว
   make mu
   ```
