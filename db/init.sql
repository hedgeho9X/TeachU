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

-- 创建一些测试数据（密码都是 123456）
INSERT INTO users (phone_number, username, password_hash) VALUES
('13800138000', '测试用户1', '$2a$10$NlBC84MVb7F/sf4e6dB1HO6RiGwIYrRtoVCXtC3YNiYzVRRH5rcMC'),
('13900139000', '测试用户2', '$2a$10$NlBC84MVb7F/sf4e6dB1HO6RiGwIYrRtoVCXtC3YNiYzVRRH5rcMC');-- 创建数据库
CREATE DATABASE IF NOT EXISTS teach_u;
USE teach_u;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    username VARCHAR(50),
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建一些测试用户数据
INSERT INTO users (phone_number, username, password_hash) VALUES
('13800138000', '测试用户1', '$2a$10$NlBC84MVb7F/sf4e6dB1HO6RiGwIYrRtoVCXtC3YNiYzVRRH5rcMC'), -- 密码: password123
('13900139000', '测试用户2', '$2a$10$NlBC84MVb7F/sf4e6dB1HO6RiGwIYrRtoVCXtC3YNiYzVRRH5rcMC');

-- 创建索引
CREATE INDEX idx_users_username ON users(username);