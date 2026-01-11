package cache

import (
	"database/sql"
	"fmt"
	"time"
)

// BufferPoolMetrics - InnoDB Buffer Pool性能メトリクス
type BufferPoolMetrics struct {
	HitRatio          float64       `json:"hit_ratio"`
	ReadRequests      int64         `json:"read_requests"`
	Reads             int64         `json:"reads"`
	WriteRequests     int64         `json:"write_requests"`
	PoolSize          int64         `json:"pool_size_bytes"`
	FreeBuffers       int64         `json:"free_buffers"`
	DataPages         int64         `json:"data_pages"`
	DirtyPages        int64         `json:"dirty_pages"`
	TestExecutionTime time.Duration `json:"test_execution_time"`
}

// InnoDBBufferPool - MySQL InnoDB Buffer Poolの分析実装
type InnoDBBufferPool struct {
	db      *sql.DB
	metrics *BufferPoolMetrics
}

// NewInnoDBBufferPool - Buffer Poolインスタンスを作成
func NewInnoDBBufferPool(db *sql.DB) *InnoDBBufferPool {
	return &InnoDBBufferPool{
		db:      db,
		metrics: &BufferPoolMetrics{},
	}
}

// TestBufferPoolPerformance - Buffer Poolの性能テストを実行
func (bp *InnoDBBufferPool) TestBufferPoolPerformance(runs int) (*BufferPoolMetrics, error) {
	fmt.Println("=== MySQL InnoDB Buffer Pool 詳細性能テスト ===")
	fmt.Printf("実行回数: %d回\n\n", runs)

	// 初期メトリクス取得
	initialMetrics, err := bp.collectMetrics()
	if err != nil {
		return nil, fmt.Errorf("初期メトリクス取得エラー: %w", err)
	}

	fmt.Println("1. Buffer Pool初期状態:")
	bp.displayMetrics(initialMetrics)

	var totalDuration time.Duration

	// 複数回のテスト実行
	for i := 0; i < runs; i++ {
		start := time.Now()

		// Buffer Poolの効果を測定するためのクエリ
		if err := bp.executeBufferPoolTest(i == 0); err != nil {
			return nil, fmt.Errorf("buffer Poolテスト実行エラー: %w", err)
		}

		duration := time.Since(start)
		totalDuration += duration

		if i < 3 {
			fmt.Printf("%d回目実行時間: %v\n", i+1, duration)
		}
	}

	// 最終メトリクス取得
	finalMetrics, err := bp.collectMetrics()
	if err != nil {
		return nil, fmt.Errorf("最終メトリクス取得エラー: %w", err)
	}

	// 差分計算
	bp.metrics = bp.calculateDifferential(initialMetrics, finalMetrics)
	bp.metrics.TestExecutionTime = totalDuration / time.Duration(runs)

	fmt.Println("\n2. Buffer Pool最終状態:")
	bp.displayMetrics(finalMetrics)

	fmt.Println("\n3. Buffer Poolテスト期間中の差分メトリクス:")
	bp.displayDifferentialMetrics()

	return bp.metrics, nil
}

// executeBufferPoolTest - Buffer Poolテスト用クエリを実行
func (bp *InnoDBBufferPool) executeBufferPoolTest(isFirstRun bool) error {
	queries := []string{
		// 1. フルテーブルスキャン（Buffer Pool効果測定）
		`SELECT COUNT(*) FROM orders WHERE order_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)`,

		// 2. 同じデータに対する複数回アクセス
		`SELECT o.order_id, o.customer_id, o.total_amount
		 FROM orders o 
		 WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		 ORDER BY o.order_id`,

		// 3. JOINによる複数テーブルアクセス
		`SELECT o.order_id, od.detail_id, od.quantity
		 FROM orders o
		 JOIN order_details od ON o.order_id = od.order_id
		 WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		 LIMIT 1000`,

		// 4. インデックスを使用したアクセス
		`SELECT e.employee_id, e.first_name, e.last_name
		 FROM employees e
		 WHERE e.department_id IN (1, 2, 3)`,
	}

	for i, query := range queries {
		if isFirstRun {
			fmt.Printf("実行中: クエリ%d\n", i+1)
		}

		rows, err := bp.db.Query(query)
		if err != nil {
			return fmt.Errorf("クエリ%d実行エラー: %w", i+1, err)
		}

		// 結果をすべて読み取り
		for rows.Next() {
			var dummy interface{}
			_ = rows.Scan(&dummy)
		}
		if err := rows.Close(); err != nil {
			fmt.Printf("rows.Close error: %v\n", err)
		}
	}

	return nil
}

