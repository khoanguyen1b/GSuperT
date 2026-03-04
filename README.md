# GSuperT - Go Gin API Service

Dự án này là một dịch vụ API backend được xây dựng bằng ngôn ngữ Go và framework Gin, hỗ trợ xác thực JWT (Access & Refresh token), phân quyền người dùng và các thao tác CRUD cơ bản cho User và Customer.

## Yêu cầu hệ thống (Prerequisites)
- Docker & Docker Compose
- Go 1.23+ (nếu chạy local không dùng Docker)
- Apple Silicon (M1/M2/M3) friendly.

## Cài đặt nhanh (Quick Start)

Thực hiện các lệnh sau để khởi chạy dự án:

1. **Sao chép file môi trường:**
   ```bash
   cp .env.example .env
   ```

2. **Khởi chạy Docker Compose:**
   ```bash
   docker compose up --build
   ```

3. **Migrations tự chạy khi API khởi động:**
   - Khi `api` service start, ứng dụng sẽ tự chạy toàn bộ SQL migrations trong `internal/migrations` (bao gồm bảng `app_settings`).
   - Nếu bạn vẫn muốn chạy thủ công, có thể dùng:
   ```bash
   docker compose exec api migrate -path internal/migrations -database "postgres://admin:abcd@123@db:5432/appdb?sslmode=disable" up
   ```

## Tài khoản mặc định (Default Accounts)
- **Admin:** `admin@example.com` / `abcd@123` (Role: `admin`)

## Ví dụ lệnh Curl

### 1. Đăng nhập (Login)
```bash
curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "admin@example.com", "password": "abcd@123"}'
```

### 2. Tạo Customer (Yêu cầu JWT Token)
```bash
curl -X POST http://localhost:8080/customers \
     -H "Authorization: Bearer <ACCESS_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"name": "Nguyen Van A", "email": "nva@example.com", "phone": "0901234567", "address": "TP. HCM"}'
```

### 3. Lấy danh sách Customers
```bash
curl -X GET http://localhost:8080/customers \
     -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 4. Tạo User mới (Chỉ dành cho Admin)
```bash
curl -X POST http://localhost:8080/users \
     -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>" \
     -H "Content-Type: application/json" \
     -d '{"email": "user1@example.com", "password": "password123", "full_name": "User One", "role": "user"}'
```

### 5. Làm mới Token (Refresh Token)
```bash
curl -X POST http://localhost:8080/auth/refresh \
     -H "Content-Type: application/json" \
     -d '{"refresh_token": "<REFRESH_TOKEN>"}'
```

### 6. Text Analyze (MVP)
`POST /text-analyze`

```bash
curl -X POST http://localhost:8080/text-analyze \
     -H "Content-Type: application/json" \
     -d '{
       "text": "I want to go out and take it with us. The beautiful car is in the city.",
       "options": {
         "linking": { "mode": "mvp", "max_chunk_words": 12 },
         "syntax": { "mode": "mvp" }
       }
     }'
```

### 7. Text Analyze với GPT Syntax
Nếu muốn kết quả `syntax` chính xác hơn, cấu hình thêm trong `.env`:

```bash
OPENAI_API_KEY=<your_openai_key>
OPENAI_MODEL=gpt-4.1-mini
OPENAI_BASE_URL=https://api.openai.com/v1
```

Sau đó gọi API với `options.syntax.mode = "gpt"`:

```bash
curl -X POST http://localhost:8080/text-analyze \
     -H "Content-Type: application/json" \
     -d '{
       "text": "I want to go out and take it with us. The beautiful car is in the city.",
       "options": {
         "linking": { "mode": "mvp", "max_chunk_words": 12 },
         "syntax": { "mode": "gpt" }
       }
     }'
```

### 8. App Settings (Lưu API key theo key-value)
Các API này yêu cầu `Authorization: Bearer <ACCESS_TOKEN_ADMIN>`.

Hiện tại enum key hỗ trợ trong source:
- `gpt_api_key`

#### 8.1 Upsert nhiều key-value
`POST /settings/bulk`

```bash
curl -X POST http://localhost:8080/settings/bulk \
     -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>" \
     -H "Content-Type: application/json" \
     -d '[
       { "key": "gpt_api_key", "value": "sk-xxxx" }
     ]'
```

#### 8.2 Lấy danh sách settings
`GET /settings`

```bash
curl -X GET http://localhost:8080/settings \
     -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>"
```

#### 8.3 Lấy setting theo key
`GET /settings/:key`

```bash
curl -X GET http://localhost:8080/settings/gpt_api_key \
     -H "Authorization: Bearer <ACCESS_TOKEN_ADMIN>"
```

## Postman Collection
Dự án cung cấp sẵn file collection để bạn có thể import vào Postman một cách nhanh chóng:
- File: `GSuperT_Collection.postman_collection.json`
- **Cách sử dụng:**
  1. Mở Postman, chọn **Import** và kéo file này vào.
  2. Collection đã được cấu hình sẵn các biến:
     - `base_url`: Mặc định là `http://localhost:8080`.
     - `access_token` & `refresh_token`: Sẽ tự động cập nhật sau khi gọi API **Login** hoặc **Refresh Token** nhờ vào script test tích hợp sẵn.
  3. Để gọi các API bảo mật (Users, Customers), hãy đảm bảo bạn đã thực hiện **Login** trước đó.

## Xử lý sự cố (Troubleshooting)

### Xung đột cổng (Port Conflict)
Nếu bạn gặp lỗi cổng 8080 hoặc 5432 đã bị sử dụng, hãy đổi `APP_PORT` hoặc `DB_PORT` trong file `.env`.

### Reset Database
Nếu muốn xóa toàn bộ dữ liệu và chạy lại từ đầu:
```bash
docker compose down -v
docker compose up --build
```

### Lưu ý cho Apple Silicon (M1/M2/M3)
Dự án sử dụng image `postgres:15-alpine` và `golang:1.23-alpine`, cả hai đều hỗ trợ tốt kiến trúc `arm64`.

## Cấu trúc thư mục
- `cmd/api`: Điểm bắt đầu của ứng dụng.
- `internal/config`: Quản lý cấu hình bằng Viper.
- `internal/db`: Kết nối cơ sở dữ liệu GORM.
- `internal/migrations`: Các file SQL migration.
- `internal/modules`: Chứa logic nghiệp vụ theo module (Auth, Users, Customers).
- `internal/common`: Các hàm bổ trợ dùng chung cho Response và Error.
- `tests`: Các bản kiểm thử đơn vị (Unit tests).
