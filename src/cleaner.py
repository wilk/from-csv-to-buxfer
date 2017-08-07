#! /usr/bin/env python3

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
with open('samples/expenses.csv') as expenses:
    reader = csv.reader(expenses, delimiter=',', quotechar='"')
    for row in reader:
        if row[DATE_COLUMN]:
            print(row[DATE_COLUMN], row[ACCOUNT_COLUMN], row[NAME_COLUMN], row[AMOUNT_COLUMN])