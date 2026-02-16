-- Password is abcd@123 (hashed)
INSERT INTO users (email, password_hash, full_name, role)
VALUES ('admin@example.com', '$2a$10$Y7A.473qU0G6n/U.xT.OHeC2P/1k5.wQ/oU9lT.t1vF9Ew9M/Gq.O', 'Admin User', 'admin')
ON CONFLICT (email) DO NOTHING;
