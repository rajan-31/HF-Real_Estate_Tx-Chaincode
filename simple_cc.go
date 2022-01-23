package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// ------------------------------------

type Admin_Super struct {
	Password string `json:"password"`
	UID      string `json:"uid"`  //  Inspector General of Registration
	Name     string `json:"name"` //  Inspector General of Registration
}

// OfficeCode: Tri letter unique code give to each Sub-Registrar's office
type Admin_OfficeCode struct {
	Password  string   `json:"password"`
	UID       string   `json:"uid"`       // Sub-Registrar
	Name      string   `json:"name"`      // Sub-Registrar
	ToApprove []string `json:"toApprove"` // transactions to approve
}

type User struct {
	Password string   `json:"password"`
	UID      string   `json:"uid"`
	Name     string   `json:"name"`
	Status   int      `json:"status"` // 0/1/2 - Not verified/Verified/Suspended
	Owned    []string `json:"owned"`
}

type Request struct {
	Buyer         string    `json:"buyer"`
	Name          string    `json:"name"`
	ProposedPrice int       `json:"proposedPrice"`
	DateTime      time.Time `json:"dateTime"`
}

type Transaction struct {
	Seller              string    `json:"seller"`
	Buyer               string    `json:"buyer"`
	TransactionDateTime time.Time `json:"transactionDateTime"` // when seller/owner accepted the request
	OfficeCode          string    `json:"officeCode"`          // Where estate resides
	ApprovedBy          string    `json:"approvedBy"`          // uid
	ApprovedDateTime    time.Time `json:"approvedDateTime"`
	Price               int       `json:"price"`  // accepted buy seller/owner
	Reason              string    `json:"reason"` // sell, inheritance, gift
}

type Estate struct {
	Owner             string    `json:"owner"`             // uid
	OfficeCode        string    `json:"officeCode"`        // Where estate resides
	Location          string    `json:"location"`          // address
	Area              int       `json:"area"`              // in sq mtr
	Status            int       `json:"status"`            // 0/1/2 - Not verified/Verified/Suspended
	PurchasedOn       time.Time `json:"purchasedOn"`       // current owner since
	SaleAvailability  bool      `json:"saleAvailability"`  // bool
	TransactionsCount int       `json:"transactionsCount"` // total transactions till now
	Requests          []Request `json:"requests"`          // all request from buyers
	BeingSold         bool      `json:"beingSold"`         // true when a request from buyer is accepted
}

// struct for events
/*
type Transaction_Event struct {
	ServeyNo         string `json:"serveyNo"`
	TransactionCount int    `json:"transactionCount"`
	Transaction      Transaction
}
*/

// ------------------------------------

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	// put super user name and password
	key := "admin_super"
	data := Admin_Super{
		Password: "123456",
		Name:     "FName MName LName",
		UID:      "123456789012",
	}

	marshaled_data, _ := json.Marshal(data)

	err := ctx.GetStub().PutState(key, marshaled_data)

	if err != nil {
		return fmt.Errorf("InitLedger >> failed to put to world state. %s", err.Error())
	}

	fmt.Println("=====================================")
	fmt.Println("Chaincode Initated.")
	fmt.Println("=====================================")
	fmt.Print("\n")
	return nil
}

// ------------------------------------

// Helper Functions - Private

func (s *SmartContract) verifyPassword(ctx contractapi.TransactionContextInterface, _username string, _password string) (bool, error) {

	// get data
	dataAsBytes, err0 := ctx.GetStub().GetState(_username)

	if err0 != nil {
		return false, fmt.Errorf("GetPassword >> Failed to read from world state. %s", err0.Error())
	}

	if dataAsBytes == nil {
		return false, fmt.Errorf("GetPassword >> %s does not exist", _username)
	}

	// extract password
	// this is to handle the data from unknown/misc structs
	data := make(map[string]interface{})
	err1 := json.Unmarshal(dataAsBytes, &data)
	if err1 != nil {
		return false, fmt.Errorf("GetPassword >> Can't Unmarshal Data")
	}

	password, ok := data["password"].(string)
	if !ok {
		// password is not a string
		return false, fmt.Errorf("GetPassword >> Password is not a string")
	}

	if _password == password {
		return true, nil
	}

	return false, nil
}

