package main

import (
	"fmt"
	"os"
	"errors"
	"strconv"
	"strings"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/mgo.v2"
	"sync"
	"log"
)

type Transaction struct {
	Date string
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

type AddResponseBody struct {
	Response struct {
	  Status string `json:"status"`
		TransactionAdded bool `json:"transactionAdded"`
		ParseStatus string `json:"parseStatus"`
  } `json:"response"`
}

var (
	DB_HOST = os.Getenv("DB_HOST")
	DB_NAME = os.Getenv("DB_NAME")
	DB_COLLECTION_NAME = os.Getenv("DB_COLLECTION_NAME")
	BUXFER_API_URL = os.Getenv("BUXFER_API_ENDPOINT")
	BUXFER_USERNAME = os.Getenv("BUXFER_USERNAME")
	BUXFER_PASSWORD = os.Getenv("BUXFER_PASSWORD")
	BULK_LEN, _ = strconv.Atoi(os.Getenv("BULK_LENGHT"))
	EXPENSE_ACCOUNT = os.Getenv("EXPENSE_ACCOUNT")
	INCOME_ACCOUNT = os.Getenv("INCOME_ACCOUNT")
	EXPENSE_ACCOUNT_BUXFER = os.Getenv("EXPENSE_ACCOUNT_BUXFER")
	INCOME_ACCOUNT_BUXFER = os.Getenv("INCOME_ACCOUNT_BUXFER")
)

func addTransaction(transaction Transaction, token string) error {
	request := gorequest.New()

	// text must follow this pattern:
	// <description> [+]<amount> acct:<account> tags:<tag1,tag2,...> date:<date>
	amount := strconv.FormatFloat(transaction.Amount, 'f', -1, 64)
	if transaction.Account == INCOME_ACCOUNT_BUXFER {
		amount = "+" + amount
	}
	text := transaction.Description + " " + amount + " acct:" + transaction.Account + " tags:" + strings.Join(transaction.Tags[:], ",") + " date:" + transaction.Date

	var body AddResponseBody
	res, _, errs := request.Post(BUXFER_API_URL + "/transaction/add").
		Query("token=" + token).
		Query("format=sms").
		Query("text=" + text).
		EndStruct(&body)

	if len(errs) > 0 {
		return errs[0]
	}

	if res.StatusCode > 399 {
		return errors.New(res.Status)
	}

	if body.Response.Status != "OK" || !body.Response.TransactionAdded || body.Response.ParseStatus != "success" {
		return errors.New("An error occurred during the transaction upload")
	}

	return nil
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

	var result LoginResponse
	request := gorequest.New()
	res, _, errs := request.Get(BUXFER_API_URL + "/login").
		Query("userid=" + BUXFER_USERNAME).
		Query("password=" + BUXFER_PASSWORD).
		EndStruct(&result)

	if len(errs) > 0 {
		panic(errs)
	}

	if res.StatusCode > 399 {
		panic(res.Status)
	}

	if result.Response.Status != "OK" || len(result.Response.Token) == 0 {
		panic(errors.New("An error occured during the Buxfer's login"))
	}

	fmt.Println("Buxfer session created!")
	fmt.Println("Fetching transactions from DB...")

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

	fmt.Println("Transactions fetched and divided into small bulk of #", strconv.Itoa(BULK_LEN), "transactions")
	fmt.Println("Pushing transactions on Buxfer...")

	transactionAddedCounter := 0
	transactionNotAddedCounter := 0
	wg := &sync.WaitGroup{}
	for index, bulk := range bulks {
		fmt.Println("Pushing bulk #", index)

		wg.Add(len(bulk))
		for _, transaction := range bulk {
			// going parallel
			go func() {
				fmt.Println("Pushing transaction:", transaction.Description, ", account:", transaction.Account)

				if err := addTransaction(transaction, token); err != nil {
					transactionNotAddedCounter++
					fmt.Println(err)
				} else {
					transactionAddedCounter++
				}

				wg.Done()
			}()
		}

		// wait the end of each request of the current bulk
		wg.Wait()
	}

	fmt.Println("Transactions succeded #", strconv.Itoa(transactionAddedCounter))
	fmt.Println("Transactions failed #", strconv.Itoa(transactionNotAddedCounter))
}