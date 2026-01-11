-- MySQL N+1問題デモ用DMLスクリプト
-- 初期データ投入

-- ============================================
-- 部署マスターデータ投入
-- ============================================

INSERT INTO departments (department_name, location) VALUES
('営業部', '東京'),
('開発部', '大阪'),
('人事部', '東京'),
('総務部', '名古屋'),
('マーケティング部', '福岡'),
('経理部', '東京'),
('法務部', '東京'),
('IT企画部', '大阪');

-- ============================================
-- 社員データ投入
-- ============================================

-- 営業部（department_id = 1）
INSERT INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES
('太郎', '田中', 'tanaka.taro@company.com', 1, 5000000, '2020-04-01'),
('花子', '佐藤', 'sato.hanako@company.com', 1, 4800000, '2021-04-01'),
('一郎', '鈴木', 'suzuki.ichiro@company.com', 1, 5200000, '2019-04-01');

-- 開発部（department_id = 2）
INSERT INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES
('次郎', '高橋', 'takahashi.jiro@company.com', 2, 6000000, '2018-04-01'),
('美咲', '山田', 'yamada.misaki@company.com', 2, 5800000, '2020-10-01'),
('健太', '中村', 'nakamura.kenta@company.com', 2, 5500000, '2022-04-01');

-- 人事部（department_id = 3）
INSERT INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES
('雅子', '伊藤', 'ito.masako@company.com', 3, 4500000, '2017-04-01'),
('正樹', '渡辺', 'watanabe.masaki@company.com', 3, 4700000, '2019-10-01');

-- 総務部（department_id = 4）
INSERT INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES
('真理', '小林', 'kobayashi.mari@company.com', 4, 4200000, '2021-04-01');

-- マーケティング部（department_id = 5）
INSERT INTO employees (first_name, last_name, email, department_id, salary, hire_date) VALUES
('大輝', '加藤', 'kato.daiki@company.com', 5, 5300000, '2020-04-01');

-- ============================================
-- 受注データ投入
-- ============================================

INSERT INTO orders (customer_id, customer_name, order_date, total_amount, status) VALUES
(1001, '株式会社ABC商事', DATE_SUB(CURDATE(), INTERVAL 10 DAY), 250000, 'COMPLETED'),
(1002, '有限会社XYZ販売', DATE_SUB(CURDATE(), INTERVAL 8 DAY), 180000, 'COMPLETED'),
(1003, '株式会社DEF企画', DATE_SUB(CURDATE(), INTERVAL 5 DAY), 320000, 'PENDING'),
(1001, '株式会社ABC商事', DATE_SUB(CURDATE(), INTERVAL 3 DAY), 150000, 'PROCESSING'),
(1004, '合同会社GHI物産', DATE_SUB(CURDATE(), INTERVAL 1 DAY), 420000, 'COMPLETED');

-- ============================================
-- 受注明細データ投入
-- ============================================

-- 受注1の明細
INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES
(1, 2001, 'ノートパソコン Type-A', 5, 80000),
(1, 2002, 'ワイヤレスマウス', 10, 3000),
(1, 2003, 'USB-Cハブ', 5, 4000);

-- 受注2の明細
INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES
(2, 2004, 'モニター 24インチ', 3, 45000),
(2, 2005, 'キーボード', 3, 15000);

-- 受注3の明細
INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES
(3, 2001, 'ノートパソコン Type-A', 2, 80000),
(3, 2006, 'プリンター', 1, 120000),
(3, 2007, 'スキャナー', 2, 30000);

-- 受注4の明細
INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES
(4, 2002, 'ワイヤレスマウス', 20, 3000),
(4, 2008, 'Webカメラ', 15, 6000);

-- 受注5の明細
INSERT INTO order_details (order_id, product_id, product_name, quantity, unit_price) VALUES
(5, 2009, 'サーバー', 1, 300000),
(5, 2010, 'ネットワーク機器', 2, 60000);

-- ============================================
-- 統計情報の更新
-- ============================================
ANALYZE TABLE departments;
ANALYZE TABLE employees;
ANALYZE TABLE orders;
ANALYZE TABLE order_details;
