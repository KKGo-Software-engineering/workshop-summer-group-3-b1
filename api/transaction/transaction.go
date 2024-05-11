package transaction

import (
	"database/sql"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type Transaction struct {
	Date            string  `db:"date" json:"date"`
	Amount          float64 `db:"amount" json:"amount"`
	Category        string  `db:"category" json:"category"`
	TransactionType string  `db:"transaction_type" json:"transaction_type"`
	Note            string  `db:"note" json:"note"`
	ImageURL        string  `db:"image_url" json:"image_url"`
}

type handler struct {
	db *sql.DB
}

var (
	updateTxStmt = "UPDATE transaction SET date = $1, amount = $2, category = $3, transaction_type = $4, note = $5, image_url = $6 WHERE ID = $7 RETURNING date, amount, category, transaction_type, note, image_url;"
)

func New(db *sql.DB) *handler {
	return &handler{
		db: db,
	}
}

func (h handler) Update(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	var tx Transaction
	err := c.Bind(&tx)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var updatedTx Transaction
	err = h.db.QueryRowContext(ctx, updateTxStmt, tx.Date, tx.Amount, tx.Category, tx.TransactionType, tx.Note, tx.ImageURL, 0).Scan(&updatedTx.Date, &updatedTx.Amount, &updatedTx.Category, &updatedTx.TransactionType, &updatedTx.Note, &updatedTx.ImageURL)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("update successfully", zap.Any("updatedTx", updatedTx))

	return c.JSON(http.StatusOK, updatedTx)
}
