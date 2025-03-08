-- 创建数据库
CREATE DATABASE IF NOT EXISTS teach_u;
USE teach_u;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
--  手机号unique
    phone_number UNIQUE VARCHAR(20) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_phone (phone_number)
);


-- 步骤5：创建数据表（使用之前设计的表结构）
CREATE TABLE IF NOT EXISTS resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    object_key VARCHAR(1000) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    grade VARCHAR(50) NOT NULL,
    subject VARCHAR(50) NOT NULL,
    file_size VARCHAR(20) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    INDEX idx_subject_grade (subject, grade),
    INDEX idx_object_key (object_key(255))
    INDEX idx_full_search (subject, grade, object_key(255))
);

-- 创建用户电话索引
CREATE UNIQUE INDEX idx_users_phone_number ON users(phone_number);
