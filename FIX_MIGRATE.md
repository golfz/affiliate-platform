# Fix Migration Tool - แก้ปัญหา migrate tool

## ปัญหา
```
error: failed to open database: database driver: unknown driver postgres (forgotten import?)
```

## สาเหตุ
`migrate` tool ถูกติดตั้งโดยไม่มี PostgreSQL driver เพราะไม่ได้ compile ด้วย `-tags 'postgres'`

## วิธีแก้ไข

### วิธีที่ 1: ลบและติดตั้งใหม่ (แนะนำ)

```bash
# 1. ลบ migrate tool เดิม
rm -f $(go env GOPATH)/bin/migrate

# 2. ติดตั้งใหม่ด้วย PostgreSQL driver
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

# 3. ตรวจสอบว่า PATH ถูกต้อง
echo $PATH | grep -q "$(go env GOPATH)/bin" || export PATH="$(go env GOPATH)/bin:$PATH"

# 4. ทดสอบว่า migrate พร้อมใช้งาน
migrate -version

# 5. รัน migration
make mu
```

### วิธีที่ 2: ใช้ Makefile (อัตโนมัติ)

Makefile จะตรวจสอบและติดตั้ง migrate tool อัตโนมัติ:

```bash
# รัน migration (จะติดตั้ง migrate tool อัตโนมัติถ้ายังไม่มี)
make mu
```

### วิธีที่ 3: ตรวจสอบและติดตั้งด้วยตนเอง

```bash
# ตรวจสอบว่า migrate ถูกติดตั้งหรือไม่
which migrate

# ถ้าไม่พบหรือไม่มี postgres driver ให้ติดตั้งใหม่
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

# ตรวจสอบ PATH (ต้องมี $(go env GOPATH)/bin)
echo $PATH

# ถ้าไม่มี ให้เพิ่มใน ~/.zshrc หรือ ~/.bashrc
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# ทดสอบ
migrate -version
```

## ตรวจสอบว่าแก้ไขแล้ว

หลังจากติดตั้งใหม่ ควรเห็น:

```bash
$ migrate -version
v4.16.2
```

และควรสามารถใช้ `postgres://` driver ได้

## Troubleshooting

### ปัญหา: migrate command not found

**แก้ไข**:
```bash
# ตรวจสอบ PATH
echo $PATH | grep -q "$(go env GOPATH)/bin"

# ถ้าไม่มี ให้เพิ่ม
export PATH="$(go env GOPATH)/bin:$PATH"

# หรือเพิ่มใน ~/.zshrc
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### ปัญหา: ยังคงพบ error "unknown driver postgres"

**แก้ไข**:
1. ลบ migrate tool เดิมออกทั้งหมด
2. ติดตั้งใหม่ด้วย `-tags 'postgres'`
3. ตรวจสอบว่า PATH ถูกต้อง
4. รัน `migrate -version` เพื่อตรวจสอบ

### ปัญหา: Database connection failed

**แก้ไข**:
```bash
# ตรวจสอบว่า Docker services running
docker-compose ps

# ถ้ายังไม่ running
docker-compose up -d

# รอสักครู่ให้ database พร้อม
sleep 5

# รัน migration อีกครั้ง
make mu
```
