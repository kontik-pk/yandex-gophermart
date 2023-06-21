package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (m *Manager) GetBalanceInfo(login string) ([]byte, error) {
	userBalance, err := m.getUserBalance(login)
	if err != nil {
		return nil, fmt.Errorf("error while getting current user userBalance: %w", err)
	}
	const getUserWithdrawn = "select sum(amount) as withdrawn from withdraw where login = $1"
	row := m.db.QueryRow(getUserWithdrawn, login)
	var userWithdrawn sql.NullFloat64
	if err = row.Scan(&userWithdrawn); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error while getting user withdrawn info: %w", err)
		}
		userWithdrawn = sql.NullFloat64{
			Float64: 0,
			Valid:   true,
		}
	}
	info := models.BalanceInfo{
		Withdrawn: userWithdrawn.Float64,
		Current:   userBalance,
	}
	result, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling user balance info: %w", err)
	}
	return result, nil
}

func (m *Manager) GetWithdrawals(login string) ([]byte, error) {
	const getUserWithdrawals = `select order_id, amount, processed_at from withdraw where login = $1 order by processed_at`
	rows, err := m.db.Query(getUserWithdrawals, login)
	if err != nil {
		return nil, fmt.Errorf("error while searching for userWithdrawals: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	userWithdrawals := make([]models.WithdrawInfo, 0)
	for rows.Next() {
		var (
			orderID     string
			amount      float64
			processedAt time.Time
		)
		if err = rows.Scan(&orderID, &amount, &processedAt); err != nil {
			return nil, fmt.Errorf("error while scanning rows from userWithdrawals: %w", err)
		}
		userWithdrawals = append(userWithdrawals, models.WithdrawInfo{
			OrderID:     orderID,
			ProcessedAt: &processedAt,
			Amount:      amount,
		})
	}
	if len(userWithdrawals) == 0 {
		return nil, ErrNoData
	}
	result, err := json.Marshal(userWithdrawals)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling user withdrawals info: %w", err)
	}
	return result, nil
}

func (m *Manager) Withdraw(login string, orderID string, sum float64) error {
	userBalance, err := m.getUserBalance(login)
	if err != nil {
		return fmt.Errorf("error while checking user userBalance: %w", err)
	}
	if userBalance < sum {
		return ErrInsufficientBalance
	}
	const withdraw = "insert into withdraw values ($1, $2, now(), $3)"
	if _, err = m.db.Exec(withdraw, login, orderID, sum); err != nil {
		return fmt.Errorf("error while trying to withdraw: %w", err)
	}
	return nil
}

func (m *Manager) GetUserOrders(login string) ([]byte, error) {
	const getUserOrders = `select order_id, status, accrual, uploaded_at from orders where login = $1`
	rows, err := m.db.Query(getUserOrders, login)
	if err != nil {
		return nil, fmt.Errorf("error while getting orders from db for user %q: %w", login, err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	userOrders := make([]models.OrderInfo, 0)
	for rows.Next() {
		var (
			orderID    string
			status     models.OrderStatus
			accrual    sql.NullFloat64
			uploadedAt time.Time
		)
		if err = rows.Scan(&orderID, &status, &accrual, &uploadedAt); err != nil {
			return nil, fmt.Errorf("error while scanning rows: %w", err)
		}
		userOrders = append(userOrders, models.OrderInfo{
			OrderID:   orderID,
			Accrual:   accrual.Float64,
			CreatedAt: &uploadedAt,
			Status:    status,
		})
	}
	if len(userOrders) == 0 {
		return nil, ErrNoData
	}
	result, err := json.Marshal(userOrders)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling user orders info: %w", err)
	}
	return result, nil
}
func (m *Manager) GetAllOrders() ([]string, error) {
	const getAllOrders = `select order_id from orders`
	rows, err := m.db.Query(getAllOrders)
	if err != nil {
		return nil, fmt.Errorf("error while getting all orders from db: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	orders := make([]string, 0)
	for rows.Next() {
		var orderID string
		if err = rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("error while scanning rows: %w", err)
		}
		orders = append(orders, orderID)
	}
	return orders, nil
}

func (m *Manager) UpdateOrderInfo(orderInfo *models.OrderInfo) error {
	const updateOrderInfo = `update orders set status=$1, accrual=$2 where order_id=$3`
	if _, err := m.db.Exec(updateOrderInfo, string(orderInfo.Status), orderInfo.Accrual, orderInfo.Order); err != nil {
		return fmt.Errorf("error while updating order info: %w", err)
	}
	return nil
}

func (m *Manager) LoadOrder(login string, orderID string) error {
	const getOrderByID = `select login from orders where order_id = $1`
	row := m.db.QueryRow(getOrderByID, orderID)

	var userName string
	err := row.Scan(&userName)
	switch err {
	case sql.ErrNoRows:
		const loadOrderQuery = `insert into orders values ($1, $2, now(), $3, $4)`
		if _, err = m.db.Exec(loadOrderQuery, orderID, login, models.OrderStatus("NEW"), 0); err != nil {
			return fmt.Errorf("error while loading order %s: %w", orderID, err)
		}
		return nil
	case nil:
		if userName == login {
			return ErrCreatedBySameUser
		}
		return ErrCreatedDiffUser
	default:
		return fmt.Errorf("error while scanning rows: %w", err)
	}
}

func (m *Manager) Register(login string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("this password is not allowed: %w", err)
	}
	const registerUser = `insert into registered_users values ($1, $2)`
	if _, err = m.db.Exec(registerUser, login, hash); err != nil {
		dublicateKeyErr := ErrDublicateKey{Key: "registered_users_pkey"}
		if err.Error() == dublicateKeyErr.Error() {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("error while executing register user query: %w", err)
	}
	return nil
}

func (m *Manager) Login(login string, password string) error {
	const getRegisteredUser = `select login, password from registered_users`
	rows, err := m.db.Query(getRegisteredUser)
	if err != nil {
		return fmt.Errorf("error while executing search query: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var loginFromDB, passwordFromDB string
		if err = rows.Scan(&loginFromDB, &passwordFromDB); err != nil {
			return fmt.Errorf("error while scanning rows: %w", err)
		}
		if loginFromDB == login {
			if err = bcrypt.CompareHashAndPassword([]byte(passwordFromDB), []byte(password)); err != nil {
				return ErrInvalidCredentials
			}
			return nil
		}
	}
	return ErrNoSuchUser
}

func (m *Manager) getUserBalance(login string) (float64, error) {
	const getUserBalance = "select coalesce(sum(accrual), 0) - coalesce(sum(amount), 0) as balance from orders o left join withdraw w on o.login = w.login where o.login = $1 group by o.login;"
	row := m.db.QueryRow(getUserBalance, login)
	var balance sql.NullFloat64
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error while getting user balance: %w", err)
	}
	return balance.Float64, nil
}

func (m *Manager) init(ctx context.Context) error {
	const createRegisteredQuery = `create table if not exists registered_users (login text primary key, password text)`
	if _, err := m.db.ExecContext(ctx, createRegisteredQuery); err != nil {
		return fmt.Errorf("error while trying to create table with registered users: %w", err)
	}
	const createOrdersQuery = `create table if not exists orders (order_id text unique, login text, uploaded_at timestamp with time zone, status text, accrual double precision, primary key(order_id))`
	if _, err := m.db.ExecContext(ctx, createOrdersQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	const createWithdrawQuery = `create table if not exists withdraw (login text, order_id text unique, processed_at timestamp with time zone, amount double precision, primary key(login, order_id))`
	if _, err := m.db.ExecContext(ctx, createWithdrawQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	return nil
}

func New(ctx context.Context, db *sql.DB) (*Manager, error) {
	m := Manager{
		db: db,
	}
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}

type Manager struct {
	db *sql.DB
}
