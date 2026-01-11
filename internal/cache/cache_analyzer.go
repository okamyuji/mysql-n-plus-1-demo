package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// PerformanceAnalyzer - キャッシュ性能分析ユーティリティ
type PerformanceAnalyzer struct {
	db              *sql.DB
	bufferPool      *InnoDBBufferPool
	analysisResults *AnalysisResults
}

// AnalysisResults - 統合分析結果
type AnalysisResults struct {
	TestDate              time.Time              `json:"test_date"`
	TestDuration          time.Duration          `json:"test_duration"`
	BufferPoolMetrics     *BufferPoolMetrics     `json:"buffer_pool_metrics"`
	PerformanceComparison *PerformanceComparison `json:"performance_comparison"`
	OptimizationAdvice    *OptimizationAdvice    `json:"optimization_advice"`
	Recommendations       []Recommendation       `json:"recommendations"`
}

// PerformanceComparison - 性能比較結果
type PerformanceComparison struct {
	MySQLAdvantages     []string             `json:"mysql_advantages"`
	ExternalCacheIssues []string             `json:"external_cache_issues"`
	EfficiencyMetrics   *EfficiencyMetrics   `json:"efficiency_metrics"`
	ResourceUtilization *ResourceUtilization `json:"resource_utilization"`
}

// EfficiencyMetrics - 効率性メトリクス
type EfficiencyMetrics struct {
	BufferPoolEfficiency   float64 `json:"buffer_pool_efficiency"`
	OverallCacheEfficiency float64 `json:"overall_cache_efficiency"`
	MemoryEfficiencyRatio  float64 `json:"memory_efficiency_ratio"`
	IOReductionRatio       float64 `json:"io_reduction_ratio"`
}

// ResourceUtilization - リソース使用率
type ResourceUtilization struct {
	TotalCacheMemoryMB    float64 `json:"total_cache_memory_mb"`
	BufferPoolUtilization float64 `json:"buffer_pool_utilization"`
	EstimatedIOSavings    int64   `json:"estimated_io_savings"`
}

// OptimizationAdvice - 最適化アドバイス
type OptimizationAdvice struct {
	PriorityLevel       string   `json:"priority_level"`
	ImmediateActions    []string `json:"immediate_actions"`
	MediumTermActions   []string `json:"medium_term_actions"`
	LongTermActions     []string `json:"long_term_actions"`
	AvoidanceStrategies []string `json:"avoidance_strategies"`
}