func searchArray(arr []string, val string) int {
	for i, s := range arr {
		if s == val {
			return i
		}
	}
	return -1
}

// ------------------------------------

// For Admin super

func (s *SmartContract) CreateOrModify_Admin(ctx contractapi.TransactionContextInterface, _username string, _password string, officeCode string, newAdminPassword string, uid string, name string) error {

	verified, err0 := s.verifyPassword(ctx, _username, _password)

	if err0 != nil {
		return fmt.Errorf("verifyPassword >> Verify password %s", err0.Error())
	} else if !verified {
		return fmt.Errorf("CreateOrModify_Admin >> Password Missmatched for %s", _username)
	}

	//=====================================

	key := "admin_" + officeCode
	data := Admin_OfficeCode{
		Password:  newAdminPassword,
		UID:       uid,
		Name:      name,
		ToApprove: []string{},
	}

	marshaled_data, _ := json.Marshal(data)
	err1 := ctx.GetStub().PutState(key, marshaled_data)
	if err1 != nil {
		return fmt.Errorf("CreateOrModify_Admin >> Failed to put to world state. %s", err1.Error())
	}
	return nil
}

// For Admin

func (s *SmartContract) Create_User(ctx contractapi.TransactionContextInterface, uid string, name string) (User, error) {

	key := "user" + "_" + uid
	data := User{
		Password: uid,
		Name:     name,
		UID:      uid,
		Status:   0,
		Owned:    []string{},
	}

	marshaled_data, _ := json.Marshal(data)
	err1 := ctx.GetStub().PutState(key, marshaled_data)
	if err1 != nil {
		return User{}, fmt.Errorf("Create_User >> Failed to put to world state. %s", err1.Error())
	}
	return data, nil
}

func (s *SmartContract) Modify_User(ctx contractapi.TransactionContextInterface, uid string, name string, status int) (User, error) {

	key := "user" + "_" + uid
	user := new(User)

	// get user data
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return User{}, fmt.Errorf("Modify_User >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return User{}, fmt.Errorf("Modify_User >> %s does not exist", key)
	}

	err2 := json.Unmarshal(dataAsBytes, &user)
	if err2 != nil {
		return User{}, fmt.Errorf("Modify_User >> Can't Unmarshal Data")
	}

	//=====================================
	temp_status := user.Status
	if status != -1 {
		temp_status = status
	}
	data := User{
		Password: user.Password,
		Name:     name,
		UID:      user.UID,
		Status:   temp_status,
		Owned:    user.Owned,
	}

	marshaled_data, _ := json.Marshal(data)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return User{}, fmt.Errorf("Modify_User >> Failed to put to world state. %s", err3.Error())
	}

	return data, nil
}

