# Quick Start Guide - ทดลองใช้โปรเจกต์

คู่มือนี้จะช่วยให้คุณ run โปรเจกต์ได้อย่างรวดเร็ว

## Prerequisites (สิ่งที่ต้องเตรียม)

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Node.js 18+** - [Download](https://nodejs.org/)
- **Docker & Docker Compose** - [Download](https://www.docker.com/get-started)
- **Make** - มักจะติดตั้งมาพร้อมกับ macOS/Linux

## Step 0: ตรวจสอบการติดตั้ง (Optional)

ก่อนเริ่มต้น คุณสามารถตรวจสอบว่ามีทุกอย่างพร้อมหรือไม่ด้วย script `CHECK_SETUP.sh`:

```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
./CHECK_SETUP.sh
```

Script นี้จะตรวจสอบ:
- ✅ Prerequisites (Go, Node.js, Docker, Make)
- ✅ ไฟล์โปรเจกต์ที่จำเป็น (go.mod, package.json)
- ✅ Configuration files (config.json)
- ✅ Docker services status
- ✅ Dependencies (Go modules, Node modules)
- ✅ CI/CD configuration

**ตัวอย่างผลลัพธ์**:
```
🔍 Checking Setup for Jenosize Affiliate Platform...
==================================================

📦 Prerequisites:
✓ Go: INSTALLED (go version go1.25.3)
✓ Node.js: INSTALLED (v25.1.0)
✓ npm: INSTALLED (11.6.2)
✓ Docker: INSTALLED (Docker version 28.5.1)
✓ Docker Compose: INSTALLED (Docker Compose version 2.40.3)
✓ Make: INSTALLED (GNU Make 3.81)

📁 Project Files:
-----------------
✓ go.mod: EXISTS
✓ package.json: EXISTS
⚠ config.json: NOT FOUND (will be created on 'make init')
✓ Swagger docs: GENERATED

✅ All checks passed! Ready to run.
```

**หมายเหตุ**: 
- ⚠️ หมายถึง Warning (ไม่จำเป็นต้องแก้ไขทันที จะถูกจัดการอัตโนมัติ)
- ✗ หมายถึง Error (ต้องแก้ไขก่อน run)
- ✓ หมายถึง OK (พร้อมใช้งาน)

ถ้าเห็น warnings เกี่ยวกับ `config.json` หรือ `node_modules` ไม่ต้องกังวล เพราะจะถูกสร้าง/ติดตั้งอัตโนมัติเมื่อรัน `make init`

## Step 1: Clone และเข้า Directory

```bash
cd /Users/golfz/Workspaces/golfz/jonosize/project
```

## Step 2: Initialize Project (ครั้งแรกเท่านั้น)

```bash
make init
```

คำสั่งนี้จะทำ:
- Install Go dependencies (`go mod download`)
- Install Node.js dependencies (ถ้ามี)
- สร้าง `configs/config.json` จาก `configs/config.example.json`
- Start Docker services (PostgreSQL และ Redis)

**หมายเหตุ**: ถ้าเป็นครั้งแรก อาจต้องรอ Docker pull images

## Step 3: Run Database Migrations

```bash
make mu
```

คำสั่งนี้จะสร้างตารางทั้งหมดในฐานข้อมูล

## Step 4: (Optional) Seed Demo Data

```bash
make seed
```

คำสั่งนี้จะใส่ข้อมูลตัวอย่าง (products, campaigns, links) สำหรับทดสอบ

## Step 5: Generate Swagger Docs

```bash
make swagger
```

คำสั่งนี้จะ generate Swagger documentation

## Step 6: Start Backend และ Frontend

### วิธีที่ 1: Start ทั้งหมดพร้อมกัน

```bash
make start
```

คำสั่งนี้จะ start:
- Backend API ที่ `http://localhost:8080`
- Frontend Next.js ที่ `http://localhost:3000`

### วิธีที่ 2: Start แยกกัน (แนะนำสำหรับ debug)

Terminal 1 - Start Backend:
```bash
make start-backend
```

Terminal 2 - Start Frontend:
```bash
make start-frontend
```

## Step 7: ทดลองใช้งาน

### 7.1 ตรวจสอบ Health Check

```bash
curl http://localhost:8080/health
```

ควรได้ response:
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

### 7.2 ดู Swagger Documentation

เปิดเบราว์เซอร์:
```
http://localhost:8080/swagger/index.html
```

### 7.3 ทดสอบ Frontend

เปิดเบราว์เซอร์:
```
http://localhost:3000
```

#### Admin Pages (ใช้ Basic Auth: `admin:admin123`)
- **Products**: `http://localhost:3000/admin/products`
- **Campaigns**: `http://localhost:3000/admin/campaigns`
- **Dashboard**: `http://localhost:3000/admin/dashboard`

#### Public Campaign Page
```
http://localhost:3000/campaign/[campaign-id]
```

(หา campaign-id จาก seed data หรือสร้างใหม่)

### 7.4 ทดสอบ API Endpoints

#### 1. Add Product (ต้องใช้ Basic Auth)

```bash
curl -X POST http://localhost:8080/api/products \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{
    "source": "https://www.lazada.co.th/products/example-i123456.html",
    "sourceType": "url"
  }'
```

#### 2. Get Product Offers

```bash
# แทน PRODUCT_ID ด้วย ID ที่ได้จากคำสั่งก่อนหน้า
curl -X GET http://localhost:8080/api/products/PRODUCT_ID/offers \
  -u admin:admin123
```

#### 3. Create Campaign

```bash
curl -X POST http://localhost:8080/api/campaigns \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Summer Deal 2025",
    "utm_campaign": "summer_2025",
    "start_at": "2025-06-01T00:00:00Z",
    "end_at": "2025-08-31T23:59:59Z",
    "product_ids": ["PRODUCT_ID_HERE"]
  }'
```

#### 4. Get Public Campaign

```bash
# แทน CAMPAIGN_ID ด้วย ID ที่ได้จากคำสั่งก่อนหน้า
curl http://localhost:8080/api/campaigns/CAMPAIGN_ID/public
```

#### 5. Get Dashboard Stats

```bash
curl http://localhost:8080/api/dashboard \
  -u admin:admin123
```

## Troubleshooting (แก้ไขปัญหา)

### ปัญหา: Database connection failed

**แก้ไข**:
```bash
# ตรวจสอบว่า Docker services running
docker-compose ps

# ถ้ายังไม่ running ให้ start
docker-compose up -d

# ตรวจสอบ logs
docker-compose logs postgres
```

### ปัญหา: Port already in use

**แก้ไข**:
- Backend (8080): ตรวจสอบว่ามี process อื่นใช้ port 8080 หรือไม่
- Frontend (3000): ตรวจสอบว่ามี process อื่นใช้ port 3000 หรือไม่

```bash
# macOS/Linux - หา process ที่ใช้ port
lsof -i :8080
lsof -i :3000

# Kill process (แทน PID ด้วย process ID)
kill -9 PID
```

### ปัญหา: Go modules not found

**แก้ไข**:
```bash
cd project
go mod download
go mod tidy
```

### ปัญหา: Frontend dependencies not installed

**แก้ไข**:
```bash
cd project/apps/web
npm install
```

### ปัญหา: Config file not found

**แก้ไข**:
```bash
cd project
cp configs/config.example.json configs/config.json
# แก้ไข configs/config.json ตามความต้องการ
```

### ปัญหา: Migration failed - "unknown driver postgres"

**สาเหตุ**: `migrate` tool ถูกติดตั้งโดยไม่มี PostgreSQL driver

**แก้ไข**:
```bash
# วิธีที่ 1: ลบและติดตั้งใหม่ (แนะนำ)
rm -f $(go env GOPATH)/bin/migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

# วิธีที่ 2: ใช้ Makefile (จะติดตั้งอัตโนมัติถ้ายังไม่มี)
make mu

# ตรวจสอบว่า PATH ถูกต้อง
echo $PATH | grep -q "$(go env GOPATH)/bin" || export PATH="$(go env GOPATH)/bin:$PATH"

# ทดสอบ
migrate -version
```

**หมายเหตุ**: 
- Makefile ตรวจสอบและติดตั้ง migrate tool อัตโนมัติเมื่อรัน `make mu`
- ถ้ายังมีปัญหา ให้ลบ migrate tool เดิมและติดตั้งใหม่ด้วย `-tags 'postgres'`
- อ่าน `FIX_MIGRATE.md` สำหรับรายละเอียดเพิ่มเติม

## Useful Commands (คำสั่งที่มีประโยชน์)

```bash
# ดู logs จาก Docker services
docker-compose logs -f

# Stop Docker services
make stop

# Clean up (ลบ volumes ด้วย)
make clean

# Run tests
make test

# Run linters
make lint

# Build backend binary
make build
```

## Next Steps (ขั้นตอนต่อไป)

1. **ทดลองใช้ Admin UI**: ไปที่ `http://localhost:3000/admin/products` เพื่อเพิ่ม products
2. **สร้าง Campaign**: ไปที่ `http://localhost:3000/admin/campaigns` เพื่อสร้าง campaign
3. **ดู Analytics**: ไปที่ `http://localhost:3000/admin/dashboard` เพื่อดูสถิติ
4. **ทดสอบ Public Page**: สร้าง campaign แล้วเปิด public page
5. **ทดสอบ Redirect**: ใช้ short code จาก campaign เพื่อทดสอบ redirect และ click tracking

## Support (ขอความช่วยเหลือ)

- **ตรวจสอบการติดตั้ง**: รัน `./CHECK_SETUP.sh` เพื่อดูสถานะทั้งหมด
- **ตรวจสอบ logs**: `docker-compose logs`
- **ตรวจสอบ Swagger docs**: `http://localhost:8080/swagger/index.html`
- **อ่าน README.md**: สำหรับรายละเอียดเพิ่มเติม
