# School Enrollment Backend

Skeleton backend cho hệ thống tuyển sinh, dùng Go + Fiber + GORM + Postgres.

## Tech stack

- Go
- Fiber
- GORM
- PostgreSQL
- golang-migrate
- JWT
- bcrypt
- slog
- Air

## Cấu trúc thư mục

```text
.
├── cmd/server
├── internal/common
├── internal/config
├── internal/database
├── internal/middleware
├── internal/modules
├── migrations
└── pkg
```

## Yêu cầu môi trường

- Go 1.24+
- PostgreSQL 14+

## Chạy local

1. Tạo file env:

```bash
cp .env.example .env
```

2. Cập nhật biến môi trường kết nối Postgres trong `.env`.

3. Tải dependency:

```bash
go mod tidy
```

4. Chạy server:

```bash
go run ./cmd/server
```

Server mặc định chạy ở `http://localhost:8080`.

## Hot reload với Air

```bash
go install github.com/air-verse/air@latest
air
```

## Health check

```bash
curl http://localhost:8080/api/v1/health
```

Kỳ vọng:

```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "app": "school-enrollment-be",
    "environment": "local",
    "status": "ok",
    "time": "2026-04-24T00:00:00Z"
  }
}
```

## Migration

Thư mục `migrations/` được tạo sẵn để dùng với `golang-migrate`. Skeleton này không tự tạo hay chỉnh sửa schema SQL.

Ví dụ tạo migration mới:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate create -ext sql -dir migrations -seq create_students_table
```

Ví dụ chạy migration:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/school_enrollment?sslmode=disable" up
```

## API hiện có

- `GET /api/v1/health`

## Ghi chú mở rộng

- `internal/modules` đang để trống để thêm từng business module sau.
- `internal/database/migrator.go` cung cấp helper để tích hợp migration khi cần.
- `internal/common/security` chứa helper JWT và bcrypt để tái sử dụng cho module auth sau.
