package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	// "log"
	// "fmt"

	"github.com/gorilla/mux"
	"github.com/xectich/paymentGateway/auth"
	"github.com/xectich/paymentGateway/constants"
	"github.com/xectich/paymentGateway/models"
	"github.com/xectich/paymentGateway/responses"
)

type Merchant struct {
	ID uint32 `json:"mid"`
}

//Login creates a token based on the merchantID, can be implemented with a proper login system
func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	merchant := Merchant{}
	err = json.Unmarshal(body, &merchant)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, err := auth.CreateToken(merchant.ID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(constants.UnableToCreateJWTToken))
		return
	}
	responses.JSON(w, http.StatusOK, token)
}

//RequestAuthorization handles the request/response for new authorization requests
func (server *Server) RequestAuthorization(w http.ResponseWriter, r *http.Request) {
	//get the request body and umarshall it into request struct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	//get the merchant ID and token and validate
	vars := mux.Vars(r)
	mid, err := strconv.ParseUint(vars["mid"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != uint32(mid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	authRequest := models.AuthorizationRequest{}
	err = json.Unmarshal(body, &authRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//call the function to request authorization from the auth interface
	authI := models.NewAuthI()

	auth, err := authI.RequestAuthorization(authRequest, server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	//construct the response
	authResponse := models.AuthorizationResponse{
		ID:              auth.ID,
		Currency:        auth.CurrencyCard,
		AmountAvailable: auth.BalanceAuthorised,
	}

	responses.JSON(w, http.StatusCreated, authResponse)
}

//Capture handles the request/response for capturing funds on a customers bank
func (server *Server) Capture(w http.ResponseWriter, r *http.Request) {
	//get the request body and umarshall it into request struct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	//get the merchant ID and token and validate
	vars := mux.Vars(r)
	mid, err := strconv.ParseUint(vars["mid"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != uint32(mid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	actionRequest := models.ActionRequest{}
	err = json.Unmarshal(body, &actionRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//call the function to request authorization from the auth interface
	authI := models.NewAuthI()

	auth, err := authI.Capture(actionRequest.ID, actionRequest.Amount, actionRequest.Final, server.DB)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	//construct the response
	authResponse := models.AuthorizationResponse{
		Currency:        auth.CurrencyCard,
		AmountAvailable: auth.BalanceAuthorised - auth.BalanceCaptured,
	}

	responses.JSON(w, http.StatusCreated, authResponse)
}

//Refynd handles the request/response for refunding funds 
func (server *Server) Refund(w http.ResponseWriter, r *http.Request) {
	//get the request body and umarshall it into request struct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	//get the merchant ID and token and validate
	vars := mux.Vars(r)
	mid, err := strconv.ParseUint(vars["mid"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != uint32(mid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	actionRequest := models.ActionRequest{}
	err = json.Unmarshal(body, &actionRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//call the function to request authorization from the auth interface
	authI := models.NewAuthI()

	auth, err := authI.Refund(actionRequest.ID, actionRequest.Amount, actionRequest.Final, server.DB)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	//construct the response
	authResponse := models.AuthorizationResponse{
		Currency:        auth.CurrencyCard,
		AmountAvailable: auth.BalanceAuthorised - auth.BalanceCaptured,
	}

	responses.JSON(w, http.StatusCreated, authResponse)
}

//Void handles the request/response for voiding a transaction
func (server *Server) Void(w http.ResponseWriter, r *http.Request) {
	//get the request body and umarshall it into request struct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	//get the merchant ID and token and validate
	vars := mux.Vars(r)
	mid, err := strconv.ParseUint(vars["mid"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != uint32(mid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	actionRequest := models.ActionRequest{}
	err = json.Unmarshal(body, &actionRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//call the function to request authorization from the auth interface
	authI := models.NewAuthI()

	auth, err := authI.Void(actionRequest.ID, server.DB)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	//construct the response
	authResponse := models.AuthorizationResponse{
		Currency:        auth.CurrencyCard,
		AmountAvailable: 0,
	}

	responses.JSON(w, http.StatusCreated, authResponse)
}
