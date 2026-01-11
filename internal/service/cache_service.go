package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"mysql-n-plus-1-demo/config"
	"mysql-n-plus-1-demo/internal/cache"
)

// CacheResult - キャッシュ性能測定結果
type CacheResult struct {
	Method        string        `json:"method"`
	ExecutionTime time.Duration `json:"execution_time"`
	MemoryUsage   int64         `json:"memory_usage_bytes"`
	HitRate       float64       `json:"hit_rate"`
	Description   string        `json:"description"`
}

// CacheService - キャッシュ性能比較サービス
type CacheService struct {
	db                  *sql.DB
	config              *config.Config
	results             []CacheResult
	performanceAnalyzer *cache.PerformanceAnalyzer
	bufferPool          *cache.InnoDBBufferPool
}

// NewCacheService - キャッシュサービスのコンストラクタ
func NewCacheService(db *sql.DB, cfg *config.Config) *CacheService {
	return &CacheService{
		db:                  db,
		config:              cfg,
		results:             make([]CacheResult, 0),
		performanceAnalyzer: cache.NewPerformanceAnalyzer(db),
		bufferPool:          cache.NewInnoDBBufferPool(db),
	}
}

// TestMySQLInternalCache - MySQL内蔵キャッシュのテスト
func (c *CacheService) TestMySQLInternalCache(runs int) error {
	fmt.Println("=== MySQL内蔵キャッシュ詳細性能テスト ===")
	fmt.Printf("詳細分析エンジンを使用（%d回実行）\n\n", runs)

	// 1. 包括的性能分析の実行
	analysisResults, err := c.performanceAnalyzer.PerformComprehensiveAnalysis(runs)
	if err != nil {
		fmt.Printf("包括的分析でエラーが発生しましたが、従来の分析を実行します: %v\n", err)
		return c.testMySQLInternalCacheFallback(runs)
	}

	// 2. 分析結果の統合
	c.integrateAnalysisResults(analysisResults)

	fmt.Println("\n✅ MySQL内蔵キャッシュ詳細分析が完了しました")

	return nil
}

// testMySQLInternalCacheFallback - 従来のMySQL内蔵キャッシュテスト（フォールバック用）
func (c *CacheService) testMySQLInternalCacheFallback(runs int) error {
	fmt.Println("=== MySQL内蔵キャッシュ基本性能テスト（フォールバック）===")

	// InnoDB Buffer Poolテスト
	if err := c.testBufferPool(runs); err != nil {
		return fmt.Errorf("buffer poolテストでエラー: %w", err)
	}

	return nil
}

// integrateAnalysisResults - 分析結果をCacheServiceに統合
func (c *CacheService) integrateAnalysisResults(results *cache.AnalysisResults) {
	// Buffer Pool結果の統合
	if results.BufferPoolMetrics != nil {
		c.results = append(c.results, CacheResult{
			Method:        "MySQL_Buffer_Pool_Advanced",
			ExecutionTime: results.BufferPoolMetrics.TestExecutionTime,
			MemoryUsage:   results.BufferPoolMetrics.PoolSize,
			HitRate:       results.BufferPoolMetrics.HitRatio,
			Description:   "MySQL InnoDB Buffer Pool（高度分析）",
		})
	}

	// 総合効率性メトリクスの統合
	if results.PerformanceComparison != nil && results.PerformanceComparison.EfficiencyMetrics != nil {
		c.results = append(c.results, CacheResult{
			Method:        "MySQL_Integrated_Cache",
			ExecutionTime: results.BufferPoolMetrics.TestExecutionTime,
			MemoryUsage:   results.BufferPoolMetrics.PoolSize,
			HitRate:       results.PerformanceComparison.EfficiencyMetrics.OverallCacheEfficiency,
			Description:   fmt.Sprintf("MySQL統合キャッシュ（総合効率%.1f%%）", results.PerformanceComparison.EfficiencyMetrics.OverallCacheEfficiency),
		})
	}
}

// testBufferPool - InnoDB Buffer Poolの性能テスト
func (c *CacheService) testBufferPool(runs int) error {
	fmt.Println("\n--- InnoDB Buffer Pool テスト ---")

	var totalDuration time.Duration
	var hitCount int64

	for i := 0; i < runs; i++ {
		start := time.Now()

		// Buffer Poolの効果を測定するクエリ
		query := `
			SELECT o.order_id, o.customer_id, o.total_amount,
			       od.detail_id, od.product_id, od.quantity
			FROM orders o
			JOIN order_details od ON o.order_id = od.order_id
			WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
			LIMIT 100`

		rows, err := c.db.Query(query)
		if err != nil {
			return fmt.Errorf("buffer poolクエリでエラー: %w", err)
		}

		var count int
		for rows.Next() {
			var orderID, customerID, detailID, productID int64
			var totalAmount float64
			var quantity int

			err := rows.Scan(&orderID, &customerID, &totalAmount, &detailID, &productID, &quantity)
			if err != nil {
				if cerr := rows.Close(); cerr != nil {
					fmt.Printf("rows.Close() failed: %v\n", cerr)
				}
				return fmt.Errorf("スキャンエラー: %w", err)
			}
			count++
		}
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}

		duration := time.Since(start)
		totalDuration += duration

		if i > 0 && duration < 50*time.Millisecond {
			hitCount++
		}

		if i == 0 {
			fmt.Printf("初回実行時間: %v (キャッシュなし)\n", duration)
		} else if i < 3 {
			fmt.Printf("%d回目実行時間: %v\n", i+1, duration)
		}
	}

	avgDuration := totalDuration / time.Duration(runs)
	hitRate := float64(hitCount) / float64(runs-1) * 100

	c.results = append(c.results, CacheResult{
		Method:        "MySQL_Buffer_Pool",
		ExecutionTime: avgDuration,
		MemoryUsage:   0,
		HitRate:       hitRate,
		Description:   "MySQL InnoDB Buffer Pool（データページキャッシュ）",
	})

	fmt.Printf("平均実行時間: %v\n", avgDuration)
	fmt.Printf("推定キャッシュヒット率: %.1f%%\n", hitRate)

	// Buffer Pool統計の取得
	if err := c.getBufferPoolStats(); err != nil {
		fmt.Printf("Buffer Pool統計取得でエラー: %v\n", err)
	}

	return nil
}

