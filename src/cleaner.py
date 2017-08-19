#! ./bin/python3

'''
Example of data to clean:

Date	Account Name	Number	Description	Notes	Memo	Category	Type	Action	Reconcile	To With Sym	From With Sym	To Num.	From Num.	To Rate/Price	From Rate/Price
04/06/2016	Abbigliamento		maglietta			Sbilancio-EUR	T		N	€ 5,00		5			
Abbigliamento	S		N	€ 5,00		5		1	
Sbilancio-EUR	S		N		-€ 5,00		-5		1

Must become like this:

Date	Account Name	Description	To Num.
04/06/2016	Abbigliamento		maglietta 5

'''

DATE_COLUMN = 0
ACCOUNT_COLUMN = 1
NAME_COLUMN = 3
AMOUNT_COLUMN = 12

import csv
import os
from os import listdir
from os.path import isfile, join

SOURCE_FOLDER = os.path.abspath(os.getenv('SOURCE_FOLDER', 'samples'))
CLEANED_FOLDER = os.path.abspath(os.getenv('CLEANED_FOLDER', 'cleaned'))
csv_files = [f for f in listdir(SOURCE_FOLDER) if isfile(join(SOURCE_FOLDER, f))]

for filename in csv_files:
  file_path = join(SOURCE_FOLDER, filename)
  with open(file_path) as file:
    reader = csv.reader(file, delimiter=',', quotechar='"')
    file_write_path = join(CLEANED_FOLDER, filename)
    with open(file_write_path, 'w+') as file_write:
      for row in reader:
        if row[DATE_COLUMN]:
          writer = csv.writer(file_write, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)
          print(row[DATE_COLUMN], row[ACCOUNT_COLUMN], row[NAME_COLUMN], row[AMOUNT_COLUMN])
          writer.writerow([row[DATE_COLUMN], row[ACCOUNT_COLUMN], row[NAME_COLUMN], row[AMOUNT_COLUMN]])


        # with open('samples/expenses.csv') as expenses:
        #    reader = csv.reader(expenses, delimiter=',', quotechar='"')
        #    for row in reader:
        #        if row[DATE_COLUMN]:
        #            print(row[DATE_COLUMN], row[ACCOUNT_COLUMN], row[NAME_COLUMN], row[AMOUNT_COLUMN])
