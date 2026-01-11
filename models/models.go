package models

import "time"

// Order - 受注モデル
type Order struct {
	OrderID     int64     `json:"order_id"`
	CustomerID  int64     `json:"customer_id"`
	OrderDate   time.Time `json:"order_date"`
	TotalAmount float64   `json:"total_amount"`
}

// OrderDetail - 受注明細モデル
type OrderDetail struct {
	DetailID  int64   `json:"detail_id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// OrderWithDetails - 受注と明細を組み合わせたモデル
type OrderWithDetails struct {
	Order   Order         `json:"order"`
	Details []OrderDetail `json:"details"`
}

// Employee - 社員モデル
type Employee struct {
	EmployeeID   int64     `json:"employee_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	DepartmentID int64     `json:"department_id"`
	HireDate     time.Time `json:"hire_date"`
	Salary       float64   `json:"salary"`
}

// Department - 部署モデル
type Department struct {
	DepartmentID   int64  `json:"department_id"`
	DepartmentName string `json:"department_name"`
	Location       string `json:"location"`
}

// EmployeeWithDepartment - 社員と部署を組み合わせたモデル
type EmployeeWithDepartment struct {
	Employee   Employee    `json:"employee"`
	Department *Department `json:"department,omitempty"`
}
