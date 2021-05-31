package models

//Haven't implemnted other methods to update cards, delete cards, wipe them, etc.
// as it was out of scope for the challenge and again to spare time

import (
	"errors"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xectich/paymentGateway/constants"
)

// generic information about the credit card
type Card struct {
	ID              uint32    `gorm:"primary_key;auto_increment" json:"id"`
	Number          string    `gorm:"size:100;not null;unique" json:"number"`
	CVV             string    `gorm:"size:4;not null;unique" json:"cvv"`
	Currency        string    `gorm:"size:4;not null;" json:"currency"`
	ExpirationMonth int       `gorm:"not null;" json:"expirationMonth"`
	ExpirationYear  int       `gorm:"not null;" json:"expirationYear"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

//Interface to call Credit Card functions
type CardI interface {
	Validate(card *Card, cvv string, expirationMonth, expirationYear int) error
	SaveCard(card *Card, db *gorm.DB) (*Card, error)
	FindCardByID(db *gorm.DB, id uint32) (*Card, error)
	FindCardByNumber(db *gorm.DB, number string) (*Card, error)
	ValidateLuhnNumber(cardNumber string) bool
}

func NewCardI() CardI {
	return &Card{}
}

// Validate returns an error if validation fails
// Checks for expiration date, CVV and the credit card's numbers when adding a existing.
func (c *Card) Validate(card *Card, cvv string, expirationMonth, expirationYear int) (err error) {
	if card.CVV != cvv {
		return errors.New(constants.NoMatchCVV)
	}

	if card.ExpirationMonth != expirationMonth || card.ExpirationYear != expirationYear {
		return errors.New(constants.NoMatchCardExpirationDate)
	}

	return nil
}

// ValidateNewCard returns an error if validation fails
// Checks for expiration date, CVV and the credit card's numbers when adding a new card.
func validateNewCard(card *Card) (err error) {
	var year, month int

	// Validate the expiration month
	month = card.ExpirationMonth
	if month < 1 || 12 < month {
		return errors.New(constants.InvalidMonth)
	}

	// Validate the expiration year
	year = card.ExpirationYear
	if year < time.Now().UTC().Year() {
		return errors.New(constants.CreditCardExpired)
	}

	// Check the expired  year and month
	if year == time.Now().UTC().Year() && month < int(time.Now().UTC().Month()) {
		return errors.New(constants.CreditCardExpired)
	}

	// Validate the CVV length
	if len(card.CVV) < 3 || len(card.CVV) > 4 {
		return errors.New(constants.InvalidCVV)
	}

	// Validate the Card number length
	if len(card.Number) < 13 {
		return errors.New(constants.InvalidCreditCardNumber)
	}

	// Valida the number using Luhn algorithm
	valid := card.ValidateLuhnNumber(card.Number)
	if !valid {
		return errors.New(constants.InvalidCreditCardNumber)
	}

	return nil
}

// ValidateLuhnNumber validates the Card number using the Luhn algorithm
func (c *Card) ValidateLuhnNumber(cardNumber string) bool {
	var sum int
	var alternate bool

	numberLen := len(cardNumber)

	// For numbers that is lower than 13 and
	// bigger than 19, must return as false
	if numberLen < 13 || numberLen > 19 {
		return false
	}

	// Traverse and verify card numbers
	for i := numberLen - 1; i > -1; i-- {
		//Mod the current number
		mod, _ := strconv.Atoi(string(cardNumber[i]))
		if alternate {
			mod *= 2
			if mod > 9 {
				mod = (mod % 10) + 1
			}
		}

		alternate = !alternate
		sum += mod
	}

	return sum%10 == 0
}

//SaveCard stores a new card to the DB
func (c *Card) SaveCard(card *Card, db *gorm.DB) (cc *Card, er error) {
	var err error
	if err := validateNewCard(card); err != nil {
		return &Card{}, err
	}

	err = db.Debug().Create(&card).Error
	if err != nil {
		return &Card{}, err
	}
	return card, nil
}

//FindCardByID retrieves a card by ID from the DB
func (c *Card) FindCardByID(db *gorm.DB, id uint32) (cc *Card, er error) {
	var err error
	var card Card
	err = db.Debug().Model(Card{}).Where("id = ?", id).Take(&card).Error
	if err != nil {
		return &Card{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Card{}, errors.New(constants.CardNotFound)
	}
	return &card, err
}

//FindCardByID retrieves a card by ID from the DB
func (c *Card) FindCardByNumber(db *gorm.DB, number string) (cc *Card, er error) {
	var err error
	var card Card
	err = db.Debug().Model(Card{}).Where("number = ?", number).Take(&card).Error
	if err != nil {
		return &Card{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Card{}, errors.New(constants.CardNotFound)
	}
	return &card, err
}
