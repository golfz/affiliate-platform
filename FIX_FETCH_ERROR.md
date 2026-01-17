# Fix "Failed to fetch" Error - แก้ปัญหา API Connection

## ปัญหา
เมื่อพยายามเพิ่ม product จาก Frontend ได้ error "Failed to fetch"

## สาเหตุที่เป็นไปได้

### 1. API Server ไม่ได้ Running (บ่อยที่สุด)

**ตรวจสอบ**:
```bash
# ตรวจสอบว่า backend running หรือไม่
curl http://localhost:8080/health
```

**แก้ไข**:
```bash
# Start backend
cd /Users/golfz/Workspaces/golfz/jonosize/project
make start-backend

# หรือ start ทั้งหมด
make start
```

### 2. Environment Variable ไม่ถูกต้อง

**ปัญหา**: Next.js ต้องใช้ `NEXT_PUBLIC_` prefix สำหรับ client-side environment variables

**ตรวจสอบ**:
```bash
# ตรวจสอบว่า NEXT_PUBLIC_API_BASE_URL ถูก set หรือไม่
# ใน browser console:
console.log(process.env.NEXT_PUBLIC_API_BASE_URL)
```

**แก้ไข**: 
- ใช้ `NEXT_PUBLIC_API_BASE_URL` แทน `API_BASE_URL` ใน client-side code
- ตรวจสอบว่า `lib/api.ts` ใช้ `process.env.NEXT_PUBLIC_API_BASE_URL` หรือ default value

### 3. CORS Error / "strict-origin-when-cross-origin"

**ตรวจสอบ**:
- เปิด Browser DevTools → Network tab
- ดู error message ที่แท็บ Console
- ดู Request Headers → Referrer Policy

**แก้ไข**:
- Backend CORS config ใน `cmd/api/main.go` ควร:
  - Allow `http://localhost:3000` ใน AllowOrigins
  - ตั้ง `AllowCredentials: true` สำหรับ Basic Auth
  - Allow `Authorization` header
- Frontend fetch ใน `lib/api.ts` ควร:
  - เพิ่ม `credentials: 'include'` ใน fetch options
  - เพิ่ม `mode: 'cors'` ใน fetch options

**หมายเหตุ**: `strict-origin-when-cross-origin` เป็น default referrer policy ของ browser ไม่ใช่ปัญหา แต่อาจทำให้เห็นใน Network tab

### 4. Basic Auth ไม่ถูกต้อง

**ตรวจสอบ**:
- เปิด Browser DevTools → Network tab
- ดู Request Headers → Authorization header

**แก้ไข**:
- ตรวจสอบว่า credentials ใน `lib/api.ts` ตรงกับ config (`admin:admin123`)

### 5. Network/Firewall Issues

**ตรวจสอบ**:
```bash
# ทดสอบ connection
curl http://localhost:8080/health

# ทดสอบ API endpoint
curl -X POST http://localhost:8080/api/products \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{"source": "https://www.lazada.co.th/products/test.html", "sourceType": "url"}'
```

## วิธีแก้ไข Step by Step

### Step 1: ตรวจสอบ Backend Running

```bash
# Terminal 1 - Start Backend
cd /Users/golfz/Workspaces/golfz/jonosize/project
make start-backend

# ตรวจสอบ logs - ควรเห็น "Server starting at 0.0.0.0:8080"
```

### Step 2: ตรวจสอบ Frontend Config

```bash
# Terminal 2 - Start Frontend
cd /Users/golfz/Workspaces/golfz/jonosize/project/apps/web
npm run dev

# ตรวจสอบว่า frontend running ที่ http://localhost:3000
```

### Step 3: ตรวจสอบ Browser Console

1. เปิด `http://localhost:3000/admin/products`
2. เปิด Browser DevTools (F12)
3. ไปที่ Console tab
4. ลองเพิ่ม product
5. ดู error message

### Step 4: ตรวจสอบ Network Request

1. เปิด Browser DevTools (F12)
2. ไปที่ Network tab
3. ลองเพิ่ม product
4. คลิกที่ failed request
5. ดู:
   - **Request URL**: ควรเป็น `http://localhost:8080/api/products`
   - **Request Headers**: ควรมี `Authorization: Basic ...`
   - **Response**: ดู error message

### Step 5: ทดสอบ API โดยตรง

```bash
# ทดสอบว่า API ทำงาน
curl -X POST http://localhost:8080/api/products \
  -u admin:admin123 \
  -H "Content-Type: application/json" \
  -d '{
    "source": "https://www.lazada.co.th/products/example-i123456.html",
    "sourceType": "url"
  }'
```

ถ้า curl ทำงานได้ → ปัญหาอยู่ที่ Frontend
ถ้า curl ไม่ทำงาน → ปัญหาอยู่ที่ Backend

## Common Errors

### Error: "Network request failed"
**สาเหตุ**: Backend ไม่ได้ running หรือ port ผิด
**แก้ไข**: Start backend ด้วย `make start-backend`

### Error: "CORS policy: No 'Access-Control-Allow-Origin'"
**สาเหตุ**: CORS config ไม่ถูกต้อง
**แก้ไข**: ตรวจสอบ CORS config ใน `cmd/api/main.go`

### Error: "401 Unauthorized"
**สาเหตุ**: Basic Auth credentials ไม่ถูกต้อง
**แก้ไข**: ตรวจสอบ credentials ใน `lib/api.ts` และ `configs/config.json`

### Error: "Failed to fetch" (generic)
**สาเหตุ**: อาจเป็น network error, CORS, หรือ API server down
**แก้ไข**: 
1. ตรวจสอบ backend logs
2. ตรวจสอบ browser console
3. ตรวจสอบ network tab

## Debug Checklist

- [ ] Backend running ที่ `http://localhost:8080`
- [ ] Frontend running ที่ `http://localhost:3000`
- [ ] `curl http://localhost:8080/health` ทำงาน
- [ ] `NEXT_PUBLIC_API_BASE_URL` ถูก set หรือ default ถูกต้อง
- [ ] CORS config allow `http://localhost:3000`
- [ ] Basic Auth credentials ถูกต้อง
- [ ] Browser console ไม่มี error อื่นๆ

## Still Not Working?

1. ตรวจสอบ logs จาก backend:
   ```bash
   # ดู backend logs
   # ใน terminal ที่ run `make start-backend`
   ```

2. ตรวจสอบ browser console:
   - Open DevTools → Console
   - ดู error messages

3. ตรวจสอบ network tab:
   - Open DevTools → Network
   - คลิก failed request
   - ดู Request และ Response details

4. ทดสอบ API โดยตรงด้วย curl (ดู Step 5 ด้านบน)
