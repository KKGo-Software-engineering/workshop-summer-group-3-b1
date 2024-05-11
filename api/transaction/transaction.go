package transaction

import (
	"database/sql"
	"github.com/KKGo-Software-engineering/workshop-summer/api/errs"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Transaction struct {
	Date            string  `db:"date" json:"date" validate:"required"`
	Amount          float64 `db:"amount" json:"amount" validate:"required,gt=0"`
	Category        string  `db:"category" json:"category" validate:"required"`
	TransactionType string  `db:"transaction_type" json:"transaction_type" validate:"required,oneof=income expense"`
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

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("ID parameter is invalid", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	var tx Transaction
	err = c.Bind(&tx)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	if err = c.Validate(tx); err != nil {
		logger.Error("validate request body failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errs.ParseError(err))
	}

	var updatedTx Transaction
	err = h.db.
		QueryRowContext(ctx, updateTxStmt, tx.Date, tx.Amount, tx.Category, tx.TransactionType, tx.Note, tx.ImageURL, id).
		Scan(&updatedTx.Date, &updatedTx.Amount, &updatedTx.Category, &updatedTx.TransactionType, &updatedTx.Note, &updatedTx.ImageURL)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, errs.ParseError(err))
	}

	logger.Info("update successfully", zap.Any("updatedTx", updatedTx))

	return c.JSON(http.StatusOK, updatedTx)
}
