package repository

import (
	"database/sql"
	"fmt"

	"mysql-n-plus-1-demo/models"
)

// ProblemOrderRepository - N+1問題のあるリポジトリ
type ProblemOrderRepository struct {
	db *sql.DB
}

// NewProblemOrderRepository - 問題のあるリポジトリのコンストラクタ
func NewProblemOrderRepository(db *sql.DB) *ProblemOrderRepository {
	return &ProblemOrderRepository{db: db}
}

// GetOrdersWithDetails - N+1問題のある受注明細取得（問題のあるアプローチ）
func (r *ProblemOrderRepository) GetOrdersWithDetails(days int) ([]models.OrderWithDetails, error) {
	// 1. 受注一覧を取得（1回のクエリ）
	orders, err := r.GetOrdersByDays(days)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var result []models.OrderWithDetails

	// 2. 各受注ごとに明細を取得（N回のクエリ - N+1問題発生！）
	for _, order := range orders {
		details, err := r.GetDetailsByOrderID(order.OrderID)
		if err != nil {
			return nil, fmt.Errorf("failed to get details for order %d: %w", order.OrderID, err)
		}

		result = append(result, models.OrderWithDetails{
			Order:   order,
			Details: details,
		})
	}

	return result, nil
}

// GetOrdersByDays - 過去N日間の受注を取得
func (r *ProblemOrderRepository) GetOrdersByDays(days int) ([]models.Order, error) {
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

// GetDetailsByOrderID - 特定の受注IDの明細を取得（N+1問題の原因）
func (r *ProblemOrderRepository) GetDetailsByOrderID(orderID int64) ([]models.OrderDetail, error) {
	query := `
		SELECT detail_id, order_id, product_id, quantity, unit_price
		FROM order_details
		WHERE order_id = ?
		ORDER BY detail_id`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute order details query: %w", err)
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
			return nil, fmt.Errorf("failed to scan order detail row: %w", err)
		}
		details = append(details, detail)
	}

	return details, nil
}

// ProblemEmployeeRepository - N+1問題のある社員管理リポジトリ
type ProblemEmployeeRepository struct {
	db *sql.DB
}

// NewProblemEmployeeRepository - 問題のある社員リポジトリのコンストラクタ
func NewProblemEmployeeRepository(db *sql.DB) *ProblemEmployeeRepository {
	return &ProblemEmployeeRepository{db: db}
}

// GetEmployeesWithDepartment - N+1問題のある社員と部署の取得
func (r *ProblemEmployeeRepository) GetEmployeesWithDepartment() ([]models.EmployeeWithDepartment, error) {
	// 1. 社員一覧を取得（1回のクエリ）
	employees, err := r.GetAllEmployees()
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	var result []models.EmployeeWithDepartment

	// 2. 各社員ごとに部署情報を取得（N回のクエリ - N+1問題発生！）
	for _, employee := range employees {
		department, err := r.GetDepartmentByID(employee.DepartmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get department for employee %d: %w", employee.EmployeeID, err)
		}

		result = append(result, models.EmployeeWithDepartment{
			Employee:   employee,
			Department: department,
		})
	}

	return result, nil
}

// GetAllEmployees - 全社員を取得
func (r *ProblemEmployeeRepository) GetAllEmployees() ([]models.Employee, error) {
	query := `
		SELECT employee_id, first_name, last_name, email, department_id, hire_date, salary
		FROM employees
		ORDER BY employee_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute employees query: %w", err)
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

// GetDepartmentByID - 特定のIDの部署情報を取得（N+1問題の原因）
func (r *ProblemEmployeeRepository) GetDepartmentByID(departmentID int64) (*models.Department, error) {
	query := `
		SELECT department_id, department_name, location
		FROM departments
		WHERE department_id = ?`

	var dept models.Department
	err := r.db.QueryRow(query, departmentID).Scan(
		&dept.DepartmentID,
		&dept.DepartmentName,
		&dept.Location,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 部署が見つからない場合
		}
		return nil, fmt.Errorf("failed to query department: %w", err)
	}

	return &dept, nil
}
