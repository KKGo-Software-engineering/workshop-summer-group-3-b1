package spender

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KKGo-Software-engineering/workshop-summer/api/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateSpender(t *testing.T) {

	t.Run("create spender succesfully when feature toggle is enable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(cStmt).WithArgs("HongJot", "hong@jot.ok").WillReturnRows(row)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id": 1, "name": "HongJot", "email": "hong@jot.ok"}`, rec.Body.String())
	})

	t.Run("create spender failed when feature toggle is disable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		cfg := config.FeatureFlag{EnableCreateSpender: false}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("create spender failed when bad request body", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{ bad request body }`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid character")
	})

	t.Run("create spender failed on database (feature toggle is enable) ", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "HongJot", "email": "hong@jot.ok"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(cStmt).WithArgs("HongJot", "hong@jot.ok").WillReturnError(assert.AnError)
		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetAllSpender(t *testing.T) {
	t.Run("get all spender succesfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(1, "HongJot", "hong@jot.ok").
			AddRow(2, "JotHong", "jot@jot.ok")
		mock.ExpectQuery(`SELECT id, name, email FROM spender`).WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetAll(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[{"id": 1, "name": "HongJot", "email": "hong@jot.ok"},
		{"id": 2, "name": "JotHong", "email": "jot@jot.ok"}]`, rec.Body.String())
	})

	t.Run("get all spender failed on database", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(`SELECT id, name, email FROM spender`).WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err := h.GetAll(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetTransactionSummary(t *testing.T) {
	t.Run("get transaction summary of spender successfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		params := utils.KeyValuePairs{"id": "1"}
		utils.SetParams(c, params)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"total", "transaction_type"}).
			AddRow(2000, "income").
			AddRow(1000, "expense")
		mock.ExpectQuery(sumStmt).WithArgs(1).WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetTransactionsSummary(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"summary": { "total_income": 2000, "total_expenses": 1000, "current_balance": 1000 }}`, rec.Body.String())
	})
}

func TestGetSpenderByID(t *testing.T) {
	t.Run("get spender by id successfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "name", "email"}).
			AddRow(1, "HongJot", "hongjot@email.com")

		mock.ExpectQuery(`SELECT id, name, email FROM spender WHERE id = $1`).WithArgs("1").WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err = h.GetByID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"id": 1, "name": "HongJot", "email": "hongjot@email.com"}`, rec.Body.String())
	})

	t.Run("get spender by id not found", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`SELECT id, name, email FROM spender WHERE id = $1`).WithArgs("1").WillReturnError(sql.ErrNoRows)

		h := New(config.FeatureFlag{}, db)
		err = h.GetByID(c)

		fmt.Println("recorder", rec)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("get spender by id failed on database", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`SELECT id, name, email FROM spender WHERE id = $1`).WithArgs("1").WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err = h.GetByID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetTransactionBySpenderID(t *testing.T) {
	t.Run("given valid information should return success", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?page=1&per_page=5", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/transactions/summary")
		params := utils.KeyValuePairs{"id": "1"}
		utils.SetParams(c, params)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(config.FeatureFlag{}, db)

		mock.ExpectQuery(getTxStmt).
			WithArgs(1, 5, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "date", "amount", "category", "transaction_type", "note", "image_url", "spender_id"}).
				AddRow(1, "2021-01-01", 100.0, "food", "expense", "", "", 1).
				AddRow(2, "2021-01-02", 200.0, "saving", "income", "", "", 1))

		mock.ExpectQuery(sumStmt).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"total", "transaction_type"}).AddRow(200, "income").AddRow(100, "expense"))

		mock.ExpectQuery(countTxStmt).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		err := h.GetTransactionBySpenderID(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, rec.Body.String(), `{"transactions":[{"id":1,"date":"2021-01-01","amount":100,"category":"food","transaction_type":"expense","note":"","image_url":"","spender_id":1},{"id":2,"date":"2021-01-02","amount":200,"category":"saving","transaction_type":"income","note":"","image_url":"","spender_id":1}],"summary":{"total_income":200,"total_expenses":100,"current_balance":100},"pagination":{"current_page":1,"total_pages":1,"per_page":5}}`)
	})

}
