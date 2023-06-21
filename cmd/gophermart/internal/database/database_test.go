package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"testing"
	"time"
)

func TestManager_GetAllOrders(t *testing.T) {
	t.Run("positive: orders exist", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery(`select order_id from orders`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}).AddRow("100500"))

		manager, err := New(ctx, db)
		assert.NoError(t, err)
		orders, err := manager.GetAllOrders()
		assert.NoError(t, err)
		assert.Equal(t, orders, []string{"100500"})
	})

	t.Run("positive: no orders", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectQuery(`select order_id from orders`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}))

		manager, err := New(ctx, db)
		assert.NoError(t, err)
		orders, err := manager.GetAllOrders()
		assert.NoError(t, err)
		assert.Equal(t, orders, []string{})
	})
}

func TestManager_GetBalanceInfo(t *testing.T) {
	testCases := []struct {
		name        string
		balance     *sqlmock.Rows
		withdrawals *sqlmock.Rows
		result      string
	}{
		{
			name:        "positive",
			balance:     sqlmock.NewRows([]string{"balance"}).AddRow(100.5),
			withdrawals: sqlmock.NewRows([]string{"withdrawn"}).AddRow(30.4),
			result:      `{"current":100.5,"withdrawn":30.4}`,
		},
		{
			name:        "positive: no withdrawals",
			balance:     sqlmock.NewRows([]string{"balance"}).AddRow(100.5),
			withdrawals: sqlmock.NewRows([]string{"withdrawn"}),
			result:      `{"current":100.5,"withdrawn":0}`,
		},
		{
			name:        "positive: no withdrawals and no accruals",
			balance:     sqlmock.NewRows([]string{"balance"}),
			withdrawals: sqlmock.NewRows([]string{"withdrawn"}),
			result:      `{"current":0,"withdrawn":0}`,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select coalesce(sum(accrual), 0) - coalesce(sum(amount), 0) as balance from orders o left join withdraw`)).WillReturnRows(tt.balance)
			mock.ExpectQuery(regexp.QuoteMeta(`select sum(amount) as withdrawn from withdraw where login`)).WillReturnRows(tt.withdrawals)
			manager, err := New(ctx, db)
			assert.NoError(t, err)

			info, err := manager.GetBalanceInfo("test-login")
			assert.NoError(t, err)
			assert.Equal(t, string(info), tt.result)
		})
	}
}

func TestManager_GetWithdrawals(t *testing.T) {
	testCases := []struct {
		name          string
		withdrawals   *sqlmock.Rows
		result        string
		expectedError error
	}{
		{
			name: "positive",
			withdrawals: sqlmock.NewRows([]string{"order_id", "amount", "processed_at"}).AddRow("100500", 100.5, time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)).
				AddRow("100501", 200.5, time.Date(2021, 9, 15, 14, 30, 45, 100, time.Local)).
				AddRow("100502", 300.5, time.Date(2021, 10, 15, 14, 30, 45, 100, time.Local)).
				AddRow("100503", 320.5, time.Date(2021, 11, 15, 14, 30, 45, 100, time.Local)),
			result: `[{"order":"100500","processed_at":"2021-08-15T14:30:45.0000001+03:00","sum":100.5},{"order":"100501","processed_at":"2021-09-15T14:30:45.0000001+03:00","sum":200.5},{"order":"100502","processed_at":"2021-10-15T14:30:45.0000001+03:00","sum":300.5},{"order":"100503","processed_at":"2021-11-15T14:30:45.0000001+03:00","sum":320.5}]`,
		},
		{
			name:          "negative: no data",
			withdrawals:   sqlmock.NewRows([]string{"order_id", "amount", "processed_at"}),
			expectedError: ErrNoData,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select order_id, amount, processed_at from withdraw`)).WillReturnRows(tt.withdrawals)
			manager, err := New(ctx, db)
			assert.NoError(t, err)

			withdrawals, err := manager.GetWithdrawals("test-login")
			if tt.expectedError == nil {
				assert.Equal(t, string(withdrawals), tt.result)
			} else {
				assert.Equal(t, err, tt.expectedError)
			}
		})
	}
}

func TestManager_Withdraw(t *testing.T) {
	testCases := []struct {
		name          string
		balance       *sqlmock.Rows
		sum           float64
		expectedError error
	}{
		{
			name:    "positive",
			sum:     50.5,
			balance: sqlmock.NewRows([]string{"balance"}).AddRow(100.5),
		},
		{
			name:          "negative: insufficient balance",
			sum:           150.5,
			balance:       sqlmock.NewRows([]string{"balance"}).AddRow(100.5),
			expectedError: ErrInsufficientBalance,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select coalesce(sum(accrual), 0) - coalesce(sum(amount), 0) as balance from orders o left join withdraw`)).WillReturnRows(tt.balance)
			mock.ExpectExec(`insert into withdraw values`).WillReturnResult(sqlmock.NewResult(0, 1))
			manager, err := New(ctx, db)
			assert.NoError(t, err)

			err = manager.Withdraw("test-login", "100500", tt.sum)
			assert.Equal(t, err, tt.expectedError)
		})
	}
}

func TestManager_GetUserOrders(t *testing.T) {
	testCases := []struct {
		name        string
		orders      *sqlmock.Rows
		expectedErr error
		result      string
	}{
		{
			name: "positive",
			orders: sqlmock.NewRows([]string{"order_id", "status", "accrual", "uploaded_at"}).
				AddRow("1", "NEW", 100.5, time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)).
				AddRow("2", "PROCESSED", 20.1, time.Date(2021, 9, 15, 14, 30, 45, 100, time.Local)).
				AddRow("3", "PROCESSING", 0.01, time.Date(2021, 10, 15, 14, 30, 45, 100, time.Local)).
				AddRow("4", "INVALID", 0.8, time.Date(2021, 11, 15, 14, 30, 45, 100, time.Local)),
			result: `[{"number":"1","uploaded_at":"2021-08-15T14:30:45.0000001+03:00","status":"NEW","accrual":100.5},{"number":"2","uploaded_at":"2021-09-15T14:30:45.0000001+03:00","status":"PROCESSED","accrual":20.1},{"number":"3","uploaded_at":"2021-10-15T14:30:45.0000001+03:00","status":"PROCESSING","accrual":0.01},{"number":"4","uploaded_at":"2021-11-15T14:30:45.0000001+03:00","status":"INVALID","accrual":0.8}]`,
		},
		{
			name:        "positive: no data",
			orders:      sqlmock.NewRows([]string{"order_id", "status", "accrual", "uploaded_at"}),
			expectedErr: ErrNoData,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select order_id, status, accrual, uploaded_at from orders`)).WithArgs("test-login").WillReturnRows(tt.orders)
			manager, err := New(ctx, db)
			assert.NoError(t, err)

			orders, err := manager.GetUserOrders("test-login")
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.Equal(t, string(orders), tt.result)
			}
		})
	}
}

