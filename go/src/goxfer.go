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
	"gopkg.in/mgo.v2/bson"
)

type Transaction struct {
	Id bson.ObjectId `json:"id" bson:"_id,omitempty"`
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

type AccountsListResponse struct {
	Response struct {
		Accounts []struct {
			Id int `json:"id"`
			Name string `json:"name"`
		} `json:"accounts"`
		Status string `json:"status"`
		Token string `json:"token"`
	} `json:"response"`
}

type AddResponseBody struct {
	Response struct {
	  Status string `json:"status"`
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
	EXPENSE_ACCOUNT_BUXFER = os.Getenv("EXPENSE_ACCOUNT_BUXFER")
	EXPENSE_ACCOUNT_ID int
	INCOME_ACCOUNT_ID int
)

func addTransaction(transaction Transaction, token string) error {
	request := gorequest.New()

	accountId := INCOME_ACCOUNT_ID
	accountType := "income"
	if transaction.Account == EXPENSE_ACCOUNT_BUXFER {
		accountId = EXPENSE_ACCOUNT_ID
		accountType = "expense"
	}

	dates := strings.Split(transaction.Date, "/")
	date := dates[2] + "-" + dates[1] + "-" + dates[0]

	payload := map[string]interface{}{
		"description": transaction.Description,
		"amount": transaction.Amount,
		"accountId": accountId,
		"tags": strings.Join(transaction.Tags[:], ","),
		"date": date,
		"token": token,
		"type": accountType,
	}

	var body AddResponseBody
	res, _, errs := request.Post(BUXFER_API_URL + "/add_transaction").
		Send(payload).
		EndStruct(&body)

	if len(errs) > 0 {
		return errs[0]
	}

	if res.StatusCode > 399 {
		return errors.New(res.Status)
	}

	if body.Response.Status != "OK" {
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

	token := result.Response.Token

	fmt.Println("Buxfer session created!")
	fmt.Println("Fetching Buxfer's accounts list...")

	var listResult AccountsListResponse
	res, _, errs = request.Get(BUXFER_API_URL + "/accounts").
		Query("token=" + token).
		EndStruct(&listResult)

	if len(errs) > 0 {
		panic(errs)
	}

	if res.StatusCode > 399 {
		panic(res.Status)
	}

	if listResult.Response.Status != "OK" || len(listResult.Response.Accounts) == 0 {
		panic(errors.New("An error occured when fetching the Buxfer accounts list"))
	}

	for _, account := range listResult.Response.Accounts {
		if account.Name == EXPENSE_ACCOUNT_BUXFER {
			EXPENSE_ACCOUNT_ID = account.Id
		} else {
			INCOME_ACCOUNT_ID = account.Id
		}
	}

	fmt.Println("Accounts list fetched!")
	fmt.Println("Fetching transactions from DB...")

	results := []Transaction{}
	err = collected.Find(nil).All(&results)
	if err != nil {
		panic(err)
	}

	bulks := [][]Transaction{}
	counter := 0
	resultsLen := len(results)
	iterations := resultsLen / BULK_LEN
	for i := 0; i < iterations; i++ {
		bulks = append(bulks, results[counter:counter + BULK_LEN])
		counter += BULK_LEN
	}

	if counter < resultsLen {
		bulks = append(bulks, results[counter:])
	}

	fmt.Println("Transactions fetched and divided into small bulk of #", strconv.Itoa(BULK_LEN), "transactions")
	fmt.Println("Pushing transactions on Buxfer...")

	transactionAddedCounter := 0
	transactionNotAddedCounter := 0
	var transactionIdNotAdded []bson.ObjectId
	wg := &sync.WaitGroup{}
	for index, bulk := range bulks {
		fmt.Println("Pushing bulk #", index)

		wg.Add(len(bulk))
		for _, transaction := range bulk {
			// going parallel
			go func(transaction Transaction) {
				fmt.Println("Pushing transaction:", transaction)

				if err := addTransaction(transaction, token); err != nil {
					transactionNotAddedCounter++
					transactionIdNotAdded = append(transactionIdNotAdded, transaction.Id)
					fmt.Println(err)
				} else {
					transactionAddedCounter++
				}

				wg.Done()
			}(transaction)
		}

		// wait the end of each request of the current bulk
		wg.Wait()
	}

	fmt.Println("Transactions succeded #", strconv.Itoa(transactionAddedCounter))
	fmt.Println("Transactions failed #", strconv.Itoa(transactionNotAddedCounter), strconv.Itoa(len(transactionIdNotAdded)))
	if len(transactionIdNotAdded) > 0 {
		fmt.Println("Transactions failed listed below:")
		for _, id := range transactionIdNotAdded {
			fmt.Println(id)
		}
	}
}