package transaction

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/errs"
	"github.com/KKGo-Software-engineering/workshop-summer/api/utils"
	cv "github.com/KKGo-Software-engineering/workshop-summer/api/validator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTransactionByID(t *testing.T) {
	t.Run("given transaction information should update transaction", func(t *testing.T) {
		e := echo.New()
		e.Validator = &cv.CustomValidator{Validator: validator.New()}
		defer e.Close()

		type Mock struct {
			Arg          Transactions
			ReturningRow Transactions
		}

		type TestCase struct {
			Request  string
			Expected string
			Mock     Mock
		}

		cols := []string{"id", "date", "amount", "category", "transaction_type", "note", "image_url", "spender_id"}
		tcs := []TestCase{
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": 25.5,"category": "food","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"id": 1, "date": "2024-05-11 15:04:05","amount": 25.5,"category": "food","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Mock: Mock{
					Arg: Transactions{
						Date:            "2024-05-11 15:04:05",
						Amount:          25.5,
						Category:        "food",
						TransactionType: "income",
						Note:            "",
						ImageURL:        "",
					},
					ReturningRow: Transactions{
						ID:              1,
						Date:            "2024-05-11 15:04:05",
						Amount:          25.5,
						Category:        "food",
						TransactionType: "income",
						Note:            "",
						ImageURL:        "",
						SpenderID:       1,
					},
				},
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"id": 1, "date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Mock: Mock{
					Arg: Transactions{
						Date:            "2024-05-11 15:04:05",
						Amount:          30,
						Category:        "food",
						TransactionType: "income",
						Note:            "",
						ImageURL:        "",
					},
					ReturningRow: Transactions{
						ID:              1,
						Date:            "2024-05-11 15:04:05",
						Amount:          30,
						Category:        "food",
						TransactionType: "income",
						Note:            "",
						ImageURL:        "",
						SpenderID:       1,
					},
				},
			},
		}

		for _, tc := range tcs {
			req := httptest.NewRequest(http.MethodPut, "/transactions/1", strings.NewReader(tc.Request))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/transactions/:id")
			params := utils.KeyValuePairs{"id": "1"}
			utils.SetParams(c, params)

			db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			defer db.Close()
			h := New(db)

			returningRow := tc.Mock.ReturningRow
			arg := tc.Mock.Arg
			row := sqlmock.NewRows(cols).AddRow(returningRow.ID, returningRow.Date, returningRow.Amount, returningRow.Category, returningRow.TransactionType, returningRow.Note, returningRow.ImageURL, returningRow.SpenderID)
			mock.ExpectQuery(updateTxStmt).WithArgs(arg.Date, arg.Amount, arg.Category, arg.TransactionType, arg.Note, arg.ImageURL, 1).WillReturnRows(row)

			err := h.Update(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, tc.Expected, rec.Body.String())
		}
	})

	t.Run("given valid information should return error when database failed", func(t *testing.T) {
		e := echo.New()
		e.Validator = &cv.CustomValidator{Validator: validator.New()}
		defer e.Close()

		arg := Transactions{
			Date:            "2024-05-11 15:04:05",
			Amount:          30,
			Category:        "food",
			TransactionType: "income",
			Note:            "",
			ImageURL:        "",
		}

		req := httptest.NewRequest(http.MethodPut, "/transactions/1", strings.NewReader(`{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": ""}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/transactions/:id")
		params := utils.KeyValuePairs{"id": "1"}
		utils.SetParams(c, params)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(db)

		mockErr := errs.ErrInternalDatabaseError
		mock.ExpectQuery(updateTxStmt).WithArgs(arg.Date, arg.Amount, arg.Category, arg.TransactionType, arg.Note, arg.ImageURL, 1).WillReturnError(mockErr)

		err := h.Update(c)

		expected := `{"messages":["internal database error"]}`

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.NoError(t, err)
		assert.JSONEq(t, expected, rec.Body.String())
	})

	t.Run("given invalid request body should return error", func(t *testing.T) {
		e := echo.New()
		e.Validator = &cv.CustomValidator{Validator: validator.New()}
		defer e.Close()

		type TestCase struct {
			Request  string
			Expected string
		}

		tcs := []TestCase{
			{
				Request: `{
					"date": "2024-05-11 15:04:05",
					"category": "food",
					"transaction_type": "income",
					"note": "", "image_url": "",
					"spender_id": 1
				}`,
				Expected: `{"messages":["field Amount is required"]}`,
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": -1,"category": "food","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"messages":["the value of Amount must be greater than 0"]}`,
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": 25,"category": "food","transaction_type": "invalid-transaction-type","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"messages":["the value of TransactionType must be one of income expense"]}`,
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": 25,"category": "","transaction_type": "income","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"messages":["field Category is required"]}`,
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": 25,"category": "","transaction_type": "","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"messages":["field Category is required","field TransactionType is required"]}`,
			},
			{
				Request:  `{"date": "2024-05-11 15:04:05","amount": -1,"category": "","transaction_type": "","note": "","image_url": "", "spender_id": 1}`,
				Expected: `{"messages":["the value of Amount must be greater than 0","field Category is required","field TransactionType is required"]}`,
			},
		}

		for _, tc := range tcs {
			req := httptest.NewRequest(http.MethodPut, "/transactions/1", strings.NewReader(tc.Request))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/transactions/:id")
			params := utils.KeyValuePairs{"id": "1"}
			utils.SetParams(c, params)

			db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			defer db.Close()

			h := New(db)

			err := h.Update(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.JSONEq(t, tc.Expected, rec.Body.String())
		}
	})

	t.Run("given invalid ID should return error", func(t *testing.T) {
		e := echo.New()
		e.Validator = &cv.CustomValidator{Validator: validator.New()}
		defer e.Close()

		req := httptest.NewRequest(http.MethodPut, "/transactions/1", strings.NewReader(`{"date": "2024-05-11 15:04:05","amount": 30,"category": "food","transaction_type": "income","note": "","image_url": ""}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/transactions/:id")
		params := utils.KeyValuePairs{"id": "invalid-id"}
		utils.SetParams(c, params)

		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(db)

		err := h.Update(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"messages":["strconv.Atoi: parsing \"invalid-id\": invalid syntax"]}`, rec.Body.String())
	})

	t.Run("given invalid request body format should return error", func(t *testing.T) {
		e := echo.New()
		e.Validator = &cv.CustomValidator{Validator: validator.New()}
		defer e.Close()

		req := httptest.NewRequest(http.MethodPut, "/transactions/1", strings.NewReader(`[]`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/transactions/:id")
		params := utils.KeyValuePairs{"id": "1"}
		utils.SetParams(c, params)

		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(db)

		err := h.Update(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"messages":["code=400, message=Unmarshal type error: expected=transaction.Transactions, got=array, field=, offset=1, internal=json: cannot unmarshal array into Go value of type transaction.Transactions"]}`, rec.Body.String())
	})
}
