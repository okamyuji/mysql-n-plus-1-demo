package service

import (
	"database/sql"
	"fmt"
	"time"

	"mysql-n-plus-1-demo/repository"
)

// PerformanceResult - パフォーマンス測定結果
type PerformanceResult struct {
	Method        string        `json:"method"`
	ExecutionTime time.Duration `json:"execution_time"`
	RecordCount   int           `json:"record_count"`
	Description   string        `json:"description"`
}

// DemoService - N+1問題のデモンストレーション用サービス
type DemoService struct {
	db               *sql.DB
	problemRepo      *repository.ProblemOrderRepository
	problemEmpRepo   *repository.ProblemEmployeeRepository
	optimizedRepo    *repository.OptimizedOrderRepository
	optimizedEmpRepo *repository.OptimizedEmployeeRepository
}

// NewDemoService - デモサービスのコンストラクタ
func NewDemoService(db *sql.DB) *DemoService {
	return &DemoService{
		db:               db,
		problemRepo:      repository.NewProblemOrderRepository(db),
		problemEmpRepo:   repository.NewProblemEmployeeRepository(db),
		optimizedRepo:    repository.NewOptimizedOrderRepository(db),
		optimizedEmpRepo: repository.NewOptimizedEmployeeRepository(db),
	}
}

// CompareOrderPerformance - 受注データの取得パフォーマンスを比較
func (s *DemoService) CompareOrderPerformance(days int) ([]PerformanceResult, error) {
	var results []PerformanceResult

	fmt.Printf("=== 受注データ取得パフォーマンス比較（過去%d日間） ===\n\n", days)

	// 1. N+1問題のあるアプローチ
	fmt.Println("1. N+1問題のあるアプローチを実行中...")
	start := time.Now()

	problemOrders, err := s.problemRepo.GetOrdersWithDetails(days)
	if err != nil {
		return nil, fmt.Errorf("N+1問題のあるアプローチでエラー: %w", err)
	}

	problemDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "N+1_Problem",
		ExecutionTime: problemDuration,
		RecordCount:   len(problemOrders),
		Description:   "N+1問題のあるアプローチ（ループ内でDBアクセス）",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", problemDuration, len(problemOrders))

	// 2. JOIN使用の最適化アプローチ
	fmt.Println("2. JOIN使用の最適化アプローチを実行中...")
	start = time.Now()

	joinOrders, err := s.optimizedRepo.GetOrdersWithDetailsJoin(days)
	if err != nil {
		return nil, fmt.Errorf("JOIN最適化アプローチでエラー: %w", err)
	}

	joinDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "JOIN_Optimized",
		ExecutionTime: joinDuration,
		RecordCount:   len(joinOrders),
		Description:   "JOIN使用の最適化アプローチ（一括取得）",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", joinDuration, len(joinOrders))

	// 3. IN句使用のバッチ取得アプローチ
	fmt.Println("3. IN句使用のバッチ取得アプローチを実行中...")
	start = time.Now()

	batchOrders, err := s.optimizedRepo.GetOrdersWithDetailsBatch(days)
	if err != nil {
		return nil, fmt.Errorf("バッチ最適化アプローチでエラー: %w", err)
	}

	batchDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "Batch_Optimized",
		ExecutionTime: batchDuration,
		RecordCount:   len(batchOrders),
		Description:   "IN句使用のバッチ取得アプローチ",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", batchDuration, len(batchOrders))

	// パフォーマンス改善率を計算して表示
	s.displayPerformanceComparison(results)

	return results, nil
}

// CompareEmployeePerformance - 社員データの取得パフォーマンスを比較
func (s *DemoService) CompareEmployeePerformance() ([]PerformanceResult, error) {
	var results []PerformanceResult

	fmt.Println("\n=== 社員データ取得パフォーマンス比較 ===")

	// 1. N+1問題のあるアプローチ
	fmt.Println("1. N+1問題のあるアプローチを実行中...")
	start := time.Now()

	problemEmployees, err := s.problemEmpRepo.GetEmployeesWithDepartment()
	if err != nil {
		return nil, fmt.Errorf("N+1問題のあるアプローチでエラー: %w", err)
	}

	problemDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "N+1_Problem",
		ExecutionTime: problemDuration,
		RecordCount:   len(problemEmployees),
		Description:   "N+1問題のあるアプローチ（ループ内でDBアクセス）",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", problemDuration, len(problemEmployees))

	// 2. JOIN使用の最適化アプローチ
	fmt.Println("2. JOIN使用の最適化アプローチを実行中...")
	start = time.Now()

	joinEmployees, err := s.optimizedEmpRepo.GetEmployeesWithDepartmentJoin()
	if err != nil {
		return nil, fmt.Errorf("JOIN最適化アプローチでエラー: %w", err)
	}

	joinDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "JOIN_Optimized",
		ExecutionTime: joinDuration,
		RecordCount:   len(joinEmployees),
		Description:   "JOIN使用の最適化アプローチ（一括取得）",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", joinDuration, len(joinEmployees))

	// 3. バッチ取得アプローチ
	fmt.Println("3. バッチ取得アプローチを実行中...")
	start = time.Now()

	batchEmployees, err := s.optimizedEmpRepo.GetEmployeesWithDepartmentBatch()
	if err != nil {
		return nil, fmt.Errorf("バッチ最適化アプローチでエラー: %w", err)
	}

	batchDuration := time.Since(start)
	results = append(results, PerformanceResult{
		Method:        "Batch_Optimized",
		ExecutionTime: batchDuration,
		RecordCount:   len(batchEmployees),
		Description:   "バッチ取得アプローチ",
	})

	fmt.Printf("   実行時間: %v, 取得件数: %d件\n", batchDuration, len(batchEmployees))

	// パフォーマンス改善率を計算して表示
	s.displayPerformanceComparison(results)

	return results, nil
}

