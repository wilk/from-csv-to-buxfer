#! ./bin/python3

'''
Collector has the purpose to insert in a MongoDB collection all the cleaned data.
The model is set as follows:

Model {
  id: ObjectID,
  description: String,
  amount: Float,
  tags: String[],
  account: String,
  date: Date
}
'''

import csv
import sys
import os
from os import listdir
from os.path import isfile, join
from pymongo import MongoClient
from datetime import datetime
import json

CLEANED_FOLDER = os.path.abspath(os.getenv('CLEANED_FOLDER', 'cleaned'))
EXPENSE_ACCOUNT = os.getenv('EXPENSE_ACCOUNT', 'expenses')
INCOME_ACCOUNT = os.getenv('INCOME_ACCOUNT', 'income')
TAGS_FILE = os.path.abspath(os.getenv('TAGS_FILE', 'config/sample-tags.json'))
DB_NAME = os.getenv('DB_NAME', 'collector')
DB_COLLECTION_NAME = os.getenv('DB_COLLECTION_NAME', 'collected')

if not os.path.isdir(CLEANED_FOLDER):
  print('Launch the cleaner before the collector')
  sys.exit()

# mongodb client
client = MongoClient(host=os.getenv('DB_HOST'), port=int(os.getenv('DB_PORT')))
db = client[DB_NAME]

# empty the collected collection
db[DB_COLLECTION_NAME].delete_many({})

# load tags mapping
with open(TAGS_FILE) as json_file:
  tags = json.load(json_file)

# read the csv files list
csv_files = [f for f in listdir(CLEANED_FOLDER) if isfile(join(CLEANED_FOLDER, f))]

totalCounter = 0
totalAmount = 0
for filename in csv_files:
  filepath = join(CLEANED_FOLDER, filename)
  account = EXPENSE_ACCOUNT if 'expenses' in filename else INCOME_ACCOUNT
  with open(filepath) as file:
    print("reading", filename)
    # read cleaned csv file
    reader = csv.reader(file, delimiter=',', quotechar='"')
    # skip headers
    next(reader, None)
    transactions_counter = 0
    for row in reader:
      amount = float(row[3].replace('.', '').replace(',', '.').replace('-', ''))
      totalAmount += amount
      transaction = dict({
        "date": row[0],
        "account": account,
        "description": row[2],
        # amount could be -20.000,54 and it needs to be converted like 20000.54
        # negative numbers are for income transactions: convert them into positive numbers
        "amount": amount,
        "tags": tags[filename][row[1]]
      })

      db[DB_COLLECTION_NAME].insert_one(transaction)
      transactions_counter += 1
    # log how transactions have been added
    print(transactions_counter, "transactions added from", filename, "as", account)
    totalCounter += transactions_counter

# test the collection
collectionCounter = 0
collectionAmount = 0
cursor = db[DB_COLLECTION_NAME].find()
for document in cursor:
  print(document)
  collectionAmount += document['amount']
  collectionCounter += 1

assert totalCounter == collectionCounter
assert totalAmount == collectionAmount

print('inserted docs:', totalCounter, 'found docs:', collectionCounter)
print('inserted amount:', totalAmount, 'found amount:', collectionAmount)
