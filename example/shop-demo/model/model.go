package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type OrderStatus int

const (
	OrderStatusPending  OrderStatus = 0
	OrderStatusPaid     OrderStatus = 1
	OrderStatusClosed   OrderStatus = 2
	OrderStatusRefunded OrderStatus = 3
)

type User struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Balance   float64        `gorm:"default:0" json:"balance"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type RechargeOrder struct {
	ID          uint64      `gorm:"primaryKey" json:"id"`
	OrderNo     string      `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	UserID      uint64      `gorm:"index;not null" json:"user_id"`
	Username    string      `gorm:"size:50;not null" json:"username"`
	PackageID   string      `gorm:"size:32;not null" json:"package_id"`
	PayAmount   float64     `gorm:"not null" json:"pay_amount"`
	BonusAmount float64     `gorm:"default:0" json:"bonus_amount"`
	Status      OrderStatus `gorm:"default:0" json:"status"`
	PaidAt      *time.Time  `json:"paid_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (RechargeOrder) TableName() string {
	return "recharge_orders"
}

func (o *RechargeOrder) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

func (o *RechargeOrder) TotalAmount() float64 {
	return o.PayAmount + o.BonusAmount
}

type BalanceLogType int

const (
	BalanceTypeRecharge BalanceLogType = 1
	BalanceTypeConsume  BalanceLogType = 2
	BalanceTypeRefund   BalanceLogType = 3
)

type BalanceLog struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	UserID    uint64         `gorm:"index;not null" json:"user_id"`
	Username  string         `gorm:"size:50;not null" json:"username"`
	Type      BalanceLogType `gorm:"not null" json:"type"`
	Amount    float64        `gorm:"not null" json:"amount"`
	Balance   float64        `gorm:"not null" json:"balance"`
	OrderNo   string         `gorm:"size:64" json:"order_no"`
	Remark    string         `gorm:"size:255" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
}

func (BalanceLog) TableName() string {
	return "balance_logs"
}

type RechargePackage struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PayAmount   float64 `json:"pay_amount"`
	BonusAmount float64 `json:"bonus_amount"`
	Sort        int     `json:"sort"`
}

func (p RechargePackage) TotalAmount() float64 {
	return p.PayAmount + p.BonusAmount
}

var DefaultPackages = []RechargePackage{
	{ID: "pkg_10", Name: "充值 10 元", PayAmount: 10, BonusAmount: 0, Sort: 1},
	{ID: "pkg_30", Name: "充值 30 元", PayAmount: 30, BonusAmount: 2, Sort: 2},
	{ID: "pkg_50", Name: "充值 50 元", PayAmount: 50, BonusAmount: 5, Sort: 3},
	{ID: "pkg_100", Name: "充值 100 元", PayAmount: 100, BonusAmount: 15, Sort: 4},
	{ID: "pkg_200", Name: "充值 200 元", PayAmount: 200, BonusAmount: 40, Sort: 5},
}

func GetPackageByID(id string) *RechargePackage {
	for i := range DefaultPackages {
		if DefaultPackages[i].ID == id {
			return &DefaultPackages[i]
		}
	}
	return nil
}

func NewCustomPackage(amount float64) *RechargePackage {
	return &RechargePackage{
		ID:          fmt.Sprintf("custom_%.0f", amount),
		Name:        fmt.Sprintf("自定义充值 %.0f 元", amount),
		PayAmount:   amount,
		BonusAmount: 0,
		Sort:        999,
	}
}