func (s *SmartContract) Create_Estate(ctx contractapi.TransactionContextInterface, officeCode string, serveyNo string, owner string, location string, area int, purchasedOn string, transactionsCount int) (Estate, error) {

	key := "estate" + "_" + serveyNo
	temp_dateTime, _ := time.Parse(time.RFC3339, purchasedOn) // purchasedOn => 2021-12-15T20:34:33+05:30
	data := Estate{
		Owner:             owner,
		OfficeCode:        officeCode,
		Location:          location,
		Area:              area,
		Status:            0,
		PurchasedOn:       temp_dateTime,
		SaleAvailability:  false,
		TransactionsCount: transactionsCount,
		Requests:          []Request{},
		BeingSold:         false,
	}

	marshaled_data, _ := json.Marshal(data)
	err1 := ctx.GetStub().PutState(key, marshaled_data)
	if err1 != nil {
		return Estate{}, fmt.Errorf("Create_Estate >> Failed to put to world state. %s", err1.Error())
	}

	//=====================================

	// get data
	key = "user" + "_" + owner
	dataAsBytes, err2 := ctx.GetStub().GetState(key)

	if err2 != nil {
		return data, fmt.Errorf("Create_Estate >> Failed to read from world state. %s", err2.Error())
	}

	if dataAsBytes == nil {
		return data, fmt.Errorf("Create_Estate >> %s does not exist", key)
	}

	user := new(User)
	err3 := json.Unmarshal(dataAsBytes, &user)
	if err3 != nil {
		return data, fmt.Errorf("GetValue >> Can't Unmarshal Data")
	}

	temp_owned := user.Owned
	i := searchArray(temp_owned, serveyNo)
	if i != -1 {
		return data, fmt.Errorf("Create_Estate >> User alredy own estate with serveyNo: %s", serveyNo)
	}

	temp_owned = append(temp_owned, serveyNo)
	user.Owned = temp_owned

	key2 := "user" + "_" + owner
	marshaled_data2, _ := json.Marshal(user)
	err4 := ctx.GetStub().PutState(key2, marshaled_data2)
	if err4 != nil {
		return Estate{}, fmt.Errorf("Create_Estate >> Failed to put to world state. %s", err4.Error())
	}

	return data, nil
}

func (s *SmartContract) Modify_Estate(ctx contractapi.TransactionContextInterface, officeCode string, serveyNo string, location string, area int, purchasedOn string, transactionsCount int) (Estate, error) {

	// get data
	key := "estate" + "_" + serveyNo
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return Estate{}, fmt.Errorf("Modify_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return Estate{}, fmt.Errorf("Modify_Estate >> %s does not exist", key)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes, &estate)
	if err2 != nil {
		return Estate{}, fmt.Errorf("GetValue >> Can't Unmarshal Data")
	}

	//=====================================

	var temp_dateTime time.Time
	if location == "" {
		location = estate.Location
	}
	if area == -1 {
		area = estate.Area
	}
	if purchasedOn == "" {
		temp_dateTime = estate.PurchasedOn
	} else {
		temp_dateTime, _ = time.Parse(time.RFC3339, purchasedOn) // purchasedOn => 2021-12-15T20:34:33+05:30
	}
	if transactionsCount == -1 {
		transactionsCount = estate.TransactionsCount
	}

	data := Estate{
		Owner:             estate.Owner,
		OfficeCode:        officeCode,
		Location:          location,
		Area:              area,
		Status:            estate.Status,
		PurchasedOn:       temp_dateTime,
		SaleAvailability:  estate.SaleAvailability,
		TransactionsCount: transactionsCount,
		Requests:          estate.Requests,
		BeingSold:         estate.BeingSold,
	}

	marshaled_data, _ := json.Marshal(data)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return Estate{}, fmt.Errorf("Modify_Estate >> Failed to put to world state. %s", err3.Error())
	}

	return data, nil
}

func (s *SmartContract) Add_Transaction(ctx contractapi.TransactionContextInterface, serveyNo string, num int, seller string, buyer string, reason string, proposedPrice int, tDateTime string, officeCode string, approvedBy string, aDateTime string) (Transaction, error) {
	key := "transaction" + "_" + serveyNo + "_" + strconv.Itoa(num)

	temp_tDateTime, _ := time.Parse(time.RFC3339, tDateTime)
	temp_aDateTime, _ := time.Parse(time.RFC3339, aDateTime)
	data := Transaction{
		Seller:              seller,
		Buyer:               buyer,
		TransactionDateTime: temp_tDateTime,
		OfficeCode:          officeCode,
		ApprovedBy:          approvedBy,
		ApprovedDateTime:    temp_aDateTime,
		Price:               proposedPrice,
		Reason:              reason,
	}

	marshaled_data0, _ := json.Marshal(data)
	err0 := ctx.GetStub().PutState(key, marshaled_data0)
	if err0 != nil {
		return Transaction{}, fmt.Errorf("Add_Transaction >> failed to put to world state. %s", err0.Error())
	}

	return data, nil
}

