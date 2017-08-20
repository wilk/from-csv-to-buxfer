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
  # nested tags need the following syntax: "<parent> / <child>"
  "Abbigliamento": ["Abbigliamento", "Uscite / Annuali"],
  "Abbonamenti": ["Abbonamenti", "Uscite / Mensili"],
  "Alimentari": ["Spesa", "Uscite / Mensili"],
  "Computer": ["Elettronica", "Uscite / Annuali"],
  "Elettronica": ["Elettronica", "Uscite / Annuali"],
  "Hobby": ["Intrattenimento / Libri", "Uscite / Annuali"],
  "Igiene": ["Estetica", "Uscite / Mensili"],
  "Lavoro": ["Spese Lavoro / Wilk", "Uscite / Mensili"],
  "Prestiti": ["Casa / Restituzione Prestito", "Uscite / Straordinarie"],
  "Regali": ["Regali Effettuati", "Uscite / Annuali"],
  "Ristoranti": ["Ristoranti Bar", "Uscite / Mensili"],
  "Servizi bancari": ["Servizi Bancari", "Uscite / Mensili"],
  "Servizi finanziari": ["Servizi Fiscali", "Uscite / Annuali"],
  "Servizi Internet": ["Casa / Servizi / ADSL", "Uscite / Mensili"],
  "Spese mediche": ["Spese Mediche", "Uscite / Mensili"],
  "Telefono": ["Telefonia", "Uscite / Mensili"],
  "Utensili": ["Utensili", "Uscite / Annuali"]
}

EXPENSES_AUTO_TAGS_MAP = {
  "Assicurazione": ["Auto / Assicurazione", "Uscite / Annuali"],
  "Benzina": ["Auto / Benzina", "Uscite / Mensili"],
  "Bollo": ["Auto / Bollo", "Uscite / Mensili"],
  "GPL": ["Auto / GPL", "Uscite / Mensili"],
  "Parcheggio": ["Auto / Parcheggio", "Uscite / Mensili"],
  "Riparazioni e manutenzione": ["Auto / Manutenzione", "Uscite / Annuali"],
  "Varie": ["Auto / Varie", "Uscite / Straordinarie"]
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

