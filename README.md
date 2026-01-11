# MySQL N+1問題デモンストレーション - Go実装

MySQL Databaseを使用したN+1問題の実例とその解決策を学習するためのGoアプリケーションです。業務アプリ開発でよく発生するN+1問題の原因、影響、そして効果的な解決策をMySQL環境で実践的にデモンストレーションします。

## 目次

- [概要](#概要)
- [N+1問題とは](#n1問題とは)
- [プロジェクト構成](#プロジェクト構成)
- [セットアップ](#セットアップ)
- [使用方法](#使用方法)
- [実装内容](#実装内容)
- [パフォーマンス比較](#パフォーマンス比較)
- [MySQL固有の最適化](#mysql固有の最適化)

## 概要

このプロジェクトは、MySQL Databaseを使用した業務アプリ開発において頻繁に発生するN+1問題を、実際のコード例とパフォーマンス測定を通して学習できるデモアプリケーションです。

### 特徴

- **実践的なコード例**: 問題のあるコードと修正されたコードの両方を提供
- **パフォーマンス測定**: 実際の実行時間を比較測定
- **MySQL最適化**: InnoDB Buffer Pool等のMySQL固有の最適化手法を活用
- **包括的な解決策**: JOIN、IN句、バッチ処理など複数のアプローチを実装
- **Docker対応**: compose.ymlで簡単に環境構築可能

## N+1問題とは

N+1問題は、データベースアクセスにおける代表的なパフォーマンス問題です：

1. **初回クエリ（1回）**: 親データを取得
2. **追加クエリ（N回）**: 各親データに対して子データを個別取得

結果として「1 + N回」のクエリが実行され、パフォーマンスが大幅に劣化します。

### 問題の例

```go
// ❌ N+1問題のあるコード
orders := getOrders()              // 1回のクエリ
for _, order := range orders {     // N回のループ
    details := getOrderDetails(order.ID)  // N回のクエリ
    // 処理...
}
```

### 解決策

```go
// ✅ 最適化されたコード
ordersWithDetails := getOrdersWithDetailsJoin()  // 1回のクエリ
for _, orderWithDetails := range ordersWithDetails {
    // 処理...
}
```

## プロジェクト構成

```shell
mysql-n-plus-1-demo/
├── cmd/
│   └── main.go                # メインアプリケーション
├── config/
│   └── config.go              # 設定管理とDB接続
├── internal/
│   ├── cache/                 # キャッシュ機能実装
│   │   ├── buffer_pool.go     # InnoDB Buffer Pool分析
│   │   └── cache_analyzer.go  # キャッシュ性能分析
│   └── service/
│       ├── cache_service.go   # キャッシュサービス
│       └── demo_service.go    # デモサービス
├── models/
│   └── models.go              # データモデル定義
├── repository/
│   ├── repository_problem.go  # N+1問題のあるリポジトリ
│   └── repository_optimized.go # 最適化されたリポジトリ
├── scripts/
│   ├── init/
│   │   ├── 01_create_tables.sql  # テーブル作成DDL
│   │   └── 02_insert_data.sql    # 初期データDML
│   └── load_test_data.sh         # 大量ダミーデータ生成スクリプト
├── compose.yml                # Docker環境構築
├── env.example                # 環境変数のサンプル
├── go.mod                     # Go modules設定
├── linter.sh                  # リンター実行スクリプト
└── README.md                  # このファイル
```

## セットアップ

### 前提条件

- Go 1.25以上
- Docker & Docker Compose（推奨）または MySQL 8.0以上

### 1. プロジェクトのクローン

```bash
git clone https://github.com/okamyuji/mysql-n-plus-1-demo
cd mysql-n-plus-1-demo
```

### 2. 依存関係のインストール

```bash
go mod tidy
```

### 3. データベースのセットアップ

#### Docker Composeを使用する場合（推奨）

```bash
# MySQLコンテナを起動
docker compose up -d

# コンテナが起動するまで待機（約30秒）
docker compose logs -f mysql
```

DDL/DMLは自動的に実行されます。

#### 手動でセットアップする場合

```bash
# DDLでテーブル作成
mysql -u demo_user -pdemo_password demo_db < scripts/init/01_create_tables.sql

# 初期データの投入
mysql -u demo_user -pdemo_password demo_db < scripts/init/02_insert_data.sql

# 大量のダミーデータ生成（負荷テスト用）
./scripts/load_test_data.sh
```

### 4. 環境設定

`.env`ファイルを作成し、MySQL接続情報を設定：

```bash
cp .env.example .env
```

`.env`ファイルを編集（Docker Composeを使用する場合はそのままでOK）：

```text
DB_HOST=localhost
DB_PORT=3306
DB_NAME=demo_db
DB_USER=demo_user
DB_PASSWORD=demo_password
```

## 使用方法

### 基本的な実行

```bash
# 全てのパフォーマンステストを実行
go run cmd/main.go

# 統計情報とサンプルデータを表示
go run cmd/main.go -stats -sample

# 過去7日間の受注データでテスト
go run cmd/main.go -days=7
```

### オプション

| オプション | 説明 |
| --------- | ---- |
| `-days=30` | 取得する受注データの日数（デフォルト: 30日） |
| `-sample` | サンプルデータを表示 |
| `-stats` | データベース統計情報を表示 |
| `-order-only` | 受注データのパフォーマンステストのみ実行 |
| `-employee-only` | 社員データのパフォーマンステストのみ実行 |
| `-cache-test` | キャッシュ性能テストを追加実行 |
| `-cache-only` | キャッシュ性能テストのみ実行 |
| `-benchmark-runs=10` | ベンチマーク実行回数（デフォルト: 10回） |
| `-help` | ヘルプを表示 |

### 使用例

```bash
# 受注データのみテスト、統計情報表示
go run cmd/main.go -order-only -stats

# 社員データのみテスト
go run cmd/main.go -employee-only

# キャッシュ性能テストのみ実行
go run cmd/main.go -cache-only

# 詳細情報付きで全テスト実行
go run cmd/main.go -days=7 -sample -stats
```

## 実装内容

### 1. 問題のあるアプローチ（N+1問題）

```go
func (r *ProblemOrderRepository) GetOrdersWithDetails(days int) ([]models.OrderWithDetails, error) {
    // 1. 受注一覧を取得（1回のクエリ）
    orders, err := r.GetOrdersByDays(days)
    if err != nil {
        return nil, err
    }

    var result []models.OrderWithDetails
    // 2. 各受注ごとに明細を取得（N回のクエリ）
    for _, order := range orders {
        details, err := r.GetDetailsByOrderID(order.OrderID)  // N+1問題発生！
        if err != nil {
            return nil, err
        }
        result = append(result, models.OrderWithDetails{
            Order:   order,
            Details: details,
        })
    }

    return result, nil
}
```

### 2. 最適化されたアプローチ

#### 解決策1: JOINを使用した一括取得

```go
func (r *OptimizedOrderRepository) GetOrdersWithDetailsJoin(days int) ([]models.OrderWithDetails, error) {
    query := `
        SELECT 
            o.order_id, o.customer_id, o.order_date, o.total_amount,
            od.detail_id, od.product_id, od.quantity, od.unit_price
        FROM orders o
        LEFT JOIN order_details od ON o.order_id = od.order_id
        WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
        ORDER BY o.order_id, od.detail_id`

    // 1回のクエリで全データを取得
    rows, err := r.db.Query(query, days)
    // ... 結果の処理
}
```

#### 解決策2: IN句を使用したバッチ取得

```go
func (r *OptimizedOrderRepository) GetOrdersWithDetailsBatch(days int) ([]models.OrderWithDetails, error) {
    // 1. 受注一覧を取得
    orders, err := r.GetOrdersByDays(days)
    if err != nil {
        return nil, err
    }

    // 2. 受注IDを抽出
    orderIDs := extractOrderIDs(orders)

    // 3. IN句で明細を一括取得
    allDetails, err := r.GetDetailsByOrderIDs(orderIDs)
    if err != nil {
        return nil, err
    }

    // 4. メモリ上でグルーピング
    return groupOrdersWithDetails(orders, allDetails), nil
}
```

## パフォーマンス比較

### 予想される結果

| アプローチ | SQL実行回数 | 改善率 |
| --------- | ----------- | ------ |
| N+1問題あり | N+1回 | 基準 |
| JOIN最適化 | 1回 | 10〜100倍高速 |
| バッチ取得 | 2回 | 5〜50倍高速 |

### データ量による影響

- データ量が増加するほど、N+1問題の影響は深刻になります
- 小規模データでも10倍以上の改善が期待できます
- 大規模データでは100倍以上の改善も可能です

## MySQL固有の最適化

### 1. InnoDB Buffer Poolの活用

MySQLのInnoDB Buffer Poolは、頻繁にアクセスされるデータページを自動的にキャッシュします。

```sql
-- Buffer Poolヒット率の確認
SELECT 
    (1 - (Innodb_buffer_pool_reads / Innodb_buffer_pool_read_requests)) * 100 AS hit_ratio
FROM (
    SELECT 
        VARIABLE_VALUE AS Innodb_buffer_pool_reads
    FROM performance_schema.global_status 
    WHERE VARIABLE_NAME = 'Innodb_buffer_pool_reads'
) r, (
    SELECT 
        VARIABLE_VALUE AS Innodb_buffer_pool_read_requests
    FROM performance_schema.global_status 
    WHERE VARIABLE_NAME = 'Innodb_buffer_pool_read_requests'
) rr;
```

### 2. インデックス戦略

```sql
-- 外部キー用インデックス
CREATE INDEX idx_order_details_order_id ON order_details(order_id);
CREATE INDEX idx_employees_department_id ON employees(department_id);

-- 検索条件用インデックス  
CREATE INDEX idx_orders_order_date ON orders(order_date);

-- 複合インデックス
CREATE INDEX idx_orders_date_customer ON orders(order_date, customer_id);
```

### 3. EXPLAIN ANALYZEでの実行計画確認

```sql
EXPLAIN ANALYZE
SELECT o.order_id, od.detail_id
FROM orders o
JOIN order_details od ON o.order_id = od.order_id
WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY);
```

## 開発・運用での注意点

### コードレビューでのチェック項目

```go
// ❌ 避けるべきパターン
for _, item := range items {
    relatedData := repository.GetByID(item.ID)  // ループ内DBアクセス
    // 処理...
}

// ✅ 推奨パターン  
ids := extractIDs(items)
allRelatedData := repository.GetByIDs(ids)     // 一括取得
relationMap := createRelationMap(allRelatedData)
for _, item := range items {
    relatedData := relationMap[item.ID]
    // 処理...
}
```

## ライセンス

MIT License

## 貢献

プルリクエストやイシューの報告を歓迎します。改善提案や追加機能のアイデアがあれば、お気軽にご連絡ください。
