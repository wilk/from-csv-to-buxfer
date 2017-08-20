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
import logging

CLEANED_FOLDER = os.path.abspath(os.getenv('CLEANED_FOLDER', 'cleaned'))
EXPENSE_ACCOUNT = os.getenv('EXPENSE_ACCOUNT', 'expenses')
INCOME_ACCOUNT = os.getenv('INCOME_ACCOUNT', 'income')
TAGS_FILE = os.path.abspath(os.getenv('TAGS_FILE', 'config/sample-tags.json'))

if not os.path.isdir(CLEANED_FOLDER):
  logging.error('Launch the cleaner before the collector')
  sys.exit()

# mongodb client
client = MongoClient()
db = client.collected

# load tags mapping
with open(TAGS_FILE) as json_file:
  tags = json.load(json_file)

# read the csv files list
csv_files = [f for f in listdir(CLEANED_FOLDER) if isfile(join(CLEANED_FOLDER, f))]

for filename in csv_files:
  filepath = join(CLEANED_FOLDER, filename)
  account = EXPENSE_ACCOUNT if 'expenses' in filename else INCOME_ACCOUNT
  with open(filepath) as file:
    logging.info("reading", filepath)
    # read cleaned csv file
    reader = csv.reader(file, delimiter=',', quotechar='"')
    # skip headers
    next(reader, None)
    transactions_counter = 0
    for row in reader:
      transaction = dict({
        "date": datetime.strptime(row[0], '%d/%m/%Y'),
        "account": account,
        "description": row[2],
        "amount": float(row[3].replace(',', '.')),
        "tags": tags[filename][row[1]]
      })

      db.insert_one(transaction)
      transactions_counter += 1
    # log how transactions have been added
    logging.info(transactions_counter, "transactions added from", filename, "as", account)