// collectMetrics - Buffer Poolメトリクスを収集
func (bp *InnoDBBufferPool) collectMetrics() (*BufferPoolMetrics, error) {
	metrics := &BufferPoolMetrics{}

	// InnoDB Buffer Pool統計を取得
	query := `
		SELECT 
			VARIABLE_NAME, VARIABLE_VALUE
		FROM performance_schema.global_status
		WHERE VARIABLE_NAME IN (
			'Innodb_buffer_pool_read_requests',
			'Innodb_buffer_pool_reads',
			'Innodb_buffer_pool_write_requests',
			'Innodb_buffer_pool_pages_total',
			'Innodb_buffer_pool_pages_free',
			'Innodb_buffer_pool_pages_data',
			'Innodb_buffer_pool_pages_dirty'
		)`

	rows, err := bp.db.Query(query)
	if err != nil {
		// performance_schemaが無効な場合は簡易メトリクスを返す
		fmt.Printf("performance_schema統計取得エラー（スキップ）: %v\n", err)
		return metrics, nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("rows.Close error: %v\n", err)
		}
	}()

	for rows.Next() {
		var name string
		var value int64

		if err := rows.Scan(&name, &value); err != nil {
			continue
		}

		switch name {
		case "Innodb_buffer_pool_read_requests":
			metrics.ReadRequests = value
		case "Innodb_buffer_pool_reads":
			metrics.Reads = value
		case "Innodb_buffer_pool_write_requests":
			metrics.WriteRequests = value
		case "Innodb_buffer_pool_pages_total":
			metrics.PoolSize = value * 16384 // 16KB per page
		case "Innodb_buffer_pool_pages_free":
			metrics.FreeBuffers = value
		case "Innodb_buffer_pool_pages_data":
			metrics.DataPages = value
		case "Innodb_buffer_pool_pages_dirty":
			metrics.DirtyPages = value
		}
	}

	// ヒット率を計算
	if metrics.ReadRequests > 0 {
		metrics.HitRatio = (1.0 - float64(metrics.Reads)/float64(metrics.ReadRequests)) * 100
	}

	return metrics, nil
}

// calculateDifferential - メトリクスの差分を計算
func (bp *InnoDBBufferPool) calculateDifferential(initial, final *BufferPoolMetrics) *BufferPoolMetrics {
	diff := &BufferPoolMetrics{
		ReadRequests:  final.ReadRequests - initial.ReadRequests,
		Reads:         final.Reads - initial.Reads,
		WriteRequests: final.WriteRequests - initial.WriteRequests,
		PoolSize:      final.PoolSize,
		FreeBuffers:   final.FreeBuffers,
		DataPages:     final.DataPages,
		DirtyPages:    final.DirtyPages,
	}

	if diff.ReadRequests > 0 {
		diff.HitRatio = (1.0 - float64(diff.Reads)/float64(diff.ReadRequests)) * 100
	} else {
		diff.HitRatio = final.HitRatio
	}

	return diff
}

// displayMetrics - メトリクスを表示
func (bp *InnoDBBufferPool) displayMetrics(metrics *BufferPoolMetrics) {
	fmt.Printf("  Buffer Pool ヒット率: %.2f%%\n", metrics.HitRatio)
	fmt.Printf("  読み取りリクエスト: %d\n", metrics.ReadRequests)
	fmt.Printf("  ディスク読み取り: %d\n", metrics.Reads)
	fmt.Printf("  書き込みリクエスト: %d\n", metrics.WriteRequests)
	fmt.Printf("  データページ数: %d\n", metrics.DataPages)
	fmt.Printf("  空きバッファ数: %d\n", metrics.FreeBuffers)
	fmt.Printf("  ダーティページ数: %d\n", metrics.DirtyPages)
	if metrics.PoolSize > 0 {
		fmt.Printf("  Buffer Poolサイズ: %.2f MB\n", float64(metrics.PoolSize)/(1024*1024))
	}
}

// displayDifferentialMetrics - 差分メトリクスを表示
func (bp *InnoDBBufferPool) displayDifferentialMetrics() {
	fmt.Printf("  テスト期間中のヒット率: %.2f%%\n", bp.metrics.HitRatio)
	fmt.Printf("  追加読み取りリクエスト: %d\n", bp.metrics.ReadRequests)
	fmt.Printf("  追加ディスク読み取り: %d\n", bp.metrics.Reads)
	fmt.Printf("  平均実行時間: %v\n", bp.metrics.TestExecutionTime)

	// 効率指標の計算
	if bp.metrics.ReadRequests > 0 {
		efficiency := (float64(bp.metrics.ReadRequests-bp.metrics.Reads) / float64(bp.metrics.ReadRequests)) * 100
		fmt.Printf("  Buffer Pool効率: %.2f%% (キャッシュから提供された割合)\n", efficiency)
	}
}

// GetOptimizationRecommendations - Buffer Pool最適化推奨事項を取得
func (bp *InnoDBBufferPool) GetOptimizationRecommendations() []string {
	recommendations := []string{
		"✓ MySQL InnoDB Buffer PoolはLRUアルゴリズムで自動管理される",
		"✓ 頻繁にアクセスされるデータページは自動的にメモリに保持される",
		"✓ 外部キャッシュと異なり、データの整合性が自動的に保証される",
		"✓ 複数のセッションでBuffer Poolを共有するため、メモリ効率が高い",
	}

	// メトリクスに基づく動的な推奨事項
	if bp.metrics != nil {
		if bp.metrics.HitRatio < 90 {
			recommendations = append(recommendations,
				"⚠ Buffer Poolヒット率が低い（90%未満）- innodb_buffer_pool_sizeの増加を検討",
			)
		}

		if bp.metrics.HitRatio >= 95 {
			recommendations = append(recommendations,
				"✓ 優秀なBuffer Poolヒット率（95%以上）- 現在の設定は適切",
			)
		}
	}

	return recommendations
}

// GetMetrics - 現在のメトリクスを取得
func (bp *InnoDBBufferPool) GetMetrics() *BufferPoolMetrics {
	return bp.metrics
}