func (s *SmartContract) ApproveSell_Estate(ctx contractapi.TransactionContextInterface, _username string, serveyNo string, action string, dateTime string) (Estate, error) {

	// get data of estate
	key1 := "estate" + "_" + serveyNo
	dataAsBytes0, err1 := ctx.GetStub().GetState(key1)

	if err1 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes0 == nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> %s does not exist", key1)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes0, &estate)
	if err2 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Can't Unmarshal Data")
	}

	// get data of transaction
	key2 := "transaction" + "_" + serveyNo + "_" + strconv.Itoa(estate.TransactionsCount+1)
	dataAsBytes1, err2 := ctx.GetStub().GetState(key2)

	if err2 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Failed to read from world state. %s", err2.Error())
	}

	if dataAsBytes1 == nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> %s does not exist", key2)
	}

	transaction := new(Transaction)
	err3 := json.Unmarshal(dataAsBytes1, &transaction)
	if err3 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Can't Unmarshal Data")
	}

	// update approvedBy in transaction
	key3 := _username
	dataAsBytes2, err3 := ctx.GetStub().GetState(key3)

	if err3 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Failed to read from world state. %s", err3.Error())
	}

	if dataAsBytes1 == nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> %s does not exist", key3)
	}

	admin := new(Admin_OfficeCode)
	err4 := json.Unmarshal(dataAsBytes2, &admin)
	if err4 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Can't Unmarshal Data")
	}

	temp_dateTime, _ := time.Parse(time.RFC3339, dateTime)

	transaction.ApprovedBy = admin.UID
	transaction.ApprovedDateTime = temp_dateTime

	// update owner, purchasedOn of estate
	estate.Owner = transaction.Buyer
	estate.PurchasedOn = temp_dateTime
	estate.SaleAvailability = false
	estate.TransactionsCount++
	estate.BeingSold = false

	// update transaction
	marshaled_data0, _ := json.Marshal(estate)
	err5 := ctx.GetStub().PutState(key1, marshaled_data0)
	if err5 != nil {
		return Estate{}, fmt.Errorf("ApproveSell_Estate >> Failed to put to world state. %s", err5.Error())
	}

	// update estate
	marshaled_data1, _ := json.Marshal(transaction)
	err6 := ctx.GetStub().PutState(key2, marshaled_data1)
	if err6 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Failed to put to world state. %s", err6.Error())
	}

	//=====================================

	// remove estate from seller's owned

	key4 := "user" + "_" + transaction.Seller
	user0 := new(User)

	// get user0 data
	dataAsBytes3, err7 := ctx.GetStub().GetState(key4)

	if err7 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Failed to read from world state. %s", err7.Error())
	}

	if dataAsBytes3 == nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> %s does not exist", key4)
	}

	err8 := json.Unmarshal(dataAsBytes3, &user0)
	if err8 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Can't Unmarshal Data")
	}

	temp_owned0 := user0.Owned
	i0 := searchArray(temp_owned0, serveyNo)
	user0.Owned = append(temp_owned0[:i0], temp_owned0[i0+1:]...)

	marshaled_data2, _ := json.Marshal(user0)
	err9 := ctx.GetStub().PutState(key4, marshaled_data2)
	if err9 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Failed to put to world state. %s", err9.Error())
	}

	// add estate to buyer's owned

	key5 := "user" + "_" + transaction.Buyer
	user1 := new(User)

	// get user1 data
	dataAsBytes4, err10 := ctx.GetStub().GetState(key5)

	if err10 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Failed to read from world state. %s", err10.Error())
	}

	if dataAsBytes4 == nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> %s does not exist", key5)
	}

	err11 := json.Unmarshal(dataAsBytes4, &user1)
	if err11 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Can't Unmarshal Data")
	}

	temp_owned1 := user1.Owned
	user1.Owned = append(temp_owned1, serveyNo)

	marshaled_data3, _ := json.Marshal(user1)
	err12 := ctx.GetStub().PutState(key5, marshaled_data3)
	if err12 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> Failed to put to world state. %s", err12.Error())
	}

	//=====================================
	// remove from admin toApprove

	temp_toApprove := admin.ToApprove
	i1 := searchArray(temp_toApprove, key2)
	admin.ToApprove = append(temp_toApprove[:i1], temp_toApprove[i1+1:]...)

	marshaled_data4, _ := json.Marshal(admin)
	err13 := ctx.GetStub().PutState(key3, marshaled_data4)
	if err13 != nil {
		return *estate, fmt.Errorf("ApproveSell_Estate >> failed to put to world state. %s", err13.Error())
	}

	return *estate, nil
}

