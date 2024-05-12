package transaction

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/KKGo-Software-engineering/workshop-summer/api/errs"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Transactions struct {
	ID              int     `db:"id" json:"id"`
	Date            string  `db:"date" json:"date" validate:"required"`
	Amount          float64 `db:"amount" json:"amount" validate:"required,gt=0"`
	Category        string  `db:"category" json:"category" validate:"required"`
	TransactionType string  `db:"transaction_type" json:"transaction_type" validate:"required,oneof=income expense"`
	Note            string  `db:"note" json:"note"`
	ImageURL        string  `db:"image_url" json:"image_url"`
	SpenderID       int     `db:"spender_id" json:"spender_id"`
}

type handler struct {
	db *sql.DB
}

var (
	updateTxStmt = "UPDATE transaction SET date = $1, amount = $2, category = $3, transaction_type = $4, note = $5, image_url = $6 WHERE ID = $7 RETURNING id, date, amount, category, transaction_type, note, image_url, spender_id;"
)

func New(db *sql.DB) *handler {
	return &handler{
		db: db,
	}
}

func (h handler) Update(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("ID parameter is invalid", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	var tx Transactions
	err = c.Bind(&tx)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	if err = c.Validate(tx); err != nil {
		logger.Error("validate request body failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	var updatedTx Transactions

	row := h.db.QueryRowContext(ctx, updateTxStmt, tx.Date, tx.Amount, tx.Category, tx.TransactionType, tx.Note, tx.ImageURL, id)
	err = row.Scan(&updatedTx.ID, &updatedTx.Date, &updatedTx.Amount, &updatedTx.Category, &updatedTx.TransactionType, &updatedTx.Note, &updatedTx.ImageURL, &updatedTx.SpenderID)

	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
	}

	logger.Info("update successfully", zap.Any("updatedTx", updatedTx))

	return c.JSON(http.StatusOK, updatedTx)

}

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	rows, err := h.db.QueryContext(ctx, `SELECT * FROM transaction`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var txs []Transactions
	for rows.Next() {
		var tx Transactions
		err := rows.Scan(&tx.ID, &tx.Date, &tx.Amount, &tx.Category, &tx.TransactionType, &tx.Note, &tx.ImageURL, &tx.SpenderID)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		txs = append(txs, tx)
	}

	return c.JSON(http.StatusOK, txs)
}

func (h handler) Create(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()
	var tx Transactions
	err := c.Bind(&tx)

	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var id int
	cstm := `INSERT INTO transaction ( date, amount, category, transaction_type, note, image_url, spender_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	err = h.db.QueryRowContext(ctx, cstm, tx.Date, tx.Amount, tx.Category, tx.TransactionType, tx.Note, tx.ImageURL, tx.SpenderID).Scan(&id)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("create successfully", zap.Int("id", id))
	tx.ID = id
	return c.JSON(http.StatusCreated, tx)
}
