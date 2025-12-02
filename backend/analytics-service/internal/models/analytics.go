package models

import (
	"time"
)

type Overview struct {
	Period          string           `json:"period"`
	TotalIncome     float64          `json:"total_income"`
	TotalExpense    float64          `json:"total_expense"`
	NetIncome       float64          `json:"net_income"`
	SavingsRate     float64          `json:"savings_rate"`
	TopCategories   []CategoryStat   `json:"top_categories"`
	MonthComparison *Comparison      `json:"month_comparison"`
	AccountBalances []AccountBalance `json:"account_balances"`
}

type CategoryStat struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
	Trend        float64 `json:"trend"` // % change from previous period
}

type Comparison struct {
	IncomeDiff    float64 `json:"income_diff"`
	ExpenseDiff   float64 `json:"expense_diff"`
	IncomeChange  float64 `json:"income_change"`  // percentage
	ExpenseChange float64 `json:"expense_change"` // percentage
}

type AccountBalance struct {
	AccountID   string  `json:"account_id"`
	AccountName string  `json:"account_name"`
	Balance     float64 `json:"balance"`
	Percentage  float64 `json:"percentage"` // % of total
}

type Trend struct {
	Date    time.Time `json:"date"`
	Income  float64   `json:"income"`
	Expense float64   `json:"expense"`
	Balance float64   `json:"balance"`
}

type Forecast struct {
	Period           string  `json:"period"`
	PredictedIncome  float64 `json:"predicted_income"`
	PredictedExpense float64 `json:"predicted_expense"`
	PredictedBalance float64 `json:"predicted_balance"`
	Confidence       float64 `json:"confidence"`
	BasedOnMonths    int     `json:"based_on_months"`
}

type Insight struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Value       float64   `json:"value,omitempty"`
	Priority    string    `json:"priority"` // high, medium, low
	Date        time.Time `json:"date"`
}

type Cashflow struct {
	Date         time.Time        `json:"date"`
	OpenBalance  float64          `json:"open_balance"`
	CloseBalance float64          `json:"close_balance"`
	TotalInflow  float64          `json:"total_inflow"`
	TotalOutflow float64          `json:"total_outflow"`
	NetCashflow  float64          `json:"net_cashflow"`
	Details      []CashflowDetail `json:"details"`
}

type CashflowDetail struct {
	CategoryName string  `json:"category_name"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Count        int     `json:"count"`
}

type UserAction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Details   string    `json:"details"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

type ExportRequest struct {
	UserID    string    `json:"user_id"`
	Format    string    `json:"format"` // csv, xlsx, pdf
	DateFrom  time.Time `json:"date_from"`
	DateTo    time.Time `json:"date_to"`
	AccountID string    `json:"account_id,omitempty"`
	Type      string    `json:"type,omitempty"`
}

type Report struct {
	Title       string                 `json:"title"`
	Period      string                 `json:"period"`
	GeneratedAt time.Time              `json:"generated_at"`
	Summary     map[string]interface{} `json:"summary"`
	Charts      []ChartData            `json:"charts"`
	Tables      []TableData            `json:"tables"`
}

type ChartData struct {
	Type   string                   `json:"type"`
	Title  string                   `json:"title"`
	Labels []string                 `json:"labels"`
	Data   []map[string]interface{} `json:"data"`
}

type TableData struct {
	Title   string          `json:"title"`
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}
