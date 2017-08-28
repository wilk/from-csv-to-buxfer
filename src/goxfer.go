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
)

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

	for _, transaction := range results {
		fmt.Println("Pushing transaction:", transaction.Description, ", account:", transaction.Account)

		// @todo: open 20 connections per time
	}
}