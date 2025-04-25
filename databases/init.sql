-- 创建数据库
CREATE DATABASE IF NOT EXISTS teach_u;
USE teach_u;

-- ----------------------------
-- 用户与班级相关表
-- ----------------------------

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL,
    subject VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_phone (phone_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建班级表
CREATE TABLE IF NOT EXISTS classes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    class_number INT NOT NULL,
    grade_level INT NOT NULL,
    created_user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (created_user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_class_number_grade (class_number, grade_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建学生表
CREATE TABLE IF NOT EXISTS students (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    student_name VARCHAR(255) NOT NULL,
    student_number VARCHAR(255) NOT NULL UNIQUE,
    class_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    INDEX idx_student_number (student_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 考试与成绩相关表
-- ----------------------------

-- 创建考试表
CREATE TABLE IF NOT EXISTS exams (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_user_id BIGINT UNSIGNED NOT NULL,
    class_id BIGINT UNSIGNED NOT NULL,
    exam_name VARCHAR(100) NOT NULL,
    subject VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (created_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    INDEX idx_deleted_at (deleted_at) -- 添加软删除索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建试题表
CREATE TABLE IF NOT EXISTS problems (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    exam_id BIGINT UNSIGNED NOT NULL,          -- 关联的考试ID
    keypoint VARCHAR(255) NOT NULL,            -- 知识点
    questions_number INT NOT NULL,             -- 题号
    total_score DECIMAL(5,2) NOT NULL,         -- 总分值
    content     TEXT NOT NULL,                  -- 题目内容
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,                 -- 软删除字段
    FOREIGN KEY (exam_id) REFERENCES exams(id) ON DELETE CASCADE,
    INDEX idx_exam_id (exam_id) ,             -- 添加索引提高查询性能
    INDEX idx_deleted_at (deleted_at)         -- 软删除索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建分数表
CREATE TABLE IF NOT EXISTS scores (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    student_id BIGINT UNSIGNED NOT NULL,       -- 关联的学生ID
    exam_id BIGINT UNSIGNED NOT NULL,          -- 关联的考试ID
    question_number INT NOT NULL,              -- 题号
    score DECIMAL(5,2) NOT NULL,               -- 得分
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,                 -- 软删除字段
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE, -- 修改为级联删除
    FOREIGN KEY (exam_id) REFERENCES exams(id) ON DELETE CASCADE,
    INDEX idx_student_exam (student_id, exam_id)  ,-- 联合索引
    INDEX idx_deleted_at (deleted_at),        -- 软删除索引
    INDEX idx_exam_id (exam_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建排名表
CREATE TABLE IF NOT EXISTS ranks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    exam_id BIGINT UNSIGNED NOT NULL,
    student_id BIGINT UNSIGNED NOT NULL,
    `rank` INT NOT NULL,
    student_name VARCHAR(255) NOT NULL,
    total_score DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (exam_id) REFERENCES exams(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    INDEX idx_exam_student (exam_id, student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建班级考试指标表
CREATE TABLE IF NOT EXISTS class_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    exam_id BIGINT UNSIGNED NOT NULL,
    student_count INT NOT NULL, -- 学生人数
    total_score FLOAT NOT NULL,  -- 该堂考试的总分
    avg_total_score DECIMAL(5,2) NOT NULL,  -- 总分平均分
    max_total_score DECIMAL(5,2) NOT NULL,  -- 总分最高分
    min_total_score DECIMAL(5,2) NOT NULL,  -- 总分最低分
    median_total_score DECIMAL(5,2) NOT NULL, -- 总分中位数
    std_dev DECIMAL(5,2) NOT NULL,  -- 标准差
    score_buckets JSON, -- 存储为 {"90":5, "80":12} 分数段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (exam_id) REFERENCES exams(id) ON DELETE CASCADE,
    INDEX idx_deleted_at (deleted_at) -- 添加软删除索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- 其他表
-- ----------------------------

-- 创建资源表
CREATE TABLE IF NOT EXISTS resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    object_key VARCHAR(768) CHARACTER SET utf8mb3 NOT NULL, -- 注意字符集可能需要调整为 utf8mb4
    file_name VARCHAR(255) NOT NULL,
    grade VARCHAR(50) NOT NULL,
    subject VARCHAR(50) NOT NULL,
    file_size VARCHAR(20) NOT NULL,
    file_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 添加创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 添加更新时间
    INDEX idx_subject_grade (subject, grade),
    INDEX idx_full_search (subject, grade, object_key),
    INDEX idx_object_key_full (object_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;