// displayPerformanceComparison - パフォーマンス比較結果を表示
func (s *DemoService) displayPerformanceComparison(results []PerformanceResult) {
	if len(results) < 2 {
		return
	}

	fmt.Println("\n=== パフォーマンス改善効果 ===")

	baseDuration := results[0].ExecutionTime

	for i, result := range results {
		if i == 0 {
			fmt.Printf("%s: %v (基準)\n", result.Method, result.ExecutionTime)
		} else {
			improvement := float64(baseDuration.Nanoseconds()) / float64(result.ExecutionTime.Nanoseconds())
			fmt.Printf("%s: %v (%.1fx高速化)\n", result.Method, result.ExecutionTime, improvement)
		}
	}

	// 最も効果的な改善を強調表示
	if len(results) >= 2 {
		bestResult := results[1]
		bestImprovement := float64(baseDuration.Nanoseconds()) / float64(bestResult.ExecutionTime.Nanoseconds())

		for i := 2; i < len(results); i++ {
			improvement := float64(baseDuration.Nanoseconds()) / float64(results[i].ExecutionTime.Nanoseconds())
			if improvement > bestImprovement {
				bestResult = results[i]
				bestImprovement = improvement
			}
		}

		fmt.Printf("\n最も効果的な改善: %s\n", bestResult.Method)
		fmt.Printf("改善効果: %.1fx高速化 (%.2fms → %.2fms)\n",
			bestImprovement,
			float64(baseDuration.Nanoseconds())/1e6,
			float64(bestResult.ExecutionTime.Nanoseconds())/1e6)
	}
}

// DisplaySampleData - サンプルデータを表示（デバッグ用）
func (s *DemoService) DisplaySampleData(orderCount, employeeCount int) error {
	fmt.Printf("\n=== サンプルデータ表示 ===\n")

	// 受注データのサンプル表示
	if orderCount > 0 {
		fmt.Printf("\n--- 受注データサンプル（最大%d件） ---\n", orderCount)
		orders, err := s.optimizedRepo.GetOrdersWithDetailsJoin(30)
		if err != nil {
			return fmt.Errorf("受注データの取得に失敗: %w", err)
		}

		displayCount := orderCount
		if len(orders) < displayCount {
			displayCount = len(orders)
		}

		for i := 0; i < displayCount; i++ {
			order := orders[i]
			fmt.Printf("受注ID: %d, 顧客ID: %d, 日付: %s, 金額: %.2f\n",
				order.Order.OrderID, order.Order.CustomerID,
				order.Order.OrderDate.Format("2006-01-02"), order.Order.TotalAmount)

			for j, detail := range order.Details {
				if j >= 3 {
					fmt.Printf("  ... 他 %d件の明細\n", len(order.Details)-3)
					break
				}
				fmt.Printf("  明細ID: %d, 商品ID: %d, 数量: %d, 単価: %.2f\n",
					detail.DetailID, detail.ProductID, detail.Quantity, detail.UnitPrice)
			}
			fmt.Println()
		}
	}

	// 社員データのサンプル表示
	if employeeCount > 0 {
		fmt.Printf("\n--- 社員データサンプル（最大%d件） ---\n", employeeCount)
		employees, err := s.optimizedEmpRepo.GetEmployeesWithDepartmentJoin()
		if err != nil {
			return fmt.Errorf("社員データの取得に失敗: %w", err)
		}

		displayCount := employeeCount
		if len(employees) < displayCount {
			displayCount = len(employees)
		}

		for i := 0; i < displayCount; i++ {
			emp := employees[i]
			departmentInfo := "不明"
			if emp.Department != nil {
				departmentInfo = fmt.Sprintf("%s (%s)", emp.Department.DepartmentName, emp.Department.Location)
			}

			fmt.Printf("社員ID: %d, 名前: %s %s, メール: %s, 部署: %s, 給与: %.2f\n",
				emp.Employee.EmployeeID, emp.Employee.FirstName, emp.Employee.LastName,
				emp.Employee.Email, departmentInfo, emp.Employee.Salary)
		}
	}

	return nil
}

// GetDatabaseStats - データベースの統計情報を取得
func (s *DemoService) GetDatabaseStats() error {
	fmt.Printf("\n=== データベース統計情報 ===\n")

	// テーブルごとの件数を取得
	tables := []string{"orders", "order_details", "employees", "departments"}

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := s.db.QueryRow(query).Scan(&count)
		if err != nil {
			fmt.Printf("%s: エラー (%v)\n", table, err)
		} else {
			fmt.Printf("%s: %d件\n", table, count)
		}
	}

	return nil
}
