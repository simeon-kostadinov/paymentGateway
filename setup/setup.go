package setup

import (
	"log"
	"github.com/jinzhu/gorm"
	"github.com/xectich/paymentGateway/models"
)

var cards = []models.Card {
	models.Card{
		Number: "4000000000000119",
		CVV: "123",
		Currency: "USD",
		ExpirationMonth: 1,
		ExpirationYear: 23,
	},
	models.Card{
		Number: "4000000000000259",
		CVV: "453",
		Currency: "BGN",
		ExpirationMonth: 4,
		ExpirationYear: 22,
	},
	models.Card{
		Number: "4000000000003238",
		CVV: "765",
		Currency: "GBP",
		ExpirationMonth: 10,
		ExpirationYear: 21,
	},
	models.Card{
		Number: "4000000000004422",
		CVV: "221",
		Currency: "EUR",
		ExpirationMonth: 10,
		ExpirationYear: 25,
	},
}

var bankAccounts = []models.BankAccount {
	models.BankAccount {
		CardID: "4000000000000119",
		Balance: 0,
		Currency: "USD",
	},
	models.BankAccount {
		CardID: "4000000000000259",
		Balance: 10,
		Currency: "CAD",
	},
	models.BankAccount {
		CardID: "4000000000003238",
		Balance: 1000,
		Currency: "GBP",
	},
	models.BankAccount {
		CardID: "4000000000004422",
		Balance: 1000,
		Currency: "EUR",
	},
}

func Load(db *gorm.DB) {

	err := db.Debug().DropTableIfExists(&models.BankAccount{}, &models.Card{},&models.Authorization{}).Error
	if err != nil {
		log.Fatalf("cannot drop table: %v", err)
	}
	err = db.Debug().AutoMigrate(&models.BankAccount{}, &models.Card{},&models.Authorization{}).Error
	if err != nil {
		log.Fatalf("cannot migrate table: %v", err)
	}

	for i, _ := range bankAccounts {
		err = db.Debug().Model(&models.BankAccount{}).Create(&bankAccounts[i]).Error
		if err != nil {
			log.Fatalf("cannot setup bank accounts table: %v", err)
		}

		err = db.Debug().Model(&models.Card{}).Create(&cards[i]).Error
		if err != nil {
			log.Fatalf("cannot setup cards table: %v", err)
		}
	}
}