package models

import (
	"errors"
	"time"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
	"github.com/xectich/paymentGateway/constants"
)

// generic information about the authorization request
type AuthorizationRequest struct {
	CardNumber      string  `json:"cardNumber"`
	Currency        string  `json:"currency"`
	CVV             string  `json:"cvv"`
	Amount          float64 `json:"amount"`
	ExpirationMonth int     `json:"expirationMonth"`
	ExpirationYear  int     `json:"expirationYear"`
}

// generic information about the authorization response
type AuthorizationResponse struct {
	ID              string  `json:"id"`
	Currency        string  `json:"currency"`
	AmountAvailable float64 `json:"amountAvailable"`
}

// generic information about the action request
type ActionRequest struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
	Final  bool    `json:"final"`
}

// generic information about the authorization
type Authorization struct {
	ID                string    `gorm:"primary_key;unique" json:"id"`
	CardNumber        string    `gorm:"size:16;not null;unique" json:"cardNumber"`
	BalanceCaptured   float64   `gorm:"not null;" json:"balanceCaptured"`
	BalanceAuthorised float64   `gorm:"not null;" json:"balanceAuthorised"`
	BalanceRefunded   float64   `gorm:"not null;" json:"balanceRefunded"`
	CurrencyRequested string    `gorm:"size:4;not null;" json:"currencyRequested"`
	CurrencyCard      string    `gorm:"size:4;not null;" json:"currencyCard"`
	Status            string    `gorm:"size:16;not null;unique" json:"status"`
	CreatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type AuthorizationI interface {
	RequestAuthorization(authRequest AuthorizationRequest, db *gorm.DB) (*Authorization, error)
	Capture(authId string, amount float64, finalCapture bool, db *gorm.DB) (*Authorization, error)
	Void(authId string, db *gorm.DB) (*Authorization, error)
	Refund(authId string, amount float64, finalRefund bool, db *gorm.DB) (*Authorization, error)
	FindAuthorizationByID(authId string, db *gorm.DB) (*Authorization, error)
}

func NewAuthI() AuthorizationI {
	return &Authorization{}
}

//RequestAuthorization performs necessary checks and authorizes the amount on the customers Bank based on the auth request data
func (a *Authorization) RequestAuthorization(authRequest AuthorizationRequest, db *gorm.DB) (auth *Authorization, err error) {
	cardI := NewCardI()
	bankI := NewBankAccountI() //Assumption: normally this will be done by invoking an API in a real life scenario to update the BankAccount

	if authRequest.Amount <= 0 {
		return &Authorization{}, errors.New(constants.InvalidAmount)
	}

	if !cardI.ValidateLuhnNumber(authRequest.CardNumber) {
		return &Authorization{}, errors.New(constants.InvalidCreditCardNumber)
	}

	card, err := cardI.FindCardByNumber(db, authRequest.CardNumber)
	if err != nil {
		return &Authorization{}, err
	}

	if err = cardI.Validate(card, authRequest.CVV, authRequest.ExpirationMonth, authRequest.ExpirationYear); err != nil {
		return &Authorization{}, err
	}

	if card.Currency != authRequest.Currency {
		authRequest.Amount,err = convertCurrency(authRequest.Amount, authRequest.Currency, card.Currency)
		if err != nil {
			return &Authorization{}, err
		}
	}

	if err = bankI.AuthorizeBalance(db, card.Number, authRequest.Amount); err != nil {
		return &Authorization{}, err
	}

	authorization := Authorization{
		ID:                ksuid.New().String(),
		CardNumber:        card.Number,
		BalanceAuthorised: authRequest.Amount,
		BalanceCaptured:   0,
		CurrencyRequested: authRequest.Currency,
		CurrencyCard:      card.Currency,
		Status:            constants.AuthStatus(constants.Authorized).String(),
	}

	if err = db.Debug().Create(&authorization).Error; err != nil {
		return &Authorization{}, err
	}

	return &authorization, nil
}

//Capture performs necessary checks and captures the amount on the customers Bank based on the request amount
func (a *Authorization) Capture(authId string, amount float64, finalCapture bool, db *gorm.DB) (auth *Authorization, err error) {
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	if auth.Status != constants.AuthStatus(constants.Authorized).String() {
		return &Authorization{}, errors.New(constants.InvalidStatus + auth.Status)
	}

	//verify if amount is valid
	if amount <= 0 || (amount > auth.BalanceAuthorised || (amount+auth.BalanceCaptured) > auth.BalanceAuthorised) {
		return &Authorization{}, errors.New(constants.InvalidAmount)
	}

	//capture the amount on the customers bank
	bankI := NewBankAccountI()
	if err = bankI.CaptureBalance(db, auth.CardNumber, amount); err != nil {
		return &Authorization{}, err

	}

	status := auth.Status
	if finalCapture || amount == auth.BalanceAuthorised {
		status = constants.AuthStatus(constants.Captured).String()
	}

	//update the auth object in db
	db = db.Debug().Model(&BankAccount{}).Where("id = ?", authId).Take(&Authorization{}).UpdateColumns(
		map[string]interface{}{
			"balanceCaptured": auth.BalanceCaptured + amount,
			"status":          status,
			"updated_at":       time.Now(),
		},
	)

	//refresh auth object
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	return auth, nil
}

//Void voids the authorization by chaning the status to void
func (a *Authorization) Void(authId string, db *gorm.DB) (auth *Authorization, err error) {
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	if auth.Status != constants.AuthStatus(constants.Authorized).String() {
		return &Authorization{}, errors.New(constants.InvalidStatus + auth.Status)
	}

	bankI := NewBankAccountI()
	if err = bankI.VoidAuthorization(db, auth.CardNumber); err != nil {
		return &Authorization{}, err
	}

	status := constants.AuthStatus(constants.Voided).String()
	//update the auth object in db
	db = db.Debug().Model(&BankAccount{}).Where("id = ?", authId).Take(&Authorization{}).UpdateColumns(
		map[string]interface{}{
			"status":    status,
			"updated_at": time.Now(),
		},
	)

	//refresh auth object
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	return auth, nil
}

//Refund refunds the specified amount to the customer if it does not exceed the captured balance
func (a *Authorization) Refund(authId string, amount float64, finalRefund bool, db *gorm.DB) (auth *Authorization, err error) {
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	if auth.Status != constants.AuthStatus(constants.Authorized).String() {
		return &Authorization{}, errors.New(constants.InvalidStatus + auth.Status)
	}

	//verify if amount is valid, checking againts authrozed balance to cpature edge case
	if amount <= 0 ||
		(amount > auth.BalanceAuthorised ||
			(amount+auth.BalanceRefunded) > auth.BalanceAuthorised ||
			(amount+auth.BalanceRefunded) > auth.BalanceCaptured) {
		return &Authorization{}, errors.New(constants.InvalidAmount)
	}

	//capture the amount on the customers bank
	bankI := NewBankAccountI()
	if err = bankI.RefundBalance(db, auth.CardNumber, amount); err != nil {
		return &Authorization{}, err
	}

	status := auth.Status
	if finalRefund || amount == auth.BalanceAuthorised {
		status = constants.AuthStatus(constants.Refunded).String()
	}

	//update the auth object in db
	db = db.Debug().Model(&BankAccount{}).Where("id = ?", authId).Take(&Authorization{}).UpdateColumns(
		map[string]interface{}{
			"balanceCaptured": auth.BalanceCaptured - amount,
			"balanceRefunded": auth.BalanceRefunded + amount,
			"status":          status,
			"updated_at":      time.Now(),
		},
	)

	//refresh auth object
	auth, err = a.FindAuthorizationByID(authId, db)
	if err != nil {
		return &Authorization{}, err
	}

	return auth, nil
}

//FindAuthorizationByID retrieves an Authorization by ID from the DB
func (a *Authorization) FindAuthorizationByID(authId string, db *gorm.DB) (auth *Authorization, err error) {
	var authorization Authorization
	err = db.Debug().Model(Authorization{}).Where("id = ?", authId).Take(&authorization).Error
	if err != nil {
		return &Authorization{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Authorization{}, errors.New(constants.AuthorizationNotFound)
	}
	return &authorization, err
}

//convertCurrency converts between the authorization request currency and the card currency
//returns the converted amount (could round it)
func convertCurrency(amount float64, baseCur, targetCur string) (amt float64, err error) {

	curConvert := baseCur + "_" + targetCur

	url := "https://free.currconv.com/api/v7/convert?q=" + curConvert + "&compact=ultra&apiKey=424d5c06590ef708b468"

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return amount, err
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")

	res, err := spaceClient.Do(req)
	if err != nil {
		return amount, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return amount, err
	}

	convert := make(map[string]float32)

	err = json.Unmarshal(body, &convert)
	if err != nil {
		return amount, err
	}

	value := convert[curConvert]

	amount = amount * float64(value)

	return amount, nil
}
