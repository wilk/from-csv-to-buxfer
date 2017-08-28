package main

import (
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"time"
	"os"
	"net/http"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Transaction struct {
	Date time.Time
	Account string
	Description string
	Amount float64
	Tags []string
}

type LoginResponse struct {
	Response struct {
		Status string `json:"status"`
		Token string `json:"token"`
	} `json:"response"`
}

const (
	DB_HOST = os.Getenv("DB_HOST")
	DB_NAME = os.Getenv("DB_NAME")
	DB_COLLECTION_NAME = os.Getenv("DB_COLLECTION_NAME")
	BUXFER_API_URL = os.Getenv("BUXFER_API_ENDPOINT")
	BUXFER_USERNAME = os.Getenv("BUXFER_USERNAME")
	BUXFER_PASSWORD = os.Getenv("BUXFER_PASSWORD")
	BULK_LEN = strconv.Atoi(os.Getenv("BULK_LENGHT"))
	EXPENSE_ACCOUNT = os.Getenv("EXPENSE_ACCOUNT")
	INCOME_ACCOUNT = os.Getenv("INCOME_ACCOUNT")
	EXPENSE_ACCOUNT_BUXFER = os.Getenv("EXPENSE_ACCOUNT_BUXFER")
	INCOME_ACCOUNT_BUXFER = os.Getenv("INCOME_ACCOUNT_BUXFER")
	ACCOUNTS_MAP = map[string]string{EXPENSE_ACCOUNT: EXPENSE_ACCOUNT_BUXFER, INCOME_ACCOUNT: INCOME_ACCOUNT_BUXFER}
)

// @todo: define a working group as a new param
func addTransaction(transaction Transaction) {
	req, err := http.NewRequest("POST", BUXFER_API_URL + "/add_transaction", nil)
	if err != nil {
		panic(err)
	}

	// @todo: make a test to insert a series of tags inside buxfer
	// @todo: check date format
	text := transaction.Description + " " + strconv.FormatFloat(transaction.Amount, "E", -1, 64) + " acct:" + ACCOUNTS_MAP[transaction.Account] + " tags:" + strings.Join(transaction.Tags[:], ",") + " date:" + transaction.Date

	qs := req.URL.Query()
	qs.Add("format", "sms")
	qs.Add("text", text)

	req.URL.RawQuery = qs.Encode()

	// @todo: pass or globalize client
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(res.Body)
	// @todo: create an ad hoc structure for buxfer response body
	var result LoginResponse
	err = decoder.Decode(&result)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}

func main() {
	session, err := mgo.Dial(DB_HOST)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	collected := session.DB(DB_NAME).C(DB_COLLECTION_NAME)

	fmt.Println("DB connected")
	fmt.Println("Creating session for Buxfer...")

	client := &http.Client{}
	req, err := http.NewRequest("GET", BUXFER_API_URL + "/login", nil)
	if err != nil {
		panic(err)
	}

	qs := req.URL.Query()
	qs.Add("userid", BUXFER_USERNAME)
	qs.Add("password", BUXFER_PASSWORD)

	req.URL.RawQuery = qs.Encode()

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(res.Body)
	var result LoginResponse
	err = decoder.Decode(&result)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if result.Response.Status != "OK" || len(result.Response.Token) == 0 {
		panic(errors.New("An error occured during the Buxfer's login"))
	}

	// @todo: use token for future requests
	token := result.Response.Token

	results := []Transaction{}
	err = collected.Find(nil).All(&results)
	if err != nil {
		log.Fatal(err)
	}

	bulks := [][]Transaction{}
	counter := 0
	iterations := len(results) / BULK_LEN
	for i := 0; i < iterations; i++ {
		append(bulks, results[counter:BULK_LEN]...)
		counter += BULK_LEN
	}

	if counter < len(iterations) {
		append(bulks, results[counter:len(iterations) - counter]...)
	}

	for index, bulk := range bulks {
		fmt.Println("Pushing bulk #", index)

		for _, transaction := range bulk {
			fmt.Println("Pushing transaction:", transaction.Description, ", account:", transaction.Account)

			// @todo: call here addTransaction
		}

		// @todo: wait for working group to be done
	}
}