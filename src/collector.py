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

CLEANED_FOLDER = os.path.abspath(os.getenv('CLEANED_FOLDER', 'cleaned'))
EXPENSE_ACCOUNT = 'Uscite'
INCOME_ACCOUNT = 'Entrate'

EXPENSES_TAGS_MAP = {
  "abbigliamento": ["Abbigliamento", "Uscite / Annuali"],
  "Abbonamenti": ["Abbonamenti", "Uscite / Mensili"],
  "Alimentari": ["Spesa", "Uscite / Mensili"],
  "Computer": ["Abbigliamento", "Uscite/Mensili"],
  "Elettronica": ["Abbigliamento", "Uscite/Mensili"],
  "Hobby": ["Abbigliamento", "Uscite/Mensili"],
  "Igiene": ["Abbigliamento", "Uscite/Mensili"],
  "Lavoro": ["Abbigliamento", "Uscite/Mensili"],
  "Prestiti": ["Abbigliamento", "Uscite/Mensili"],
  "Regali": ["Abbigliamento", "Uscite/Mensili"],
  "Ristoranti": ["Abbigliamento", "Uscite/Mensili"],
  "Servizi bancari": ["Abbigliamento", "Uscite/Mensili"],
  "Servizi finanziari": ["Abbigliamento", "Uscite/Mensili"],
  "Servizi Internet": ["Abbigliamento", "Uscite/Mensili"],
  "Spese mediche": ["Abbigliamento", "Uscite/Mensili"],
  "Telefono": ["Abbigliamento", "Uscite/Mensili"],
  "Utensili": ["Abbigliamento", "Uscite/Mensili"]
}

EXPENSES_AUTO_TAGS_MAP = {
  "Assicurazione": ["Abbigliamento", "Uscite/Mensili"],
  "Benzina": ["Abbigliamento", "Uscite/Mensili"],
  "Bollo": ["Abbigliamento", "Uscite/Mensili"],
  "GPL": ["Abbigliamento", "Uscite/Mensili"],
  "Parcheggio": ["Abbigliamento", "Uscite/Mensili"],
  "Riparazioni e manutenzione": ["Abbigliamento", "Uscite/Mensili"],
  "Varie": ["Abbigliamento", "Uscite/Mensili"]
}

if not os.path.isdir(CLEANED_FOLDER):
  print('Launch the cleaner before the collector')
  sys.exit()

# mongodb client
client = MongoClient()
db = client.collected

# read the csv files list
csv_files = [f for f in listdir(CLEANED_FOLDER) if isfile(join(CLEANED_FOLDER, f))]

for filename in csv_files:
  filepath = join(CLEANED_FOLDER, filename)
  account = EXPENSE_ACCOUNT if 'expenses' in filename else INCOME_ACCOUNT
  with open(filepath) as file:
    print("reading", filepath)
    # read cleaned csv file
    reader = csv.reader(file, delimiter=',', quotechar='"')
    # skip headers
    next(reader, None)
    for row in reader:
      transaction = {
        "date": datetime.strptime(row[0], '%d/%m/%Y'),
        "account": account,
        "description": row[2],
        "amount": float(row[3].replace(',', '.'))
      }

      transaction['tags'] = []

      db.insert_one(transaction)

