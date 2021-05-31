package tests

import (
	"testing"
	"os"
	"log"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/segmentio/ksuid"
	"github.com/xectich/paymentGateway/controllers"
	"github.com/xectich/paymentGateway/constants"
	"github.com/xectich/paymentGateway/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)


var server = controllers.Server{}
var authorizationInstance = models.Authorization{}
var bankAccountInstance = models.BankAccount{}
var cardInstance = models.Card{}

func TestMain(m *testing.M) {
	var err error
	err = godotenv.Load(os.ExpandEnv("../.env"))
	if err != nil {
		log.Fatalf("Error getting env %v\n", err)
	}
	Database()

	os.Exit(m.Run())
}

func Database() {
	var err error

	TestDbDriver := os.Getenv("TEST_DB_DRIVER")

	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", os.Getenv("TEST_DB_HOST"), os.Getenv("TEST_DB_PORT"), os.Getenv("TEST_DB_USER"), os.Getenv("TEST_DB_NAME"), os.Getenv("TEST_DB_PASSWORD"))
	server.DB, err = gorm.Open(TestDbDriver, DBURL)
	if err != nil {
		fmt.Printf("Cannot connect to %s database\n", TestDbDriver)
		log.Fatal("This is the error:", err)
	} else {
		fmt.Printf("We are connected to the %s database\n", TestDbDriver)
	}
}

func refreshAuthorizationTable() error {
	err := server.DB.DropTableIfExists(&models.Authorization{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.Authorization{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed table")
	return nil
}

func refreshCardTable() error {
	err := server.DB.DropTableIfExists(&models.Card{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.Card{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed table")
	return nil
}

func refreshBankAccountTable() error {
	err := server.DB.DropTableIfExists(&models.BankAccount{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.BankAccount{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed table")
	return nil
}

func addAuthorization() (models.Authorization, error) {

	refreshAuthorizationTable()

	auth := models.Authorization{
		ID:                ksuid.New().String(),
		CardNumber:        "4000000000000119",
		BalanceAuthorised: 5,
		BalanceCaptured:   0,
		CurrencyRequested: "USD",
		CurrencyCard:      "USD",
		Status:            constants.AuthStatus(constants.Authorized).String(),
	}

	err := server.DB.Model(&models.Authorization{}).Create(&auth).Error
	if err != nil {
		log.Fatalf("cannot add to authorization table: %v", err)
	}
	return auth, nil
}


func addBankAccount() (models.BankAccount, error) {

	refreshBankAccountTable()

	ba := models.BankAccount{
		CardID: "4000000000000119",
		Balance: 100,
		Currency: "USD",
	}

	err := server.DB.Model(&models.BankAccount{}).Create(&ba).Error
	if err != nil {
		log.Fatalf("cannot add to bank account table: %v", err)
	}
	return ba, nil
}


func addCard() (models.Card, error) {

	refreshCardTable()

	card := models.Card{
		Number: "4000000000000119",
		CVV: "123",
		Currency: "USD",
		ExpirationMonth: 1,
		ExpirationYear: 23,
	}

	err := server.DB.Model(&models.Card{}).Create(&card).Error
	if err != nil {
		log.Fatalf("cannot add to bank account table: %v", err)
	}
	return card, nil
}