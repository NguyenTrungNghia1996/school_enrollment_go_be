# School Enrollment Backend - APP_SUMMARY.md

Tài liệu này tóm tắt thông tin quan trọng về ứng dụng School Enrollment Backend dành cho developer và AI coding agent.

## 1. Tổng quan ứng dụng
- **Ứng dụng**: Hệ thống backend cho quy trình tuyển sinh học sinh/sinh viên.
- **Mục tiêu**: Cung cấp API để quản lý người dùng (thí sinh), admin, phân quyền (RBAC), và trong tương lai là quản lý hồ sơ xét tuyển.
- **Đối tượng sử dụng**: Thí sinh (User) và Cán bộ tuyển sinh/Quản trị viên (Admin).
- **Trạng thái hiện tại**: Skeleton dự án với đầy đủ nền tảng về Auth, Admin management, và User management.

## 2. Công nghệ sử dụng
- **Backend Framework**: [Go Fiber v2](https://gofiber.io/) - Framework web nhanh, hiệu suất cao.
- **Database**: PostgreSQL (v14+).
- **ORM**: [GORM](https://gorm.io/) - Hỗ trợ thao tác database mạnh mẽ.
- **Authentication**: JWT (JSON Web Token) với `golang-jwt/jwt/v5`.
- **Password Hashing**: `golang.org/x/crypto/bcrypt`.
- **Database Migration**: `golang-migrate/migrate/v4`.
- **Hot Reload**: `Air` (để phát triển local).
- **Logging**: `slog` (thư viện log chuẩn của Go).

## 3. Cấu trúc thư mục
- `cmd/server/`: Điểm khởi chạy ứng dụng (main.go).
- `internal/config/`: Quản lý cấu hình ứng dụng từ file `.env` và struct.
- `internal/database/`: Kết nối database, định nghĩa GORM models (`models.go`), và seeder dữ liệu.
- `internal/common/`: Chứa các hàm tiện ích dùng chung (Response, Pagination, Errors, Security).
- `internal/middleware/`: Các middleware cho Fiber (Auth, Logger, v.v.).
- `internal/modules/`: Chứa logic nghiệp vụ chia theo module (Handler-Service-Repository).
- `migrations/`: Các file SQL migration.
- `docs/`: Chứa OpenAPI specification (`openapi.yaml`).
- `pkg/`: Thư viện tiện ích có thể tái sử dụng bên ngoài project.

## 4. Luồng nghiệp vụ chính
- **Đăng nhập / Phân quyền**:
    - **Admin Auth**: Đăng nhập admin bằng username/password, trả về JWT. Hỗ trợ phân quyền dựa trên Role Groups và Permissions (bitwise OR).
    - **User Auth**: Đăng nhập cho người dùng thường (thí sinh).
- **Quản trị hệ thống (Admin)**:
    - Quản lý Role Groups: Tạo nhóm quyền, gán permission key và value (bit).
    - Quản lý Menus: Phân cấp menu và gán bit quyền để hiển thị trên Dashboard.
    - Quản lý Admin Users: Tạo/Sửa/Reset pass và gán Role Groups cho cán bộ.
- **Quản lý người dùng (Users)**: Admin quản lý danh sách thí sinh.
- **Chức năng dự kiến**: Nộp hồ sơ, Quản lý điểm thi, Upload file minh chứng (đang trong quá trình phát triển).

## 5. Database
- `admin_users`: Lưu thông tin quản trị viên.
- `role_groups`: Nhóm các vai trò (ví dụ: Admin, Reviewer).
- `admin_user_role_groups`: Bảng trung gian gán Admin vào các Role Groups (Many-to-Many).
- `role_group_permissions`: Lưu trữ các permission key và bit value tương ứng cho từng nhóm quyền.
- `menus`: Lưu cấu trúc menu dashboard, liên kết với bit quyền.
- `users`: Lưu thông tin thí sinh/người dùng thường.

## 6. API chính (v1)
### Module Auth
- `POST /api/v1/admin/auth/login`: Admin đăng nhập.
- `GET /api/v1/admin/auth/me`: Lấy thông tin admin hiện tại.
- `GET /api/v1/admin/auth/permissions`: Lấy danh sách quyền đã gộp của admin.
- `POST /api/v1/user/auth/login`: Thí sinh đăng nhập.

### Module Admin/System
- `GET/POST/PUT/PATCH /api/v1/admin/admin-users`: Quản lý cán bộ.
- `GET/POST/PUT/DELETE /api/v1/admin/role-groups`: Quản lý nhóm quyền.
- `GET/POST/PUT/DELETE /api/v1/admin/menus`: Quản lý menu.

### Module Users
- `GET/POST/PUT/PATCH /api/v1/admin/users`: Admin quản lý thí sinh.

### Hệ thống
- `GET /api/v1/health`: Kiểm tra trạng thái hệ thống và database.

## 7. Quy ước code
- **Tên file**: Viết thường, phân cách bằng dấu gạch dưới (ví dụ: `http.go`, `service_test.go`).
- **Tổ chức module**: Mỗi module trong `internal/modules` gồm:
    - `repo.go`: Thao tác DB (Repository pattern).
    - `service.go`: Logic nghiệp vụ (Service pattern).
    - `http.go`: Định nghĩa Route và Handler (Controller).
- **Response JSON**: Luôn trả về format chuẩn qua `common.Success` hoặc `common.Error`.
    - Thành công: `{ "success": true, "message": "...", "data": ... }`
    - Thất bại: `{ "success": false, "error": { "code": "...", "message": "...", "details": ... } }`
- **Lỗi**: Sử dụng `common.APIError` với mã code định danh rõ ràng.

## 8. Cách chạy project
- **Cài đặt dependency**: `go mod tidy`
- **Chạy dev (Hot reload)**: `air`
- **Chạy trực tiếp**: `go run ./cmd/server`
- **Migrate database**:
    - Sử dụng tool `migrate`.
    - Lệnh: `migrate -path migrations -database "postgres://..." up`
- **Biến môi trường**: Xem tại `.env.example`.

## 9. Những lưu ý quan trọng
- **Cập nhật OpenAPI**: Khi thay đổi Route hoặc Request/Response body, **BẮT BUỘC** cập nhật `docs/openapi.yaml`.
- **Kiểm tra mã nguồn**: Luôn chạy `go vet ./...` và `go test ./...` trước khi commit.
- **Xác thực Admin**: Hầu hết các API admin yêu cầu Bearer Token và được kiểm tra qua middleware `RequireAdminAuth`.
- **Phân quyền**: Hệ thống dùng cơ chế bitwise cho permissions, cần cẩn thận khi gán bit để tránh xung đột.