// User

func (s *SmartContract) ChangeAvail_Estate(ctx contractapi.TransactionContextInterface, serveyNo string, saleAvailability bool) error {
	// get data
	key := "estate" + "_" + serveyNo
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return fmt.Errorf("ChangeAvail_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return fmt.Errorf("ChangeAvail_Estate >> %s does not exist", key)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes, &estate)
	if err2 != nil {
		return fmt.Errorf("GetValue >> Can't Unmarshal Data")
	}

	//=====================================

	estate.SaleAvailability = saleAvailability

	marshaled_data, _ := json.Marshal(estate)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return fmt.Errorf("ChangeAvail_Estate >> Failed to put to world state. %s", err3.Error())
	}

	return nil
}

func (s *SmartContract) RequestToBuy_Estate(ctx contractapi.TransactionContextInterface, _buyer string, _name string, serveyNo string, proposedPrice int, dateTime string) (Request, error) {
	// get data
	key := "estate" + "_" + serveyNo
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return Request{}, fmt.Errorf("RequestToBuy_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return Request{}, fmt.Errorf("RequestToBuy_Estate >> %s does not exist", serveyNo)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes, &estate)
	if err2 != nil {
		return Request{}, fmt.Errorf("RequestToBuy_Estate >> Can't Unmarshal Data")
	}

	//=====================================

	temp_requests := estate.Requests
	temp_dateTime, _ := time.Parse(time.RFC3339, dateTime)

	flag := false
	index := 0
	for i, r := range temp_requests {
		if r.Buyer == _buyer {
			temp_requests[i].ProposedPrice = proposedPrice
			temp_requests[i].DateTime = temp_dateTime
			flag = true
			index = i
			break
		}
	}

	if !flag {
		temp_requests = append(temp_requests, Request{
			Buyer:         _buyer,
			Name:          _name,
			ProposedPrice: proposedPrice,
			DateTime:      temp_dateTime,
		})
		index = len(temp_requests) - 1
	}
	estate.Requests = temp_requests

	//=====================================

	marshaled_data, _ := json.Marshal(estate)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return Request{}, fmt.Errorf("RequestToBuy_Estate >> Failed to put to world state. %s", err3.Error())
	}

	return temp_requests[index], nil
}

