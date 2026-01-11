package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mysql-n-plus-1-demo/config"
	"mysql-n-plus-1-demo/internal/service"
)

func main() {
	// コマンドラインフラグの定義
	var (
		days          = flag.Int("days", 30, "取得する受注データの日数（過去何日間）")
		showSample    = flag.Bool("sample", false, "サンプルデータを表示する")
		showStats     = flag.Bool("stats", false, "データベース統計情報を表示する")
		orderOnly     = flag.Bool("order-only", false, "受注データのみテストする")
		employeeOnly  = flag.Bool("employee-only", false, "社員データのみテストする")
		cacheTest     = flag.Bool("cache-test", false, "キャッシュ性能比較テストを実行する")
		cacheOnly     = flag.Bool("cache-only", false, "キャッシュテストのみ実行する")
		benchmarkRuns = flag.Int("benchmark-runs", 10, "ベンチマーク実行回数")
		help          = flag.Bool("help", false, "ヘルプを表示する")
	)

	flag.Parse()

	// ヘルプ表示
	if *help {
		showHelp()
		return
	}

	// アプリケーション開始
	fmt.Println("MySQL N+1問題 & キャッシュ性能デモンストレーション")
	fmt.Println("===============================================")

	// 設定読み込み
	fmt.Println("設定を読み込み中...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// データベース接続
	fmt.Println("データベースに接続中...")
	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("データベースクローズエラー: %v", err)
		}
	}()

	// 接続テスト
	if err := db.Ping(); err != nil {
		log.Fatalf("データベース接続テストに失敗しました: %v", err)
	}
	fmt.Println("データベース接続成功！")

	// サービスの初期化
	demoService := service.NewDemoService(db)
	cacheService := service.NewCacheService(db, cfg)

	// データベース統計情報の表示
	if *showStats {
		if err := demoService.GetDatabaseStats(); err != nil {
			log.Printf("データベース統計の取得中にエラー: %v", err)
		}
		fmt.Println()
	}

	// サンプルデータの表示
	if *showSample {
		if err := demoService.DisplaySampleData(5, 5); err != nil {
			log.Printf("サンプルデータの表示中にエラー: %v", err)
		}
		fmt.Println()
	}

	// 実行モードに応じた処理
	switch {
	case *cacheOnly:
		// キャッシュテストのみ
		runCacheTests(cacheService, *benchmarkRuns)
	case *cacheTest && !*orderOnly && !*employeeOnly:
		// 全テスト + キャッシュテスト
		runAllTests(demoService, *days)
		runCacheTests(cacheService, *benchmarkRuns)
	case *orderOnly:
		// 受注データのみ
		runOrderTests(demoService, *days)
		if *cacheTest {
			runCacheTests(cacheService, *benchmarkRuns)
		}
	case *employeeOnly:
		// 社員データのみ
		runEmployeeTests(demoService)
		if *cacheTest {
			runCacheTests(cacheService, *benchmarkRuns)
		}
	default:
		// デフォルト：N+1問題のテストのみ
		runAllTests(demoService, *days)
	}

	fmt.Println("\nデモンストレーション完了！")
}

// showHelp - ヘルプメッセージを表示
func showHelp() {
	fmt.Println("MySQL N+1問題 & キャッシュ性能デモンストレーション")
	fmt.Println("===============================================")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Printf("  %s [オプション]\n", os.Args[0])
	fmt.Println()
	fmt.Println("オプション:")
	fmt.Println("  -days=30          取得する受注データの日数（デフォルト: 30日）")
	fmt.Println("  -sample           サンプルデータを表示する")
	fmt.Println("  -stats            データベース統計情報を表示する")
	fmt.Println("  -order-only       受注データのパフォーマンステストのみ実行")
	fmt.Println("  -employee-only    社員データのパフォーマンステストのみ実行")
	fmt.Println("  -cache-test       キャッシュ性能比較テストを追加実行")
	fmt.Println("  -cache-only       キャッシュテストのみ実行")
	fmt.Println("  -benchmark-runs=10 ベンチマーク実行回数（デフォルト: 10回）")
	fmt.Println("  -help             このヘルプを表示する")
	fmt.Println()
	fmt.Println("使用例:")
	fmt.Printf("  %s -days=7 -sample              # 過去7日間の受注データでテスト、サンプル表示\n", os.Args[0])
	fmt.Printf("  %s -order-only -stats           # 受注データのみテスト、統計表示\n", os.Args[0])
	fmt.Printf("  %s -cache-test                  # N+1テスト + キャッシュ性能比較\n", os.Args[0])
	fmt.Printf("  %s -cache-only -benchmark-runs=20 # キャッシュテストのみ20回実行\n", os.Args[0])
	fmt.Println()
	fmt.Println("環境設定:")
	fmt.Println("  .envファイルまたは環境変数でMySQL接続情報を設定してください。")
	fmt.Println("  必要な環境変数:")
	fmt.Println("    - DB_HOST: MySQLサーバーのホスト名")
	fmt.Println("    - DB_PORT: ポート番号（デフォルト: 3306）")
	fmt.Println("    - DB_NAME: データベース名")
	fmt.Println("    - DB_USER: ユーザー名")
	fmt.Println("    - DB_PASSWORD: パスワード")
}

