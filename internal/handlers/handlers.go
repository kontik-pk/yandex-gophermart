package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/database"
	"github.com/kontik-pk/go-musthave-diploma-tpl/internal/models"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var jwtKey = []byte("my_secret_key")

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	login, status := h.getUsernameFromToken(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	userBalance, err := h.db.GetBalanceInfo(login)
	if err != nil {
		h.log.Errorf("error while getting user balance from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userBalance)
}

func (h *handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	login, status := h.getUsernameFromToken(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	userWithdrawals, err := h.db.GetWithdrawals(login)
	if err != nil {
		if errors.Is(err, database.ErrNoData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.log.Errorf("error while getting withdrawals from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userWithdrawals)
}

func (h *handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var withdrawInfo *models.WithdrawInfo
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &withdrawInfo); err != nil {
		h.log.Errorf("error while unmarshalling request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !h.checkOrder(withdrawInfo.OrderID) {
		h.log.Error("invalid order format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	login, status := h.getUsernameFromToken(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	if err := h.db.Withdraw(login, withdrawInfo.OrderID, withdrawInfo.Amount); err != nil {
		if errors.Is(err, database.ErrInsufficientBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		h.log.Errorf("error while trying to withdraw %f from user %q: %s", withdrawInfo.Amount, login, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.log.Infof("withdrawn %f from user %q for order %q", withdrawInfo.Amount, login, withdrawInfo.OrderID)
}

func (h *handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	login, status := h.getUsernameFromToken(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	userOrders, err := h.db.GetUserOrders(login)
	if err != nil {
		if errors.Is(err, database.ErrNoData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.log.Errorf("error while getting orders from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userOrders)
}

func (h *handler) LoadOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	var data bytes.Buffer
	if _, err := data.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	login, status := h.getUsernameFromToken(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	order := data.String()
	if !h.checkOrder(order) {
		h.log.Error("invalid order format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if err := h.db.LoadOrder(login, order); err != nil {
		if errors.Is(err, database.ErrCreatedBySameUser) {
			h.log.Info(fmt.Sprintf("order %q was alredy created by the same user", order))
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, database.ErrCreatedDiffUser) {
			h.log.Info(fmt.Sprintf("order %q was alredy created by the other user", order))
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Errorf("error while loading order to db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	user, success := h.parseInputUser(r)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.db.Login(user.Login, user.Password); err != nil {
		h.log.Errorf("error while login user: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(time.Hour)
	token, err := createToken(user.Login, expirationTime)
	if err != nil {
		h.log.Errorf("error while create token for user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
	h.log.Info(fmt.Sprintf("user %q is successfully authorized", user.Login))
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	user, success := h.parseInputUser(r)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.db.Register(user.Login, user.Password); err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			h.log.Errorf("login is already taken: %s", err.Error())
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Errorf("error while register user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := h.db.Login(user.Login, user.Password); err != nil {
		h.log.Errorf("error while login user: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(time.Hour)
	token, err := createToken(user.Login, expirationTime)
	if err != nil {
		h.log.Errorf("error while create token for user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
	h.log.Info(fmt.Sprintf("user %q is successfully registered and authorized", user.Login))
}

func (h *handler) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			h.log.Errorf("token is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tkn, err := h.extractJwtToken(r)
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) ||
				errors.Is(err, jwt.ErrTokenExpired) ||
				errors.Is(err, ErrTokenIsEmpty) ||
				errors.Is(err, ErrNoToken) {
				h.log.Errorf(err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h.log.Errorf(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			h.log.Errorf("invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Add("Authorization", tokenHeader)
		next.ServeHTTP(w, r)
	})
}

func (h *handler) extractJwtToken(r *http.Request) (*jwt.Token, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		h.log.Errorf("token is empty")
		return nil, ErrTokenIsEmpty
	}
	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		h.log.Errorf("no token")
		return nil, ErrNoToken
	}

	tknStr := splitted[1]
	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	return tkn, err
}

func (h *handler) parseInputUser(r *http.Request) (*models.User, bool) {
	var userFromRequest *models.User
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		return nil, false
	}
	if err := json.Unmarshal(buf.Bytes(), &userFromRequest); err != nil {
		h.log.Errorf("error while unmarshalling request body: %s", err.Error())
		return nil, false
	}
	if userFromRequest.Login == "" || userFromRequest.Password == "" {
		h.log.Errorf("login or password is empty")
		return nil, false
	}
	return userFromRequest, true
}

func (h *handler) checkOrder(orderID string) bool {
	orderAsInteger, err := strconv.Atoi(orderID)
	if err != nil {
		return false
	}
	number := orderAsInteger / 10
	luhn := 0
	for i := 0; number > 0; i++ {
		c := number % 10
		if i%2 == 0 {
			c *= 2
			if c > 9 {
				c = c%10 + c/10
			}
		}
		luhn += c
		number /= 10
	}
	return (orderAsInteger%10+luhn)%10 == 0
}

func (h *handler) getUsernameFromToken(r *http.Request) (string, int) {
	var data bytes.Buffer
	if _, err := data.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		return "", http.StatusBadRequest
	}
	tkn, err := h.extractJwtToken(r)
	if err != nil {
		h.log.Errorf("error while extracting token: %s", err.Error())
		return "", http.StatusInternalServerError
	}
	claims, ok := tkn.Claims.(*models.Claims)
	if !ok {
		h.log.Errorf("error while getting claims")
		return "", http.StatusInternalServerError
	}
	return claims.Username, http.StatusOK
}

func New(db dbManager, log *zap.SugaredLogger) *handler {
	return &handler{
		db:  db,
		log: log,
	}
}

type handler struct {
	db  dbManager
	log *zap.SugaredLogger
}

//go:generate mockery --disable-version-string --filename db_mock.go --inpackage --name dbManager
type dbManager interface {
	GetBalanceInfo(login string) ([]byte, error)
	GetWithdrawals(login string) ([]byte, error)
	Withdraw(login string, orderID string, sum float64) error
	GetUserOrders(login string) ([]byte, error)
	LoadOrder(login string, orderID string) error
	Register(login string, password string) error
	Login(login string, password string) error
}

func createToken(userName string, expirationTime time.Time) (string, error) {
	claims := &models.Claims{
		Username: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
