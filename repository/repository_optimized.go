package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"mysql-n-plus-1-demo/models"
)

// OptimizedOrderRepository - N+1問題を解決したリポジトリ
type OptimizedOrderRepository struct {
	db *sql.DB
}

// NewOptimizedOrderRepository - 最適化されたリポジトリのコンストラクタ
func NewOptimizedOrderRepository(db *sql.DB) *OptimizedOrderRepository {
	return &OptimizedOrderRepository{db: db}
}

// GetOrdersWithDetailsJoin - JOINを使用した一括取得（推奨方法1）
func (r *OptimizedOrderRepository) GetOrdersWithDetailsJoin(days int) ([]models.OrderWithDetails, error) {
	query := `
		SELECT 
			o.order_id,
			o.customer_id,
			o.order_date,
			o.total_amount,
			od.detail_id,
			od.product_id,
			od.quantity,
			od.unit_price
		FROM orders o
		LEFT JOIN order_details od ON o.order_id = od.order_id
		WHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		ORDER BY o.order_id, od.detail_id`

	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to execute join query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	orderMap := make(map[int64]*models.OrderWithDetails)

	for rows.Next() {
		var order models.Order
		var detailID, productID sql.NullInt64
		var quantity sql.NullInt32
		var unitPrice sql.NullFloat64

		err := rows.Scan(
			&order.OrderID, &order.CustomerID, &order.OrderDate, &order.TotalAmount,
			&detailID, &productID, &quantity, &unitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 受注がまだマップに存在しない場合は作成
		if _, exists := orderMap[order.OrderID]; !exists {
			orderMap[order.OrderID] = &models.OrderWithDetails{
				Order:   order,
				Details: []models.OrderDetail{},
			}
		}

		// 明細が存在する場合は追加
		if detailID.Valid {
			detail := models.OrderDetail{
				DetailID:  detailID.Int64,
				OrderID:   order.OrderID,
				ProductID: productID.Int64,
				Quantity:  int(quantity.Int32),
				UnitPrice: unitPrice.Float64,
			}
			orderMap[order.OrderID].Details = append(orderMap[order.OrderID].Details, detail)
		}
	}

	// マップからスライスに変換
	result := make([]models.OrderWithDetails, 0, len(orderMap))
	for _, order := range orderMap {
		result = append(result, *order)
	}

	return result, nil
}

// GetOrdersWithDetailsBatch - IN句を使用したバッチ取得（推奨方法2）
func (r *OptimizedOrderRepository) GetOrdersWithDetailsBatch(days int) ([]models.OrderWithDetails, error) {
	// 1. 受注一覧を取得
	orders, err := r.GetOrdersByDays(days)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		return []models.OrderWithDetails{}, nil
	}

	// 2. 受注IDをリストで抽出
	orderIDs := make([]int64, len(orders))
	for i, order := range orders {
		orderIDs[i] = order.OrderID
	}

	// 3. 明細を一括取得
	allDetails, err := r.GetDetailsByOrderIDs(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get details: %w", err)
	}

	// 4. メモリ上でグルーピング
	detailsByOrderID := make(map[int64][]models.OrderDetail)
	for _, detail := range allDetails {
		detailsByOrderID[detail.OrderID] = append(detailsByOrderID[detail.OrderID], detail)
	}

	// 5. 結果を組み立て
	result := make([]models.OrderWithDetails, len(orders))
	for i, order := range orders {
		result[i] = models.OrderWithDetails{
			Order:   order,
			Details: detailsByOrderID[order.OrderID],
		}
		// nilスライスを空スライスに変換
		if result[i].Details == nil {
			result[i].Details = []models.OrderDetail{}
		}
	}

	return result, nil
}