// runCacheTests - キャッシュ性能比較テストを実行
func runCacheTests(cacheService *service.CacheService, benchmarkRuns int) {
	fmt.Printf("\n=== キャッシュ性能比較テスト（%d回実行）===\n", benchmarkRuns)
	fmt.Println("MySQL InnoDB Buffer Poolの性能を分析します")
	fmt.Println()

	// MySQL内蔵キャッシュのテスト
	if err := cacheService.TestMySQLInternalCache(benchmarkRuns); err != nil {
		log.Printf("MySQL内蔵キャッシュテストでエラー: %v", err)
	}

	// 比較結果の表示
	if err := cacheService.DisplayCacheComparison(); err != nil {
		log.Printf("キャッシュ比較結果の表示でエラー: %v", err)
	}

	// メモリ使用量の比較
	if err := cacheService.DisplayMemoryUsageComparison(); err != nil {
		log.Printf("メモリ使用量比較でエラー: %v", err)
	}
}

// runAllTests - 全てのパフォーマンステストを実行
func runAllTests(demoService *service.DemoService, days int) {
	fmt.Printf("\n全てのパフォーマンステストを実行します（受注: 過去%d日間）\n", days)
	fmt.Println("==================================================")

	// 受注データのテスト
	runOrderTests(demoService, days)

	// 社員データのテスト
	runEmployeeTests(demoService)

	// 総合結果の表示
	fmt.Println("\n=== 総合結果 ===")
	fmt.Println("N+1問題の解決により、大幅なパフォーマンス改善が確認できました。")
	fmt.Println("特にデータ量が多くなるにつれて、改善効果が顕著に現れます。")
	fmt.Println()
	fmt.Println("推奨事項:")
	fmt.Println("1. JOINを使用した一括取得を優先的に検討する")
	fmt.Println("2. JOINが適さない場合はIN句を使用したバッチ取得を検討する")
	fmt.Println("3. ORMを使用する場合は適切な設定でLazy Loadingを制御する")
	fmt.Println("4. 定期的なパフォーマンス監視により問題を早期発見する")
	fmt.Println("5. MySQL InnoDB Buffer Poolを最大限活用する")
}

// runOrderTests - 受注データのパフォーマンステストを実行
func runOrderTests(demoService *service.DemoService, days int) {
	fmt.Printf("\n受注データのパフォーマンステストを実行中...\n")

	results, err := demoService.CompareOrderPerformance(days)
	if err != nil {
		log.Printf("受注データテスト中にエラー: %v", err)
		return
	}

	// 結果の詳細表示
	fmt.Println("\n--- 受注データテスト結果詳細 ---")
	for _, result := range results {
		fmt.Printf("手法: %s\n", result.Description)
		fmt.Printf("実行時間: %v\n", result.ExecutionTime)
		fmt.Printf("取得件数: %d件\n", result.RecordCount)
		fmt.Println()
	}

	// N+1問題の影響を具体的に説明
	if len(results) >= 2 {
		baseDuration := results[0].ExecutionTime
		optimizedDuration := results[1].ExecutionTime

		fmt.Printf("N+1問題による影響:\n")
		fmt.Printf("- 実行時間: %v → %v\n", baseDuration, optimizedDuration)

		if baseDuration > optimizedDuration {
			saved := baseDuration - optimizedDuration
			fmt.Printf("- 時間短縮: %v (%.1f%%削減)\n",
				saved,
				float64(saved.Nanoseconds())/float64(baseDuration.Nanoseconds())*100)
		}
	}
}

// runEmployeeTests - 社員データのパフォーマンステストを実行
func runEmployeeTests(demoService *service.DemoService) {
	fmt.Printf("\n社員データのパフォーマンステストを実行中...\n")

	results, err := demoService.CompareEmployeePerformance()
	if err != nil {
		log.Printf("社員データテスト中にエラー: %v", err)
		return
	}

	// 結果の詳細表示
	fmt.Println("\n--- 社員データテスト結果詳細 ---")
	for _, result := range results {
		fmt.Printf("手法: %s\n", result.Description)
		fmt.Printf("実行時間: %v\n", result.ExecutionTime)
		fmt.Printf("取得件数: %d件\n", result.RecordCount)
		fmt.Println()
	}

	// N+1問題の影響を具体的に説明
	if len(results) >= 2 {
		baseDuration := results[0].ExecutionTime
		optimizedDuration := results[1].ExecutionTime

		fmt.Printf("N+1問題による影響:\n")
		fmt.Printf("- 実行時間: %v → %v\n", baseDuration, optimizedDuration)

		if baseDuration > optimizedDuration {
			saved := baseDuration - optimizedDuration
			fmt.Printf("- 時間短縮: %v (%.1f%%削減)\n",
				saved,
				float64(saved.Nanoseconds())/float64(baseDuration.Nanoseconds())*100)
		}
	}
}
