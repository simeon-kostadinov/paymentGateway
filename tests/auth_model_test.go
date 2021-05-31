package tests

import (
	"log"
	"testing"

	"github.com/xectich/paymentGateway/models"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFindAuthorizationByID(t *testing.T) {
	err := refreshAuthorizationTable()
	if err != nil {
		log.Fatal(err)
	}

	auth, err := addAuthorization()
	if err != nil {
		log.Fatal(err)
	}

	Convey("When I call FindAuthorizationByID..", t, func() {
		a, err := authorizationInstance.FindAuthorizationByID(auth.ID, server.DB)
		Convey("Auth.ID should match the one from the DB", func() {
			So(auth.ID, ShouldEqual, a.ID)
			So(err, ShouldBeNil)
		})
	})

}

func TestRequestAuthorization(t *testing.T) {
	err := refreshAuthorizationTable()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addCard()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	authRequest := models.AuthorizationRequest{
		CardNumber:      "400000000000011",
		Currency:        "USD",
		CVV:             "123",
		Amount:          10,
		ExpirationMonth: 1,
		ExpirationYear:  23,
	}

	Convey("When I call RequestAuthorization..", t, func() {
		Convey("And Card Number is invalid", func() {
			_, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And CVV is invalid", func() {
			authRequest.CardNumber = "4000000000000119"
			authRequest.CVV = "124"
			_, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And Amount is invalid", func() {
			authRequest.CVV = "123"
			authRequest.Amount = 101
			_, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And Amount is Expiration Month is invalid", func() {
			authRequest.Amount = 10
			authRequest.ExpirationMonth = 2
			_, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And request is authorized", func() {
			authRequest.ExpirationMonth = 1
			_, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)
			So(err, ShouldBeNil)
		})
	})

}

func TestCapture(t *testing.T) {

	err := refreshAuthorizationTable()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addCard()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	authRequest := models.AuthorizationRequest{
		CardNumber:      "4000000000000119",
		Currency:        "USD",
		CVV:             "123",
		Amount:          10,
		ExpirationMonth: 1,
		ExpirationYear:  23,
	}

	auth, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)

	Convey("When I call Capture...", t, func() {
		Convey("And amount is 0", func() {
			_, err = authorizationInstance.Capture(auth.ID, 0, false, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And amount is more than capture amount", func() {
			_, err = authorizationInstance.Capture(auth.ID, 1000, false, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And amount is captured", func() {
			auth, err = authorizationInstance.Capture(auth.ID, 5, false, server.DB)
			So(err, ShouldBeNil)
			So(auth.BalanceCaptured, ShouldEqual, 5)
		})

	})

}

func TestRefund(t *testing.T) {

	err := refreshAuthorizationTable()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addCard()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	authRequest := models.AuthorizationRequest{
		CardNumber:      "4000000000000119",
		Currency:        "USD",
		CVV:             "123",
		Amount:          10,
		ExpirationMonth: 1,
		ExpirationYear:  23,
	}

	auth, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)

	Convey("When I call Refund...", t, func() {
		Convey("And amount is 0", func() {
			_, err = authorizationInstance.Refund(auth.ID, 0, false, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And amount is more than capture amount", func() {
			_, err = authorizationInstance.Capture(auth.ID, 1000, false, server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And amount is refunded", func() {
			auth, err = authorizationInstance.Capture(auth.ID, 5, false, server.DB)
			So(err, ShouldBeNil)
			So(auth.BalanceCaptured, ShouldEqual, 5)
			auth, err = authorizationInstance.Refund(auth.ID, 5, false, server.DB)
			So(err, ShouldBeNil)
			So(auth.BalanceRefunded, ShouldEqual, 5)			
		})

	})
}

func TestVoid(t *testing.T) {

	err := refreshAuthorizationTable()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addCard()
	if err != nil {
		log.Fatal(err)
	}

	_, err = addBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	authRequest := models.AuthorizationRequest{
		CardNumber:      "4000000000000119",
		Currency:        "USD",
		CVV:             "123",
		Amount:          10,
		ExpirationMonth: 1,
		ExpirationYear:  23,
	}

	auth, err := authorizationInstance.RequestAuthorization(authRequest, server.DB)

	Convey("When I call Void...", t, func() {
		Convey("Authorization is not found", func() {
			_, err = authorizationInstance.Void("test", server.DB)
			So(err, ShouldNotBeNil)
		})
		Convey("And transaction is voided", func() {
			_, err = authorizationInstance.Void(auth.ID, server.DB)
			So(err, ShouldBeNil)
		})

	})
}