// GetDetailsByOrderIDs - IN句を使用した明細の一括取得
func (r *OptimizedOrderRepository) GetDetailsByOrderIDs(orderIDs []int64) ([]models.OrderDetail, error) {
	if len(orderIDs) == 0 {
		return []models.OrderDetail{}, nil
	}

	// IN句用のプレースホルダーを生成
	placeholders := make([]string, len(orderIDs))
	args := make([]interface{}, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT detail_id, order_id, product_id, quantity, unit_price
		FROM order_details
		WHERE order_id IN (%s)
		ORDER BY order_id, detail_id`,
		strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var details []models.OrderDetail
	for rows.Next() {
		var detail models.OrderDetail
		err := rows.Scan(
			&detail.DetailID,
			&detail.OrderID,
			&detail.ProductID,
			&detail.Quantity,
			&detail.UnitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detail row: %w", err)
		}
		details = append(details, detail)
	}

	return details, nil
}

// OptimizedEmployeeRepository - 社員管理の最適化されたリポジトリ
type OptimizedEmployeeRepository struct {
	db *sql.DB
}

// NewOptimizedEmployeeRepository - 最適化された社員リポジトリのコンストラクタ
func NewOptimizedEmployeeRepository(db *sql.DB) *OptimizedEmployeeRepository {
	return &OptimizedEmployeeRepository{db: db}
}

// GetEmployeesWithDepartmentJoin - JOINを使用した社員と部署の一括取得
func (r *OptimizedEmployeeRepository) GetEmployeesWithDepartmentJoin() ([]models.EmployeeWithDepartment, error) {
	query := `
		SELECT 
			e.employee_id,
			e.first_name,
			e.last_name,
			e.email,
			e.department_id,
			e.hire_date,
			e.salary,
			d.department_name,
			d.location
		FROM employees e
		LEFT JOIN departments d ON e.department_id = d.department_id
		ORDER BY e.employee_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute employee join query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var result []models.EmployeeWithDepartment
	for rows.Next() {
		var emp models.EmployeeWithDepartment
		var departmentName, location sql.NullString

		err := rows.Scan(
			&emp.Employee.EmployeeID,
			&emp.Employee.FirstName,
			&emp.Employee.LastName,
			&emp.Employee.Email,
			&emp.Employee.DepartmentID,
			&emp.Employee.HireDate,
			&emp.Employee.Salary,
			&departmentName,
			&location,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee row: %w", err)
		}

		if departmentName.Valid && location.Valid {
			emp.Department = &models.Department{
				DepartmentID:   emp.Employee.DepartmentID,
				DepartmentName: departmentName.String,
				Location:       location.String,
			}
		}

		result = append(result, emp)
	}

	return result, nil
}

// GetEmployeesWithDepartmentBatch - バッチ取得を使用した社員と部署の取得
func (r *OptimizedEmployeeRepository) GetEmployeesWithDepartmentBatch() ([]models.EmployeeWithDepartment, error) {
	// 1. 社員一覧を取得
	employees, err := r.GetAllEmployees()
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	if len(employees) == 0 {
		return []models.EmployeeWithDepartment{}, nil
	}

	// 2. ユニークな部署IDを抽出
	departmentIDSet := make(map[int64]bool)
	for _, emp := range employees {
		departmentIDSet[emp.DepartmentID] = true
	}

	departmentIDs := make([]int64, 0, len(departmentIDSet))
	for id := range departmentIDSet {
		departmentIDs = append(departmentIDs, id)
	}

	// 3. 部署情報を一括取得
	departments, err := r.GetDepartmentsByIDs(departmentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get departments: %w", err)
	}

	// 4. 部署情報をマップに変換
	departmentMap := make(map[int64]models.Department)
	for _, dept := range departments {
		departmentMap[dept.DepartmentID] = dept
	}

	// 5. 結果を組み立て
	result := make([]models.EmployeeWithDepartment, len(employees))
	for i, emp := range employees {
		result[i] = models.EmployeeWithDepartment{
			Employee: emp,
		}
		if dept, exists := departmentMap[emp.DepartmentID]; exists {
			result[i].Department = &dept
		}
	}

	return result, nil
}

// GetAllEmployees - 全社員を取得
func (r *OptimizedEmployeeRepository) GetAllEmployees() ([]models.Employee, error) {
	query := `
		SELECT employee_id, first_name, last_name, email, department_id, hire_date, salary
		FROM employees
		ORDER BY employee_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute employee query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		err := rows.Scan(
			&emp.EmployeeID,
			&emp.FirstName,
			&emp.LastName,
			&emp.Email,
			&emp.DepartmentID,
			&emp.HireDate,
			&emp.Salary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee row: %w", err)
		}
		employees = append(employees, emp)
	}

	return employees, nil
}

// GetDepartmentsByIDs - 指定されたIDの部署情報を一括取得
func (r *OptimizedEmployeeRepository) GetDepartmentsByIDs(departmentIDs []int64) ([]models.Department, error) {
	if len(departmentIDs) == 0 {
		return []models.Department{}, nil
	}

	// IN句用のプレースホルダーを生成
	placeholders := make([]string, len(departmentIDs))
	args := make([]interface{}, len(departmentIDs))
	for i, id := range departmentIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT department_id, department_name, location
		FROM departments
		WHERE department_id IN (%s)`,
		strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute department batch query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var departments []models.Department
	for rows.Next() {
		var dept models.Department
		err := rows.Scan(
			&dept.DepartmentID,
			&dept.DepartmentName,
			&dept.Location,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department row: %w", err)
		}
		departments = append(departments, dept)
	}

	return departments, nil
}

// GetOrdersByDays - 過去N日間の受注を取得
func (r *OptimizedOrderRepository) GetOrdersByDays(days int) ([]models.Order, error) {
	query := `
		SELECT order_id, customer_id, order_date, total_amount
		FROM orders
		WHERE order_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		ORDER BY order_id`

	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to execute orders query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("rows.Close() failed: %v\n", cerr)
		}
	}()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderID,
			&order.CustomerID,
			&order.OrderDate,
			&order.TotalAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
