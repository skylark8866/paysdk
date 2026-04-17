package repo

import (
	"shop-demo/model"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dbPath string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.User{}, &model.RechargeOrder{}, &model.BalanceLog{}); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateOrder(order *model.RechargeOrder) error {
	return r.db.Create(order).Error
}

func (r *Repository) GetOrderByOrderNo(orderNo string) (*model.RechargeOrder, error) {
	var order model.RechargeOrder
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) MarkOrderPaid(orderNo string, paidAt time.Time) error {
	return r.db.Model(&model.RechargeOrder{}).
		Where("order_no = ? AND status = ?", orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":  model.OrderStatusPaid,
			"paid_at": paidAt,
		}).Error
}

func (r *Repository) GetUserOrders(userID uint64, limit int) ([]model.RechargeOrder, error) {
	var orders []model.RechargeOrder
	err := r.db.Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

func (r *Repository) AddBalanceLog(log *model.BalanceLog) error {
	return r.db.Create(log).Error
}

func (r *Repository) GetUserBalanceLogs(userID uint64, limit int) ([]model.BalanceLog, error) {
	var logs []model.BalanceLog
	err := r.db.Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *Repository) AddUserBalance(userID uint64, amount float64) (float64, error) {
	var user model.User
	err := r.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount)).
		First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (r *Repository) ProcessPayment(orderNo string, paidAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var order model.RechargeOrder
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}

		if order.IsPaid() {
			return nil
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":  model.OrderStatusPaid,
			"paid_at": paidAt,
		}).Error; err != nil {
			return err
		}

		totalAmount := order.TotalAmount()
		var user model.User
		if err := tx.Model(&model.User{}).
			Where("id = ?", order.UserID).
			Update("balance", gorm.Expr("balance + ?", totalAmount)).
			First(&user).Error; err != nil {
			return err
		}

		log := &model.BalanceLog{
			UserID:   order.UserID,
			Username: order.Username,
			Type:     model.BalanceTypeRecharge,
			Amount:   totalAmount,
			Balance:  user.Balance,
			OrderNo:  order.OrderNo,
			Remark:   "充值成功",
		}
		return tx.Create(log).Error
	})
}