// Recommendation - 推奨事項
type Recommendation struct {
	Category    string   `json:"category"`
	Priority    string   `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
	Benefits    []string `json:"benefits"`
}

// NewPerformanceAnalyzer - 性能分析器を作成
func NewPerformanceAnalyzer(db *sql.DB) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		db:         db,
		bufferPool: NewInnoDBBufferPool(db),
	}
}

// PerformComprehensiveAnalysis - 包括的なキャッシュ性能分析を実行
func (pa *PerformanceAnalyzer) PerformComprehensiveAnalysis(runs int) (*AnalysisResults, error) {
	fmt.Println("\n=== MySQL内蔵キャッシュ包括的性能分析 ===")
	fmt.Printf("実行回数: %d回\n", runs)
	fmt.Println("分析項目: InnoDB Buffer Pool, 統合効率性, 外部キャッシュ比較")
	fmt.Println(strings.Repeat("=", 70))

	startTime := time.Now()

	// 1. Buffer Poolの詳細分析
	fmt.Println("\n1. InnoDB Buffer Pool分析中...")
	bufferMetrics, err := pa.bufferPool.TestBufferPoolPerformance(runs)
	if err != nil {
		return nil, fmt.Errorf("buffer pool分析エラー: %w", err)
	}

	// 2. 統合分析の実行
	fmt.Println("\n2. 統合性能分析中...")
	analysisResults := &AnalysisResults{
		TestDate:          startTime,
		TestDuration:      time.Since(startTime),
		BufferPoolMetrics: bufferMetrics,
	}

	// 3. 性能比較の計算
	if err := pa.calculatePerformanceComparison(analysisResults); err != nil {
		return nil, fmt.Errorf("性能比較計算エラー: %w", err)
	}

	// 4. 最適化アドバイスの生成
	if err := pa.generateOptimizationAdvice(analysisResults); err != nil {
		return nil, fmt.Errorf("最適化アドバイス生成エラー: %w", err)
	}

	// 5. 推奨事項の生成
	if err := pa.generateRecommendations(analysisResults); err != nil {
		return nil, fmt.Errorf("推奨事項生成エラー: %w", err)
	}

	pa.analysisResults = analysisResults

	// 6. 最終結果の表示
	pa.DisplayComprehensiveResults()

	return analysisResults, nil
}

// calculatePerformanceComparison - 性能比較を計算
func (pa *PerformanceAnalyzer) calculatePerformanceComparison(results *AnalysisResults) error {
	// 効率性メトリクスの計算
	efficiency := &EfficiencyMetrics{}

	// Buffer Pool効率性
	efficiency.BufferPoolEfficiency = results.BufferPoolMetrics.HitRatio
	efficiency.OverallCacheEfficiency = efficiency.BufferPoolEfficiency

	// メモリ効率比
	totalMemory := float64(results.BufferPoolMetrics.PoolSize)
	if totalMemory > 0 {
		efficiency.MemoryEfficiencyRatio = efficiency.OverallCacheEfficiency / (totalMemory / (1024 * 1024 * 1024))
	}

	// I/O削減比
	if results.BufferPoolMetrics.ReadRequests > 0 {
		efficiency.IOReductionRatio = (float64(results.BufferPoolMetrics.Reads) / float64(results.BufferPoolMetrics.ReadRequests)) * 100
	}

	// リソース使用率の計算
	utilization := &ResourceUtilization{
		TotalCacheMemoryMB:    totalMemory / (1024 * 1024),
		BufferPoolUtilization: float64(results.BufferPoolMetrics.DataPages) / float64(results.BufferPoolMetrics.DataPages+results.BufferPoolMetrics.FreeBuffers) * 100,
		EstimatedIOSavings:    results.BufferPoolMetrics.ReadRequests - results.BufferPoolMetrics.Reads,
	}

	// MySQL内蔵キャッシュの優位性
	mysqlAdvantages := []string{
		"✓ データページレベルの自動キャッシュ管理（LRU + Midpoint Insertion）",
		"✓ ACID特性による完全なデータ整合性保証",
		"✓ 自動的なダーティページのフラッシュ管理",
		"✓ 複数プロセス間での効率的なキャッシュ共有",
		"✓ ネットワークI/O削除による低レイテンシ",
		"✓ シリアライゼーションオーバーヘッドの排除",
		"✓ Adaptive Hash Indexによる追加の高速化",
	}

	// 外部キャッシュの課題
	externalCacheIssues := []string{
		"✗ ネットワーク通信による追加レイテンシ（平均1-5ms）",
		"✗ JSONシリアライゼーション/デシリアライゼーションのCPU負荷",
		"✗ データ整合性の手動管理による複雑性とバグリスク",
		"✗ キャッシュ無効化ロジックの実装・メンテナンス負荷",
		"✗ 追加インフラストラクチャの運用コスト",
		"✗ メモリの二重使用による非効率性",
		"✗ 外部システム障害時の影響範囲拡大",
	}

	results.PerformanceComparison = &PerformanceComparison{
		MySQLAdvantages:     mysqlAdvantages,
		ExternalCacheIssues: externalCacheIssues,
		EfficiencyMetrics:   efficiency,
		ResourceUtilization: utilization,
	}

	return nil
}

// generateOptimizationAdvice - 最適化アドバイスを生成
func (pa *PerformanceAnalyzer) generateOptimizationAdvice(results *AnalysisResults) error {
	advice := &OptimizationAdvice{}

	// 優先度の決定
	if results.PerformanceComparison.EfficiencyMetrics.OverallCacheEfficiency >= 90 {
		advice.PriorityLevel = "最適化済み（維持管理）"
	} else if results.PerformanceComparison.EfficiencyMetrics.OverallCacheEfficiency >= 70 {
		advice.PriorityLevel = "中優先度（改善推奨）"
	} else {
		advice.PriorityLevel = "高優先度（早急な対応必要）"
	}

	// 即座の対応事項
	advice.ImmediateActions = []string{
		"MySQL InnoDB Buffer Poolの活用を最優先に検討",
		"外部キャッシュ導入前にN+1問題の根本的解決（SQL最適化）を実施",
		"EXPLAIN ANALYZEでクエリ実行計画を確認",
		"Buffer Pool統計の定期監視を開始",
	}

	// 中期的対応事項
	advice.MediumTermActions = []string{
		"アプリケーション層のキャッシュロジック簡素化",
		"インデックス設計の最適化",
		"クエリキャッシュ代替としてのアプリケーションレベルキャッシュ検討",
		"外部キャッシュ依存度の段階的削減",
	}

	// 長期的対応事項
	advice.LongTermActions = []string{
		"データベース中心設計への移行",
		"外部キャッシュインフラの段階的廃止検討",
		"MySQL統合監視体制の構築",
		"チーム全体のMySQL最適化スキル向上",
	}

	// 回避すべき戦略
	advice.AvoidanceStrategies = []string{
		"単純な外部キャッシュ追加による問題の先送り",
		"N+1問題の根本解決を避けた対症療法",
		"MySQL機能を活用しない独自キャッシュ実装",
		"データ整合性を犠牲にした性能最適化",
	}

	results.OptimizationAdvice = advice
	return nil
}

// generateRecommendations - 推奨事項を生成
func (pa *PerformanceAnalyzer) generateRecommendations(results *AnalysisResults) error {
	var recommendations []Recommendation

	// Buffer Pool関連推奨事項
	if results.BufferPoolMetrics.HitRatio < 90 {
		recommendations = append(recommendations, Recommendation{
			Category:    "Buffer Pool最適化",
			Priority:    "高",
			Title:       "Buffer Poolサイズの最適化",
			Description: "Buffer Poolヒット率が90%未満のため、サイズ調整を検討",
			Impact:      "I/O削減による大幅な性能向上",
			Effort:      "低（パラメータ調整のみ）",
			Benefits:    []string{"物理I/O削減", "レスポンス時間向上", "CPU使用率改善"},
		})
	}

	// 外部キャッシュ見直し推奨事項
	recommendations = append(recommendations, Recommendation{
		Category:    "アーキテクチャ最適化",
		Priority:    "高",
		Title:       "外部キャッシュ依存度の削減",
		Description: "MySQL内蔵キャッシュを最大限活用し、外部キャッシュ依存を削減",
		Impact:      "システム複雑性の削減と運用性向上",
		Effort:      "高（アーキテクチャ変更）",
		Benefits:    []string{"システム複雑性削減", "運用コスト削減", "データ整合性向上", "レイテンシ削減"},
	})

	// N+1問題根本解決推奨事項
	recommendations = append(recommendations, Recommendation{
		Category:    "SQL最適化",
		Priority:    "最高",
		Title:       "N+1問題の根本的解決",
		Description: "JOINやIN句を使用してN+1問題を根本から解決",
		Impact:      "クエリ実行回数の劇的削減",
		Effort:      "中（SQL設計見直し）",
		Benefits:    []string{"実行時間短縮", "データベース負荷軽減", "スケーラビリティ向上"},
	})

	// 優先度でソート
	sort.Slice(recommendations, func(i, j int) bool {
		priorities := map[string]int{"最高": 4, "高": 3, "中": 2, "低": 1}
		return priorities[recommendations[i].Priority] > priorities[recommendations[j].Priority]
	})

	results.Recommendations = recommendations
	return nil
}

// DisplayComprehensiveResults - 包括的結果を表示
func (pa *PerformanceAnalyzer) DisplayComprehensiveResults() {
	if pa.analysisResults == nil {
		fmt.Println("分析結果がありません")
		return
	}

	results := pa.analysisResults

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("MySQL内蔵キャッシュ vs 外部キャッシュ 包括的分析結果")
	fmt.Println(strings.Repeat("=", 80))

	// 1. エグゼクティブサマリー
	pa.displayExecutiveSummary(results)

	// 2. 性能メトリクス比較
	pa.displayPerformanceMetrics(results)

	// 3. 推奨事項
	pa.displayRecommendations(results)

	// 4. 結論
	pa.displayConclusion(results)
}

// displayExecutiveSummary - エグゼクティブサマリーを表示
func (pa *PerformanceAnalyzer) displayExecutiveSummary(results *AnalysisResults) {
	fmt.Println("\n■ エグゼクティブサマリー")
	fmt.Println(strings.Repeat("-", 50))

	efficiency := results.PerformanceComparison.EfficiencyMetrics.OverallCacheEfficiency

	if efficiency >= 90 {
		fmt.Println("✅ 総合評価: 優秀（90%以上の効率性）")
		fmt.Println("   MySQL InnoDB Buffer Poolが効果的に機能しています")
	} else if efficiency >= 70 {
		fmt.Println("⚠️  総合評価: 良好（70-90%の効率性）")
		fmt.Println("   改善の余地がありますが、基本的な機能は正常です")
	} else {
		fmt.Println("❌ 総合評価: 要改善（70%未満の効率性）")
		fmt.Println("   早急な最適化が必要です")
	}

	fmt.Printf("\n• Buffer Pool効率: %.1f%%\n", efficiency)
	fmt.Printf("• 総キャッシュメモリ: %.1f MB\n", results.PerformanceComparison.ResourceUtilization.TotalCacheMemoryMB)
	fmt.Printf("• 推定I/O削減: %d回\n", results.PerformanceComparison.ResourceUtilization.EstimatedIOSavings)
}

// displayPerformanceMetrics - 性能メトリクス比較を表示
func (pa *PerformanceAnalyzer) displayPerformanceMetrics(results *AnalysisResults) {
	fmt.Println("\n■ MySQL内蔵キャッシュ vs 外部キャッシュ 比較")
	fmt.Println(strings.Repeat("-", 50))

	fmt.Println("\n✅ MySQL InnoDB Buffer Poolの優位性:")
	for _, advantage := range results.PerformanceComparison.MySQLAdvantages {
		fmt.Printf("   %s\n", advantage)
	}

	fmt.Println("\n❌ 外部キャッシュの課題:")
	for _, issue := range results.PerformanceComparison.ExternalCacheIssues {
		fmt.Printf("   %s\n", issue)
	}
}

// displayRecommendations - 推奨事項を表示
func (pa *PerformanceAnalyzer) displayRecommendations(results *AnalysisResults) {
	fmt.Println("\n■ 推奨事項（優先度順）")
	fmt.Println(strings.Repeat("-", 50))

	for i, rec := range results.Recommendations {
		fmt.Printf("\n%d. [%s優先度] %s\n", i+1, rec.Priority, rec.Title)
		fmt.Printf("   カテゴリ: %s\n", rec.Category)
		fmt.Printf("   説明: %s\n", rec.Description)
		fmt.Printf("   期待効果: %s\n", rec.Impact)
		if len(rec.Benefits) > 0 {
			fmt.Printf("   利益: %s\n", strings.Join(rec.Benefits, ", "))
		}
	}
}

// displayConclusion - 結論を表示
func (pa *PerformanceAnalyzer) displayConclusion(results *AnalysisResults) {
	fmt.Println("\n■ 結論とNext Steps")
	fmt.Println(strings.Repeat("-", 50))

	fmt.Println("\n🎯 重要な結論:")
	fmt.Println("   1. MySQL InnoDB Buffer Poolは外部キャッシュよりも効率的")
	fmt.Println("   2. N+1問題は根本的なSQL設計で解決すべき")
	fmt.Println("   3. 外部キャッシュは複雑性を増加させ運用コストを高める")
	fmt.Println("   4. データ整合性はMySQLの自動機能に任せるべき")

	fmt.Printf("\n📊 今回の分析結果: %s\n", results.OptimizationAdvice.PriorityLevel)

	fmt.Println("\n🚀 Next Steps:")
	if len(results.OptimizationAdvice.ImmediateActions) > 0 {
		fmt.Println("   即座の対応:")
		for _, action := range results.OptimizationAdvice.ImmediateActions {
			fmt.Printf("     • %s\n", action)
		}
	}

	fmt.Println("\n💡 長期的な方向性:")
	fmt.Println("   → MySQL Database中心のアーキテクチャ採用")
	fmt.Println("   → 外部キャッシュ依存度の段階的削減")
	fmt.Println("   → SQL最適化による根本的問題解決")
}

// ExportAnalysisResults - 分析結果をJSONでエクスポート
func (pa *PerformanceAnalyzer) ExportAnalysisResults() (string, error) {
	if pa.analysisResults == nil {
		return "", fmt.Errorf("分析結果がありません")
	}

	jsonData, err := json.MarshalIndent(pa.analysisResults, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON変換エラー: %w", err)
	}

	return string(jsonData), nil
}

// GetAnalysisResults - 分析結果を取得
func (pa *PerformanceAnalyzer) GetAnalysisResults() *AnalysisResults {
	return pa.analysisResults
}