// DisplayCacheComparison - キャッシュ比較結果を表示
func (c *CacheService) DisplayCacheComparison() error {
	if len(c.results) == 0 {
		return fmt.Errorf("比較結果がありません")
	}

	fmt.Println("\n=== キャッシュ性能比較結果 ===")

	// 結果を表形式で表示
	fmt.Printf("%-30s | %-15s | %-10s | %s\n", "キャッシュ方式", "平均実行時間", "ヒット率", "説明")
	fmt.Println(strings.Repeat("-", 90))

	for _, result := range c.results {
		hitRateStr := fmt.Sprintf("%.1f%%", result.HitRate)
		if result.HitRate == 0 {
			hitRateStr = "N/A"
		}

		fmt.Printf("%-30s | %-15v | %-10s | %s\n",
			result.Method,
			result.ExecutionTime,
			hitRateStr,
			result.Description)
	}

	// 性能分析
	c.analyzePerformance()

	return nil
}

// analyzePerformance - 性能分析とMySQL内蔵キャッシュの優位性を説明
func (c *CacheService) analyzePerformance() {
	fmt.Println("\n=== 性能分析結果 ===")

	fmt.Println("\n1. MySQL InnoDB Buffer Poolの優位性:")
	fmt.Println("   ✓ データの移動が不要（メモリ効率）")
	fmt.Println("   ✓ シリアライゼーション/デシリアライゼーション不要")
	fmt.Println("   ✓ ネットワークI/Oなし")
	fmt.Println("   ✓ 自動的なキャッシュ無効化とデータ整合性")
	fmt.Println("   ✓ LRU + Midpoint Insertion戦略による効率的なメモリ管理")
	fmt.Println("   ✓ Adaptive Hash Indexによる追加の高速化")

	fmt.Println("\n2. 推奨事項:")
	fmt.Println("   → MySQL環境ではInnoDB Buffer Poolを最大限活用する")
	fmt.Println("   → 外部キャッシュは以下の場合のみ検討:")
	fmt.Println("     - マイクロサービス間でのデータ共有")
	fmt.Println("     - 外部APIからの取得データ")
	fmt.Println("     - MySQLでカバーできない計算集約的な結果")
	fmt.Println("   → N+1問題はSQL設計の改善で根本的に解決する")
}

// DisplayMemoryUsageComparison - メモリ使用量比較を表示
func (c *CacheService) DisplayMemoryUsageComparison() error {
	fmt.Println("\n=== メモリ使用量分析 ===")

	// InnoDB Buffer Pool情報の取得
	if err := c.getBufferPoolInfo(); err != nil {
		fmt.Printf("Buffer Pool情報取得でエラー: %v\n", err)
	}

	fmt.Println("\nメモリ効率性:")
	fmt.Println("MySQL InnoDB Buffer Pool:")
	fmt.Println("  ✓ 専用のメモリ領域で統合管理")
	fmt.Println("  ✓ LRU + Midpoint Insertion戦略による自動最適化")
	fmt.Println("  ✓ 複数のアプリケーションで共有")
	fmt.Println("  ✓ Adaptive Hash Indexの自動生成")

	return nil
}

// getBufferPoolStats - Buffer Pool統計を取得
func (c *CacheService) getBufferPoolStats() error {
	query := `
		SELECT 
			VARIABLE_NAME, VARIABLE_VALUE
		FROM performance_schema.global_status
		WHERE VARIABLE_NAME IN (
			'Innodb_buffer_pool_read_requests',
			'Innodb_buffer_pool_reads'
		)`

	rows, err := c.db.Query(query)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var readRequests, reads int64
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		switch name {
		case "Innodb_buffer_pool_read_requests":
			readRequests = value
		case "Innodb_buffer_pool_reads":
			reads = value
		}
	}

	if readRequests > 0 {
		hitRatio := (1.0 - float64(reads)/float64(readRequests)) * 100
		fmt.Printf("InnoDB Buffer Pool ヒット率: %.2f%%\n", hitRatio)
	}

	return nil
}

// getBufferPoolInfo - Buffer Pool情報を取得
func (c *CacheService) getBufferPoolInfo() error {
	query := `
		SELECT 
			@@innodb_buffer_pool_size / 1024 / 1024 as pool_size_mb,
			@@innodb_buffer_pool_instances as instances`

	var poolSizeMB float64
	var instances int

	err := c.db.QueryRow(query).Scan(&poolSizeMB, &instances)
	if err != nil {
		return err
	}

	fmt.Printf("\nInnoDB Buffer Pool構成:\n")
	fmt.Printf("  サイズ: %.1f MB\n", poolSizeMB)
	fmt.Printf("  インスタンス数: %d\n", instances)

	return nil
}
