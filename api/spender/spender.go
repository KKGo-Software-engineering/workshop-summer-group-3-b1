package spender

import (
	"database/sql"
	"net/http"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Spender struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

const (
	cStmt = `INSERT INTO spender (name, email) VALUES ($1, $2) RETURNING id;`
)

func (h handler) Create(c echo.Context) error {
	if !h.flag.EnableCreateSpender {
		return c.JSON(http.StatusForbidden, "create new spender feature is disabled")
	}

	logger := mlog.L(c)
	ctx := c.Request().Context()
	var sp Spender
	err := c.Bind(&sp)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var lastInsertId int64
	err = h.db.QueryRowContext(ctx, cStmt, sp.Name, sp.Email).Scan(&lastInsertId)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("create successfully", zap.Int64("id", lastInsertId))
	sp.ID = lastInsertId
	return c.JSON(http.StatusCreated, sp)
}

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, `SELECT id, name, email FROM spender`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var sps []Spender
	for rows.Next() {
		var sp Spender
		err := rows.Scan(&sp.ID, &sp.Name, &sp.Email)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		sps = append(sps, sp)
	}

	return c.JSON(http.StatusOK, sps)
}

// GetTransactionsSummary returns the summary of total income, total expense, and current balance
// from the transaction table. The response is a map with the following keys:
// - total_income: total income amount
// - total_expense: total expense amount
// - current_balance: current balance (income - expense)
func (h handler) GetTransactionsSummary(c echo.Context) error {
	ctx := c.Request().Context()
	logger := mlog.L(c)
	id := c.Param("id")
	rows, err := h.db.QueryContext(ctx, `
		SELECT
			SUM(amount) AS total, transaction_type
		FROM "transaction"
		WHERE spender_id = $1
		GROUP BY transaction_type
	`, id)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	totalIncome, totalExpense := 0.0, 0.0
	for rows.Next() {
		var amount float64
		var transactionType string
		if err := rows.Scan(&amount, &transactionType); err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		if transactionType == "income" {
			totalIncome += amount
		} else {
			totalExpense += amount
		}
	}

	summary := map[string]float64{
		"total_income":    totalIncome,
		"total_expense":   totalExpense,
		"current_balance": totalIncome - totalExpense,
	}
	res := make(map[string]interface{})
	res["summary"] = summary
	return c.JSON(http.StatusOK, res)
}

func (h handler) GetSpenderByID(c echo.Context) error {

	ctx := c.Request().Context()
	logger := mlog.L(c)
	id := c.Param("id")
	rows, err := h.db.QueryContext(ctx, `SELECT id, "name", email FROM spender WHERE id = $1`, id)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	data := Spender{}
	for rows.Next() {
		if err := rows.Scan(&data.ID, &data.Name, &data.Email); err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	}
	return c.JSON(http.StatusOK, data)
}
