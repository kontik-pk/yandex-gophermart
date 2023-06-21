package models

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderStatus string

type OrderInfo struct {
	UserName  *string     `json:"user,omitempty"`
	OrderID   string      `json:"number"`
	Order     *string     `json:"order,omitempty"`
	CreatedAt *time.Time  `json:"uploaded_at,omitempty"`
	Status    OrderStatus `json:"status"`
	Accrual   float64     `json:"accrual"`
}

type WithdrawInfo struct {
	UserName    *string    `json:"user,omitempty"`
	OrderID     string     `json:"order"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Amount      float64    `json:"sum"`
}

type BalanceInfo struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"login"`
}

type Option func(params *Params)

type Params struct {
	ServerRunAddr        string
	DatabaseAddress      string
	AccrualSystemAddress string
}
