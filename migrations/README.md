Thư mục này dùng cho các file migration của `golang-migrate`.

Migration đầu tiên:

- `000001_init_admin_schema.up.sql`
- `000001_init_admin_schema.down.sql`

Cài CLI:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Chạy `up`:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/school_enrollment?sslmode=disable" up
```

Chạy `down`:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/school_enrollment?sslmode=disable" down 1
```
