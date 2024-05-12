package spender

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/KKGo-Software-engineering/workshop-summer/api/errs"
	"github.com/KKGo-Software-engineering/workshop-summer/api/transaction"
	"github.com/KKGo-Software-engineering/workshop-summer/api/utils"

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

type Summary struct {
	TotalIncome    float64 `json:"total_income"`
	TotalExpenses  float64 `json:"total_expenses"`
	CurrentBalance float64 `json:"current_balance"`
}

type TransactionResponse struct {
	Transactions []transaction.Transaction `json:"transactions"`
	Summary      Summary                   `json:"summary"`
	Pagination   utils.Pagination          `json:"pagination"`
}

type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

const (
	cStmt       = `INSERT INTO spender (name, email) VALUES ($1, $2) RETURNING id;`
	getTxStmt   = `SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction WHERE spender_id = $1 LIMIT $2 OFFSET $3`
	countTxStmt = `SELECT COUNT(*) FROM transaction WHERE spender_id = $1`
	sumStmt     = `SELECT SUM(amount) AS total, transaction_type FROM "transaction" WHERE spender_id = $1 GROUP BY transaction_type`
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

func (h handler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	logger := mlog.L(c)

	id := c.Param("id")
	rows, err := h.db.QueryContext(ctx, `SELECT id, name, email FROM spender WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("spender not found", zap.String("id", id))
			return c.JSON(http.StatusNotFound, "spender not found")
		}

		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	sp := Spender{}
	for rows.Next() {
		if err := rows.Scan(&sp.ID, &sp.Name, &sp.Email); err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(http.StatusOK, sp)
}

// GetTransactionsSummary returns the summary of total income, total expense, and current balance
// from the transaction table. The response is a map with the following keys:
// - total_income: total income amount
// - total_expense: total expense amount
// - current_balance: current balance (income - expense)
func (h handler) GetTransactionsSummary(c echo.Context) error {
	ctx := c.Request().Context()
	logger := mlog.L(c)
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("ID parameter is invalid", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	summary, err := h.getSummaryBySpenderID(ctx, uint(id))
	if err != nil {
		logger.Error("get transaction summary error")
		return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
	}

	res := make(map[string]Summary)
	res["summary"] = *summary
	return c.JSON(http.StatusOK, res)
}

func (h handler) getSummaryBySpenderID(ctx context.Context, ID uint) (*Summary, error) {
	rows, err := h.db.QueryContext(ctx, sumStmt, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalIncome, totalExpense := 0.0, 0.0
	for rows.Next() {
		var amount float64
		var transactionType string
		if err := rows.Scan(&amount, &transactionType); err != nil {
			return nil, err
		}

		if transactionType == "income" {
			totalIncome += amount
		} else {
			totalExpense += amount
		}
	}

	return &Summary{
		TotalIncome:    totalIncome,
		TotalExpenses:  totalExpense,
		CurrentBalance: totalIncome - totalExpense,
	}, nil
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

func (h handler) GetTransactionBySpenderID(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	page := 1
	perPage := 10
	var err error

	pageQuery := c.QueryParam("page")
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			logger.Error("page query is invalid", zap.Error(err), zap.String("page query", pageQuery))
			return c.JSON(http.StatusBadRequest, errs.ParseError(err))
		}
	}

	perPageQuery := c.QueryParam("per_page")
	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			logger.Error("per_page query is invalid", zap.Error(err))
			return c.JSON(http.StatusBadRequest, errs.ParseError(err))
		}
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("ID parameter is invalid", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	offset := (page - 1) * perPage
	transactions := make([]transaction.Transaction, 0)
	rows, err := h.db.QueryContext(ctx, getTxStmt, id, perPage, offset)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var tx transaction.Transaction

		err := rows.Scan(&tx.ID, &tx.Date, &tx.Amount, &tx.Category, &tx.TransactionType, &tx.Note, &tx.ImageURL, &tx.SpenderID)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
		}

		transactions = append(transactions, tx)
	}

	summary, err := h.getSummaryBySpenderID(ctx, uint(id))
	if err != nil {
		logger.Error("get transaction summary error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
	}

	var totalRows int64
	err = h.db.QueryRowContext(ctx, countTxStmt, id).Scan(&totalRows)
	if err != nil {
		logger.Error("count total rows error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
	}

	totalPages := math.Ceil(float64(totalRows) / float64(perPage))

	return c.JSON(http.StatusOK, TransactionResponse{
		Transactions: transactions,
		Summary:      *summary,
		Pagination: utils.Pagination{
			CurrentPage: uint(page),
			TotalPages:  uint(totalPages),
			PerPage:     uint(perPage),
		},
	})
}
