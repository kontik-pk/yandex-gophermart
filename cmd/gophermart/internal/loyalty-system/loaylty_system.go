package loyalty

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/kontik-pk/go-musthave-diploma-tpl/cmd/gophermart/internal/models"
	"go.uber.org/zap"
)

func (ls *LoyaltySystem) UpdateOrdersInfo() error {
	allOrders, err := ls.db.GetAllOrders()
	if err != nil {
		return fmt.Errorf("error while getting all orders from db for updating info: %w", err)
	}
	for _, o := range allOrders {
		actualInfo, err := ls.getActualInfo(o)
		if err != nil {
			return fmt.Errorf("error while getting actual info for order %q: %w", o, err)
		}
		if err = ls.db.UpdateOrderInfo(actualInfo); err != nil {
			return fmt.Errorf("error while updating order info: %w", err)
		}
		ls.log.Infof("order %q updated with accrual: %f", *actualInfo.Order, actualInfo.Accrual)
	}
	return nil
}

func (ls *LoyaltySystem) getActualInfo(orderID string) (*models.OrderInfo, error) {
	orderFromSystem, err := resty.New().R().Get(fmt.Sprintf("%s/api/orders/%s", ls.addr, orderID))
	if err != nil {
		return nil, fmt.Errorf("error while requesting for order %q: %w", orderID, err)
	}
	var info models.OrderInfo
	if err = json.Unmarshal(orderFromSystem.Body(), &info); err != nil {
		return nil, fmt.Errorf("error while unmarshalling order body: %w", err)
	}
	return &info, nil
}

func New(addr string, db dbManager, logger *zap.SugaredLogger) *LoyaltySystem {
	return &LoyaltySystem{
		addr: addr,
		db:   db,
		log:  logger,
	}
}

type LoyaltySystem struct {
	addr string
	db   dbManager
	log  *zap.SugaredLogger
}

type dbManager interface {
	GetAllOrders() ([]string, error)
	UpdateOrderInfo(orderInfo *models.OrderInfo) error
}
