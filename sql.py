import csv
import pymysql

# 数据库配置
DB_CONFIG = {
    'host': 'localhost',
    'user': 'root',
    'password': '123456',
    'db': 'teach_u',
    'charset': 'utf8mb4'
}


def import_csv_to_db(csv_path):
    # 统一路径格式
    def format_path(path):
        return path.replace('\\', '/')  # 统一转斜杠

    conn = pymysql.connect(**DB_CONFIG)
    try:
        with conn.cursor() as cursor:
            with open(csv_path, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                for row in reader:
                    sql = """INSERT INTO resources 
                            (object_key, file_name, grade, subject, file_size, file_type)
                            VALUES (%s, %s, %s, %s, %s, %s)"""
                    cursor.execute(sql, (
                        format_path(row['object_key']),
                        row['file_name'],
                        row['grade'],
                        row['subject'],
                        row['file_size'],
                        row['file_type']
                    ))
        conn.commit()
    finally:
        conn.close()


def search_resources(subject, grade, keyword):
    conn = pymysql.connect(**DB_CONFIG)
    try:
        with conn.cursor(pymysql.cursors.DictCursor) as cursor:
            sql = """SELECT * FROM resources
                    WHERE subject = %s
                    AND grade = %s
                    AND object_key LIKE %s"""
            cursor.execute(sql, (subject, grade, f'%{keyword}%'))
            return cursor.fetchall()
    finally:
        conn.close()

# 使用示例
if __name__ == "__main__":
    # CSV导入
    #import_csv_to_db('resources.csv')
    
    
    # 资源搜索
    results = search_resources('英语', '八年级', 'Section A')
    for res in results:
        print(f"{res['file_name']}")