func (s *SmartContract) AcceptRequest_Estate(ctx contractapi.TransactionContextInterface, _username string, _password string, serveyNo string, buyer string, dateTime string, reason string) (Transaction, error) {
	verified, err0 := s.verifyPassword(ctx, _username, _password)

	if err0 != nil {
		return Transaction{}, fmt.Errorf("verifyPassword >> Verify password %s", err0.Error())
	} else if !verified {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Password Missmatched for %s", _username)
	}

	//=====================================

	// get data
	key1 := "estate" + "_" + serveyNo
	dataAsBytes, err1 := ctx.GetStub().GetState(key1)

	if err1 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> %s does not exist", serveyNo)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes, &estate)
	if err2 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Can't Unmarshal Data")
	}

	//=====================================

	// check if being sold already
	if estate.BeingSold {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Estate is already being sold")
	}

	// stop from accepting other requests
	estate.BeingSold = true

	// it is updated in later step

	//=====================================

	flag := false
	index := 0
	for i, r := range estate.Requests {
		if r.Buyer == buyer {
			flag = true
			index = i
		}
	}

	if !flag {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> No request found for given buyer: %s", buyer)
	}

	temp_dateTime, _ := time.Parse(time.RFC3339, dateTime)
	temp_transaction := Transaction{
		Seller:              strings.TrimPrefix(_username, "user_"),
		Buyer:               estate.Requests[index].Buyer,
		TransactionDateTime: temp_dateTime,
		OfficeCode:          estate.OfficeCode,
		ApprovedBy:          "",
		ApprovedDateTime:    time.Time{},
		Price:               estate.Requests[index].ProposedPrice,
		Reason:              reason,
	}

	key2 := "transaction" + "_" + serveyNo + "_" + strconv.Itoa(estate.TransactionsCount+1)
	marshaled_data1, _ := json.Marshal(temp_transaction)
	err4 := ctx.GetStub().PutState(key2, marshaled_data1)
	if err4 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> failed to put to world state. %s", err4.Error())
	}

	// update being sold flag and delete all requests
	estate.Requests = []Request{}

	marshaled_data0, _ := json.Marshal(estate)
	err3 := ctx.GetStub().PutState(key1, marshaled_data0)
	if err3 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> failed to put to world state. %s", err3.Error())
	}

	//=====================================
	// add request in toApprove of admin

	key3 := "admin" + "_" + estate.OfficeCode
	dataAsBytes1, err4 := ctx.GetStub().GetState(key3)

	if err4 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Failed to read from world state. %s", err4.Error())
	}

	if dataAsBytes1 == nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> %s does not exist", key3)
	}

	admin := new(Admin_OfficeCode)
	err5 := json.Unmarshal(dataAsBytes1, &admin)
	if err5 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> Can't Unmarshal Data")
	}

	admin.ToApprove = append(admin.ToApprove, key2)

	marshaled_data2, _ := json.Marshal(admin)
	err6 := ctx.GetStub().PutState(key3, marshaled_data2)
	if err6 != nil {
		return Transaction{}, fmt.Errorf("AcceptRequest_Estate >> failed to put to world state. %s", err6.Error())
	}

	//=====================================
	// set event

	/*
		transaction_event := Transaction_Event{
			ServeyNo:         serveyNo,
			TransactionCount: estate.TransactionsCount + 1,
			Transaction:      temp_transaction,
		}

		marshaled_data3, _ := json.Marshal(transaction_event)

		ctx.GetStub().SetEvent("newEstateSell", marshaled_data3)
	*/

	return temp_transaction, nil
}

// For System

func (s *SmartContract) Verify_User(ctx contractapi.TransactionContextInterface, _username string, _password string, status int, newPassword string) error {
	verified, err0 := s.verifyPassword(ctx, _username, _password)

	if err0 != nil {
		return fmt.Errorf("verifyPassword >> Verify password %s", err0.Error())
	} else if !verified {
		return fmt.Errorf("Verify_User >> Password Missmatched for %s", _username)
	}

	//=====================================

	key := _username
	user := new(User)

	// get user data
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return fmt.Errorf("Verify_User >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return fmt.Errorf("Verify_User >> %s does not exist", key)
	}

	err2 := json.Unmarshal(dataAsBytes, &user)
	if err2 != nil {
		return fmt.Errorf("Verify_User >> Can't Unmarshal Data")
	}

	//=====================================

	user.Status = status
	user.Password = newPassword

	marshaled_data, _ := json.Marshal(user)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return fmt.Errorf("Verify_User >> Failed to put to world state. %s", err3.Error())
	}

	return nil
}

