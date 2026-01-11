#!/bin/bash

# MySQL N+1問題デモ - 大量テストデータ生成スクリプト

set -e

# 設定
MYSQL_HOST="${DB_HOST:-localhost}"
MYSQL_PORT="${DB_PORT:-3306}"
MYSQL_USER="${DB_USER:-demo_user}"
MYSQL_PASSWORD="${DB_PASSWORD:-demo_password}"
MYSQL_DATABASE="${DB_NAME:-demo_db}"

# データ生成数
DEPARTMENT_COUNT=50
EMPLOYEE_COUNT=1000
ORDER_COUNT=500
MAX_DETAILS_PER_ORDER=10

echo "=== MySQL N+1問題デモ - 大量テストデータ生成 ==="
echo ""
echo "接続先: ${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DATABASE}"
echo "生成データ:"
echo "  - 部署: ${DEPARTMENT_COUNT}件"
echo "  - 社員: ${EMPLOYEE_COUNT}件"
echo "  - 受注: ${ORDER_COUNT}件"
echo "  - 受注明細: 最大${MAX_DETAILS_PER_ORDER}件/受注"
echo ""

# MySQLコマンド
MYSQL_CMD="mysql -h ${MYSQL_HOST} -P ${MYSQL_PORT} -u ${MYSQL_USER} -p${MYSQL_PASSWORD} ${MYSQL_DATABASE}"

# 部署データ生成
echo "1. 部署データを生成中..."
for i in $(seq 1 ${DEPARTMENT_COUNT}); do
    LOCATIONS=("東京" "大阪" "名古屋" "福岡" "札幌" "横浜" "神戸" "京都" "仙台" "広島")
    LOCATION=${LOCATIONS[$((RANDOM % ${#LOCATIONS[@]}))]}
    
    echo "INSERT IGNORE INTO departments (department_name, location) VALUES ('部署${i}', '${LOCATION}');" | ${MYSQL_CMD} 2>/dev/null || true
done
echo "   ✓ 部署データ生成完了"

# 社員データ生成
echo ""
echo "2. 社員データを生成中..."
FIRST_NAMES=("太郎" "花子" "一郎" "美咲" "健太" "雅子" "正樹" "真理" "大輝" "優子")
LAST_NAMES=("田中" "佐藤" "鈴木" "高橋" "山田" "中村" "伊藤" "渡辺" "小林" "加藤")

for i in $(seq 1 ${EMPLOYEE_COUNT}); do
    FIRST_NAME=${FIRST_NAMES[$((RANDOM % ${#FIRST_NAMES[@]}))]}
    LAST_NAME=${LAST_NAMES[$((RANDOM % ${#LAST_NAMES[@]}))]}
    DEPT_ID=$((RANDOM % DEPARTMENT_COUNT + 1))
    SALARY=$((3000000 + RANDOM % 5000000))
    HIRE_YEAR=$((2015 + RANDOM % 10))
    HIRE_MONTH=$((RANDOM % 12 + 1))
    HIRE_DAY=$((RANDOM % 28 + 1))
    
    echo "INSERT IGNORE INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES ('${FIRST_NAME}', '${LAST_NAME}', 'employee${i}@company.com', ${DEPT_ID}, ${SALARY}, '${HIRE_YEAR}-$(printf '%02d' ${HIRE_MONTH})-$(printf '%02d' ${HIRE_DAY})');" | ${MYSQL_CMD} 2>/dev/null || true
    
    if [ $((i % 100)) -eq 0 ]; then
        echo "   ${i}/${EMPLOYEE_COUNT} 件処理完了..."
    fi
done
echo "   ✓ 社員データ生成完了"

# 受注データ生成
echo ""
echo "3. 受注データを生成中..."
CUSTOMER_NAMES=("株式会社ABC商事" "有限会社XYZ販売" "株式会社DEF企画" "合同会社GHI物産" "株式会社JKL通商")

for i in $(seq 1 ${ORDER_COUNT}); do
    CUSTOMER_ID=$((1000 + RANDOM % 100))
    CUSTOMER_NAME=${CUSTOMER_NAMES[$((RANDOM % ${#CUSTOMER_NAMES[@]}))]}
    DAYS_AGO=$((RANDOM % 90))
    TOTAL_AMOUNT=$((10000 + RANDOM % 500000))
    STATUSES=("PENDING" "PROCESSING" "COMPLETED" "CANCELLED")
    STATUS=${STATUSES[$((RANDOM % ${#STATUSES[@]}))]}
    
    echo "INSERT INTO orders (customer_id, customer_name, order_date, total_amount, status) VALUES (${CUSTOMER_ID}, '${CUSTOMER_NAME}', DATE_SUB(CURDATE(), INTERVAL ${DAYS_AGO} DAY), ${TOTAL_AMOUNT}, '${STATUS}');" | ${MYSQL_CMD} 2>/dev/null || true
    
    # 受注明細を生成
    ORDER_ID=$i
    DETAIL_COUNT=$((RANDOM % MAX_DETAILS_PER_ORDER + 1))
    
    for j in $(seq 1 ${DETAIL_COUNT}); do
        PRODUCT_ID=$((2000 + RANDOM % 100))
        PRODUCT_NAMES=("ノートパソコン" "モニター" "キーボード" "マウス" "USB-Cハブ" "Webカメラ" "ヘッドセット" "外付けHDD")
        PRODUCT_NAME=${PRODUCT_NAMES[$((RANDOM % ${#PRODUCT_NAMES[@]}))]}
        QUANTITY=$((1 + RANDOM % 20))
        UNIT_PRICE=$((1000 + RANDOM % 100000))
        
        echo "INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES (${ORDER_ID}, ${PRODUCT_ID}, '${PRODUCT_NAME}', ${QUANTITY}, ${UNIT_PRICE});" | ${MYSQL_CMD} 2>/dev/null || true
    done
    
    if [ $((i % 50)) -eq 0 ]; then
        echo "   ${i}/${ORDER_COUNT} 件処理完了..."
    fi
done
echo "   ✓ 受注データ生成完了"

# 統計情報更新
echo ""
echo "4. 統計情報を更新中..."
echo "ANALYZE TABLE departments, employees, orders, order_details;" | ${MYSQL_CMD} 2>/dev/null || true
echo "   ✓ 統計情報更新完了"

# 結果表示
echo ""
echo "=== データ生成完了 ==="
echo ""
echo "生成されたデータ件数:"
echo "SELECT 'departments' as table_name, COUNT(*) as count FROM departments UNION ALL SELECT 'employees', COUNT(*) FROM employees UNION ALL SELECT 'orders', COUNT(*) FROM orders UNION ALL SELECT 'order_details', COUNT(*) FROM order_details;" | ${MYSQL_CMD} 2>/dev/null || echo "(件数取得エラー)"
