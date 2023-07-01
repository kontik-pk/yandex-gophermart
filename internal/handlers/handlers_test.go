package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/database"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandler_Register(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/register", handler.Register)

		srv := httptest.NewServer(r)
		defer srv.Close()

		jar, _ := cookiejar.New(nil)
		response, err := resty.New().SetCookieJar(jar).R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "200 OK")
		assert.True(t, response.Header().Get("Authorization") != "")
		assert.True(t, response.Header().Get("Set-Cookie") != "")
		fmt.Println(response.Header())
	})
	t.Run("bad request", func(t *testing.T) {
		manager := newMockDbManager(t)
		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/register", handler.Register)

		srv := httptest.NewServer(r)
		defer srv.Close()

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test" }`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "400 Bad Request")
	})
	t.Run("login is unavailable", func(t *testing.T) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(database.ErrUserAlreadyExists)

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/register", handler.Register)

		srv := httptest.NewServer(r)
		defer srv.Close()

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "409 Conflict")
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Login", "test", "test").Return(nil)

		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/login", handler.Login)

		srv := httptest.NewServer(r)
		defer srv.Close()

		responce, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/login", srv.URL))
		assert.NoError(t, err)
		assert.Equal(t, responce.Status(), "200 OK")
	})
	t.Run("bad request", func(t *testing.T) {
		manager := newMockDbManager(t)
		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/login", handler.Login)

		srv := httptest.NewServer(r)
		defer srv.Close()

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test" }`).
			Post(fmt.Sprintf("%s/api/user/login", srv.URL))
		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "400 Bad Request")
	})
	t.Run("incorrect password", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Login", "test", "incorrect-password").Return(database.ErrInvalidCredentials)
		logger, err := zap.NewDevelopment()
		if err != nil {
			os.Exit(1)
		}
		defer logger.Sync()

		log := *logger.Sugar()
		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Post("/api/user/login", handler.Login)

		srv := httptest.NewServer(r)
		defer srv.Close()

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "incorrect-password"}`).
			Post(fmt.Sprintf("%s/api/user/login", srv.URL))
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "401 Unauthorized")
	})
}

func TestHandler_LoadOrder(t *testing.T) {
	orderNum := "614371538763429"
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log := *logger.Sugar()

	t.Run("positive: new order created", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		manager.On("LoadOrder", "test", "614371538763429").Return(nil)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Post("/api/user/orders", handler.LoadOrder)
		})

		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").
			SetBody([]byte(orderNum)).
			SetHeader("Authorization", user.Header().Get("Authorization")).Post(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "202 Accepted")
	})

	t.Run("positive: order was already created by the same user", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		manager.On("LoadOrder", "test", "614371538763429").Return(database.ErrCreatedBySameUser)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Post("/api/user/orders", handler.LoadOrder)
		})

		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").
			SetBody([]byte(orderNum)).
			SetHeader("Authorization", user.Header().Get("Authorization")).Post(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "200 OK")
	})

	t.Run("negative: bad order", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Post("/api/user/orders", handler.LoadOrder)
		})

		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").
			SetBody([]byte(`193892`)).
			SetHeader("Authorization", user.Header().Get("Authorization")).Post(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "422 Unprocessable Entity")
	})

	t.Run("negative: unauthorized", func(t *testing.T) {
		manager := newMockDbManager(t)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Post("/api/user/orders", handler.LoadOrder)
		})

		srv := httptest.NewServer(r)
		defer srv.Close()

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").
			SetBody([]byte(`193892`)).
			Post(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "401 Unauthorized")
	})

	t.Run("negative: order was already created by the other user", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		manager.On("LoadOrder", "test", "614371538763429").Return(database.ErrCreatedDiffUser)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Post("/api/user/orders", handler.LoadOrder)
		})

		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").
			SetBody([]byte(orderNum)).
			SetHeader("Authorization", user.Header().Get("Authorization")).Post(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "409 Conflict")
	})
}

func TestHandler_GetOrders(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log := *logger.Sugar()

	t.Run("positive: success", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		manager.On("GetUserOrders", "test").Return([]byte(`[{"number":"1","uploaded_at":"2021-08-15T14:30:45.0000001+03:00","status":"NEW","accrual":100.5}]`), nil)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Get("/api/user/orders", handler.GetOrders)
		})
		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Authorization", user.Header().Get("Authorization")).
			Get(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "200 OK")
	})
	t.Run("positive: no data", func(t *testing.T) {
		manager := newMockDbManager(t)
		manager.On("Register", "test", "test").Return(nil)
		manager.On("Login", "test", "test").Return(nil)
		manager.On("GetUserOrders", "test").Return(nil, database.ErrNoData)

		handler := New(manager, &log)
		r := chi.NewRouter()
		r.Group(func(r chi.Router) {
			r.Post("/api/user/register", handler.Register)
			r.Post("/api/user/login", handler.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.BasicAuth)
			r.Get("/api/user/orders", handler.GetOrders)
		})
		srv := httptest.NewServer(r)
		defer srv.Close()

		user, err := resty.New().R().
			SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
			Post(fmt.Sprintf("%s/api/user/register", srv.URL))
		assert.NoError(t, err)

		response, err := resty.New().R().
			SetHeader("Authorization", user.Header().Get("Authorization")).
			Get(fmt.Sprintf("%s/api/user/orders", srv.URL))

		assert.NoError(t, err)
		assert.Equal(t, response.Status(), "204 No Content")
	})
}

func TestHandler_Withdraw(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log := *logger.Sugar()

	testCases := []struct {
		name           string
		balance        float64
		order          string
		withdraw       float64
		expectedStatus string
		errDB          error
	}{
		{
			name:           "positive: success withdraw",
			order:          "2377225624",
			balance:        55,
			withdraw:       20,
			expectedStatus: "200 OK",
		},
		{
			name:           "negative: insufficient balance",
			order:          "2377225624",
			balance:        20,
			withdraw:       55,
			expectedStatus: "402 Payment Required",
			errDB:          database.ErrInsufficientBalance,
		},
		{
			name:           "negative: bad order num",
			order:          "123",
			balance:        55,
			withdraw:       20,
			expectedStatus: "422 Unprocessable Entity",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			manager := newMockDbManager(t)
			manager.On("Register", "test", "test").Return(nil)
			manager.On("Login", "test", "test").Return(nil)
			if tt.expectedStatus != "422 Unprocessable Entity" {
				manager.On("Withdraw", "test", tt.order, tt.withdraw).Return(tt.errDB)
			}

			handler := New(manager, &log)
			r := chi.NewRouter()
			r.Group(func(r chi.Router) {
				r.Post("/api/user/register", handler.Register)
				r.Post("/api/user/login", handler.Login)
			})
			r.Group(func(r chi.Router) {
				r.Use(handler.BasicAuth)
				r.Post("/api/user/balance/withdraw", handler.Withdraw)
			})
			srv := httptest.NewServer(r)
			defer srv.Close()

			user, err := resty.New().R().
				SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
				Post(fmt.Sprintf("%s/api/user/register", srv.URL))
			assert.NoError(t, err)

			response, err := resty.New().R().
				SetHeader("Authorization", user.Header().Get("Authorization")).
				SetBody(fmt.Sprintf(`{"order": %q, "sum": %f}`, tt.order, tt.withdraw)).
				Post(fmt.Sprintf("%s/api/user/balance/withdraw", srv.URL))

			assert.NoError(t, err)
			assert.Equal(t, response.Status(), tt.expectedStatus)
		})
	}
}

func TestHandler_GetBalance(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log := *logger.Sugar()

	testCases := []struct {
		name           string
		balanceFromDB  string
		dbErr          error
		expectedStatus string
	}{
		{
			name:           "positive",
			balanceFromDB:  `{"current": 500.5,"withdrawn": 42}`,
			expectedStatus: "200 OK",
		},
		{
			name:           "negative: db err",
			dbErr:          errors.New("db error"),
			expectedStatus: "500 Internal Server Error",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			manager := newMockDbManager(t)
			manager.On("Register", "test", "test").Return(nil)
			manager.On("Login", "test", "test").Return(nil)
			manager.On("GetBalanceInfo", "test").Return([]byte(tt.balanceFromDB), tt.dbErr)

			handler := New(manager, &log)
			r := chi.NewRouter()
			r.Group(func(r chi.Router) {
				r.Post("/api/user/register", handler.Register)
				r.Post("/api/user/login", handler.Login)
			})
			r.Group(func(r chi.Router) {
				r.Use(handler.BasicAuth)
				r.Get("/api/user/balance", handler.GetBalance)
			})
			srv := httptest.NewServer(r)
			defer srv.Close()

			user, err := resty.New().R().
				SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
				Post(fmt.Sprintf("%s/api/user/register", srv.URL))
			assert.NoError(t, err)

			response, err := resty.New().R().
				SetHeader("Authorization", user.Header().Get("Authorization")).
				Get(fmt.Sprintf("%s/api/user/balance", srv.URL))

			assert.NoError(t, err)
			assert.Equal(t, response.Status(), tt.expectedStatus)
			if tt.dbErr == nil {
				assert.Equal(t, response.String(), tt.balanceFromDB)
			}
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	log := *logger.Sugar()

	testCases := []struct {
		name           string
		withdrawals    string
		dbErr          error
		expectedStatus string
	}{
		{
			name:           "positive",
			withdrawals:    `[{"order": "2377225624","sum": 500,"processed_at": "2020-12-09T16:09:57+03:00"}]`,
			expectedStatus: "200 OK",
		},
		{
			name:           "positive: no withdrawals",
			dbErr:          database.ErrNoData,
			expectedStatus: "204 No Content",
		},
		{
			name:           "negative: db err",
			dbErr:          errors.New("db error"),
			expectedStatus: "500 Internal Server Error",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			manager := newMockDbManager(t)
			manager.On("Register", "test", "test").Return(nil)
			manager.On("Login", "test", "test").Return(nil)
			manager.On("GetWithdrawals", "test").Return([]byte(tt.withdrawals), tt.dbErr)

			handler := New(manager, &log)
			r := chi.NewRouter()
			r.Group(func(r chi.Router) {
				r.Post("/api/user/register", handler.Register)
				r.Post("/api/user/login", handler.Login)
			})
			r.Group(func(r chi.Router) {
				r.Use(handler.BasicAuth)
				r.Get("/api/user/withdrawals", handler.GetWithdrawals)
			})
			srv := httptest.NewServer(r)
			defer srv.Close()

			user, err := resty.New().R().
				SetHeader("Content-Type", "text/plain").SetBody(`{"login": "test", "password": "test"}`).
				Post(fmt.Sprintf("%s/api/user/register", srv.URL))
			assert.NoError(t, err)

			response, err := resty.New().R().
				SetHeader("Authorization", user.Header().Get("Authorization")).
				Get(fmt.Sprintf("%s/api/user/withdrawals", srv.URL))

			assert.NoError(t, err)
			assert.Equal(t, response.Status(), tt.expectedStatus)
			if tt.dbErr == nil {
				assert.Equal(t, response.String(), tt.withdrawals)
			}
		})
	}
}
