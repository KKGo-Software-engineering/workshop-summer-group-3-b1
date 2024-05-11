package transaction

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/errs"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateTransactionByID(t *testing.T) {
	t.Run("given transaction information should update transaction", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date": "2024-05-11 15:04:05","amount": 25.5,"category": "food","transaction_type": "income","note": "","image_url": ""}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()
		h := New(db)

		row := sqlmock.NewRows([]string{"date", "amount", "category", "transaction_type", "note", "image_url"}).AddRow("2024-05-11 15:04:05", 25.5, "food", "income", "", "")
		mock.ExpectQuery(updateTxStmt).WithArgs("2024-05-11 15:04:05", 25.5, "food", "income", "", "", 0).WillReturnRows(row)

		err := h.Update(c)

		expected := `{"date": "2024-05-11 15:04:05","amount": 25.5,"category": "food","transaction_type": "income","note": "","image_url": ""}`

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, expected, rec.Body.String())
	})

	t.Run("given transaction information should update transaction", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": ""}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(db)

		row := sqlmock.NewRows([]string{"date", "amount", "category", "transaction_type", "note", "image_url"}).AddRow("2024-05-11 15:04:05", 30, "food", "income", "", "")
		mock.ExpectQuery(updateTxStmt).WithArgs("2024-05-11 15:04:05", float64(30), "food", "income", "", "", 0).WillReturnRows(row)

		err := h.Update(c)

		expected := `{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": ""}`

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, expected, rec.Body.String())
	})

	t.Run("given valid information should return error when database failed", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": ""}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(db)

		mockErr := errs.ErrInternalDatabaseError
		mock.ExpectQuery(updateTxStmt).WithArgs("2024-05-11 15:04:05", float64(30), "food", "income", "", "", 0).WillReturnError(mockErr)

		h.Update(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