func (s *SmartContract) Verify_Estate(ctx contractapi.TransactionContextInterface, _username string, _password string, serveyNo string, status int) error {
	verified, err0 := s.verifyPassword(ctx, _username, _password)

	if err0 != nil {
		return fmt.Errorf("verifyPassword >> Verify password %s", err0.Error())
	} else if !verified {
		return fmt.Errorf("Verify_Estate >> Password Missmatched for %s", _username)
	}

	//=====================================

	// get data
	key := "estate" + "_" + serveyNo
	dataAsBytes, err1 := ctx.GetStub().GetState(key)

	if err1 != nil {
		return fmt.Errorf("Verify_Estate >> Failed to read from world state. %s", err1.Error())
	}

	if dataAsBytes == nil {
		return fmt.Errorf("Verify_Estate >> %s does not exist", key)
	}

	estate := new(Estate)
	err2 := json.Unmarshal(dataAsBytes, &estate)
	if err2 != nil {
		return fmt.Errorf("GetValue >> Can't Unmarshal Data")
	}

	//=====================================

	estate.Status = status

	marshaled_data, _ := json.Marshal(estate)
	err3 := ctx.GetStub().PutState(key, marshaled_data)
	if err3 != nil {
		return fmt.Errorf("Verify_Estate >> Failed to put to world state. %s", err3.Error())
	}

	return nil
}

// ------------------------------------

// Testing

func (s *SmartContract) GetValue(ctx contractapi.TransactionContextInterface, _key string) (map[string]interface{}, error) {

	// this is to handle the data from unknown/misc structs
	data := make(map[string]interface{})

	dataAsBytes, err0 := ctx.GetStub().GetState(_key)

	if err0 != nil {
		return data, fmt.Errorf("GetValue >> Failed to read from world state. %s", err0.Error())
	}

	if dataAsBytes == nil {
		return data, fmt.Errorf("GetValue >> %s does not exist", _key)
	}

	err1 := json.Unmarshal(dataAsBytes, &data)
	if err1 != nil {
		return data, fmt.Errorf("GetValue >> Can't Unmarshal Data")
	}

	fmt.Println(data)
	fmt.Print("\n")

	return data, nil
}

func (s *SmartContract) DeleteValue(ctx contractapi.TransactionContextInterface, _key string) error {

	err0 := ctx.GetStub().DelState(_key)
	if err0 != nil {
		return fmt.Errorf("DeleteValue >> Can't Delete Value for %s. %s", _key, err0.Error())
	}

	fmt.Println("Deleted value for key:", _key)
	fmt.Print("\n")

	return nil
}

func (s *SmartContract) GetAll(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]string, error) {

	resultsIterator, err0 := ctx.GetStub().GetStateByRange(startKey, endKey)

	arrMap := []string{}

	if err0 != nil {
		return arrMap, err0
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err1 := resultsIterator.Next()

		if err1 != nil {
			return arrMap, err1
		}

		// data := make(map[string]interface{})
		// err2 := json.Unmarshal(queryResponse.Value, &data)
		// if err2 != nil {
		// 	return arrMap, fmt.Errorf("GetAll >> Can't Unmarshal Data")
		// }

		data := string(queryResponse.Value)

		fmt.Println("=====================================")
		fmt.Println("Key: "+queryResponse.Key+", Value: ", data)
		fmt.Println("=====================================")

		arrMap = append(arrMap, "Key: "+string(queryResponse.Key)+", Value: "+data)
	}

	fmt.Print("\n")

	return arrMap, nil
}

// ------------------------------------

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create Real Estate chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting Real Estate chaincode: %s", err.Error())
	}

}

//=====================================

// ------------------------------------