func TestManager_UpdateOrderInfo(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		login := "test-login"
		order := "100500"
		orderTime := time.Now()
		info := models.OrderInfo{
			UserName:  &login,
			OrderID:   order,
			Order:     &order,
			CreatedAt: &orderTime,
			Status:    "NOW",
			Accrual:   100.5,
		}

		mock.ExpectExec(regexp.QuoteMeta(`update orders set`)).WithArgs(info.Status, info.Accrual, info.OrderID).WillReturnResult(sqlmock.NewResult(0, 0))
		manager, err := New(ctx, db)
		assert.NoError(t, err)

		err = manager.UpdateOrderInfo(&info)
		assert.NoError(t, err)
	})
	t.Run("negative", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		login := "test-login"
		order := "100500"
		orderTime := time.Now()
		info := models.OrderInfo{
			UserName:  &login,
			OrderID:   order,
			Order:     &order,
			CreatedAt: &orderTime,
			Status:    "NOW",
			Accrual:   100.5,
		}

		mock.ExpectExec(regexp.QuoteMeta(`update orders set`)).WithArgs(info.Status, info.Accrual, info.OrderID).WillReturnError(errors.New("some error"))
		manager, err := New(ctx, db)
		assert.NoError(t, err)

		err = manager.UpdateOrderInfo(&info)
		assert.EqualError(t, err, "error while updating order info: some error")
	})
}

func TestManager_LoadOrder(t *testing.T) {
	testCases := []struct {
		name        string
		orders      *sqlmock.Rows
		ordersErr   error
		expectedErr error
	}{
		{
			name:      "positive",
			orders:    sqlmock.NewRows([]string{"login"}),
			ordersErr: sql.ErrNoRows,
		},
		{
			name:        "negative: same user",
			orders:      sqlmock.NewRows([]string{"login"}).AddRow("test-login"),
			expectedErr: ErrCreatedBySameUser,
		},
		{
			name:        "negative: diff user",
			orders:      sqlmock.NewRows([]string{"login"}).AddRow("test-diff-login"),
			expectedErr: ErrCreatedDiffUser,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select login from orders`)).WithArgs("100500").WillReturnRows(tt.orders)
			if errors.Is(tt.ordersErr, sql.ErrNoRows) {
				mock.ExpectExec(regexp.QuoteMeta(`insert into orders`)).WillReturnResult(sqlmock.NewResult(0, 0))
			}
			manager, err := New(ctx, db)
			assert.NoError(t, err)

			err = manager.LoadOrder("test-login", "100500")
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_Register(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(regexp.QuoteMeta(`insert into registered_users values`)).WillReturnResult(sqlmock.NewResult(0, 0))
		manager, err := New(ctx, db)
		assert.NoError(t, err)

		err = manager.Register("test-login", "test-password")
		assert.NoError(t, err)
	})
	t.Run("negative: user already exists", func(t *testing.T) {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(regexp.QuoteMeta(`insert into registered_users values`)).WillReturnError(ErrDublicateKey{Key: "registered_users_pkey"})
		manager, err := New(ctx, db)
		assert.NoError(t, err)

		err = manager.Register("test-login", "test-password")
		assert.EqualError(t, err, ErrUserAlreadyExists.Error())
	})
}

func TestManager_Login(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)

	testCases := []struct {
		name        string
		login       string
		password    string
		creds       *sqlmock.Rows
		expectedErr error
	}{
		{
			name:     "positive",
			login:    "test-login",
			password: "test-password",
			creds:    sqlmock.NewRows([]string{"login", "password"}).AddRow("test-login", hash),
		},
		{
			name:        "negative: invalid creds",
			login:       "test-login",
			password:    "test-password",
			creds:       sqlmock.NewRows([]string{"login", "password"}).AddRow("test-login", "wrong-pass"),
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:        "negative: no such user",
			login:       "test-login",
			password:    "test-password",
			creds:       sqlmock.NewRows([]string{"login", "password"}).AddRow("other-login", hash),
			expectedErr: ErrNoSuchUser,
		},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectExec(`create table if not exists registered_users`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists orders`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`create table if not exists withdraw`).WillReturnResult(sqlmock.NewResult(0, 0))

		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`select login, password from registered_users`)).WillReturnRows(tt.creds)
			manager, err := New(ctx, db)
			assert.NoError(t, err)
			err = manager.Login(tt.login, tt.password)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
