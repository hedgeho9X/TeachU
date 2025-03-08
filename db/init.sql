-- 创建数据库
CREATE DATABASE IF NOT EXISTS teach_u;
USE teach_u;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_phone (phone_number)
);

-- 创建资源表
CREATE TABLE IF NOT EXISTS resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    object_key VARCHAR(768) CHARACTER SET utf8mb3 NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    grade VARCHAR(50) NOT NULL,
    subject VARCHAR(50) NOT NULL,
    file_size VARCHAR(20) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    INDEX idx_subject_grade (subject, grade),
    INDEX idx_full_search (subject, grade, object_key),
    INDEX idx_object_key_full (object_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
