package models

import (
	"errors"
	"time"
	
	"github.com/jinzhu/gorm"
	"github.com/xectich/paymentGateway/constants"
)

// generic information about the Bank Account
type BankAccount struct {
	ID                uint32    `gorm:"primary_key;auto_increment" json:"id"`
	CardID            string    `gorm:"size:100;not null;unique" json:"cardId"`
	Balance           float64   `gorm:"not null;" json:"balance"`
	BalanceAuthorised float64   `gorm:"not null;" json:"balanceAuthorised"`
	Currency          string    `gorm:"size:4;not null;" json:"currency"`
	CreatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

//Interface to call Bank Account functions
type BankAccountI interface {
	CreateBankAccount(ba *BankAccount, db *gorm.DB) (*BankAccount, error)
	FindBankAccountByCardID(db *gorm.DB, id string) (*BankAccount, error)
	AuthorizeBalance(db *gorm.DB, id string, amount float64) error
	RefundBalance(db *gorm.DB, id string, amount float64) error
	CaptureBalance(db *gorm.DB, id string, amount float64) error
	VoidAuthorization(db *gorm.DB, id string) error
}

func NewBankAccountI() BankAccountI {
	return &BankAccount{}
}

//CreateBankAccount stores a new BankAccount to the DB
func (b *BankAccount) CreateBankAccount(bankAccount *BankAccount, db *gorm.DB) (ba *BankAccount, er error) {
	var err error
	err = db.Debug().Create(&bankAccount).Error
	if err != nil {
		return &BankAccount{}, err
	}
	return bankAccount, nil
}

//FindBankAccountByCardID retrieves a BA by CardID from the DB
func (b *BankAccount) FindBankAccountByCardID(db *gorm.DB, cardId string) (ba *BankAccount, er error) {
	var err error
	var bankAccount BankAccount
	err = db.Debug().Model(BankAccount{}).Where("card_id = ?", cardId).Take(&bankAccount).Error
	if err != nil {
		return &BankAccount{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &BankAccount{}, errors.New(constants.BankAccountNotFound)
	}
	return &bankAccount, err
}

//AuthorizeBalance authrozises the requested balance if possible
func (b *BankAccount) AuthorizeBalance(db *gorm.DB, cardID string, amount float64) (err error) {
	ba, err := b.FindBankAccountByCardID(db, cardID)
	if err != nil {
		return err
	}

	if amount == 0 {
		return errors.New(constants.InvalidAmount)
	}
		
	if amount > ba.Balance {
		return errors.New(constants.AmountExeedsBalance)
	}

	db = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&BankAccount{}).UpdateColumns(
		map[string]interface{}{
			"balanceAuthorised": amount,
			"updated_at":         time.Now(),
		},
	)

	if db.Error != nil {
		return db.Error
	}

	err = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&ba).Error
	if err != nil {
		return err
	}

	return err
}

//RefundBalance refunds the amount and updates the authorized balance
func (b *BankAccount) RefundBalance(db *gorm.DB, cardID string, amount float64) (err error) {
	ba, err := b.FindBankAccountByCardID(db, cardID)
	if err != nil {
		return err
	}

	if ba.BalanceAuthorised < amount {
		return errors.New(constants.AmountExeedsAuthorizedBalance)
	}

	db = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&BankAccount{}).UpdateColumns(
		map[string]interface{}{
			"balance":           ba.Balance + amount,
			"balanceAuthorised": ba.BalanceAuthorised - amount,
			"updated_at":         time.Now(),
		},
	)

	if db.Error != nil {
		return db.Error
	}

	err = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&ba).Error
	if err != nil {
		return err
	}

	return err
}

//RefundBalance refunds the amount and updates the authorized balance
func (b *BankAccount) CaptureBalance(db *gorm.DB, cardID string, amount float64) (err error) {
	ba, err := b.FindBankAccountByCardID(db, cardID)
	if err != nil {
		return err
	}

	if ba.BalanceAuthorised < amount {
		return errors.New(constants.AmountExeedsAuthorizedBalance)
	}

	if ba.Balance < amount {
		return errors.New(constants.AmountExeedsBalance)
	}

	db = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&BankAccount{}).UpdateColumns(
		map[string]interface{}{
			"balance":           ba.Balance - amount,
			"balanceAuthorised": ba.BalanceAuthorised - amount,
			"updated_at":         time.Now(),
		},
	)

	if db.Error != nil {
		return db.Error
	}

	err = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&ba).Error
	if err != nil {
		return err
	}

	return err
}

//VoidAuthorization voids the authorization by reseting the authorized amount
func (b *BankAccount) VoidAuthorization(db *gorm.DB, cardID string) (err error) {
	ba, err := b.FindBankAccountByCardID(db, cardID)
	if err != nil {
		return err
	}

	db = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&BankAccount{}).UpdateColumns(
		map[string]interface{}{
			"balanceAuthorised": 0,
			"updated_at":         time.Now(),
		},
	)

	if db.Error != nil {
		return db.Error
	}

	err = db.Debug().Model(&BankAccount{}).Where("card_id = ?", cardID).Take(&ba).Error
	if err != nil {
		return err
	}

	return err
}
