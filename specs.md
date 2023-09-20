## how to design a digital ledger

- we are trying to offer an offline way to transact money, bypassing EBS
- we want to update accounts balances and at the same time store all of these transactions in a ledger
- we don't want to play around with whatever complicated logic that is there in noebs, instead we are just want to share references from the userID onto a dynamoDB instance


## move money logic
- grab a reference from both sender and receiver', better to use their mobile numbers
- update sender's and receivers respective balances
- this whole process shall be done in a transaction way

## ledger logic
- create a credit operation (crediting money from account A to account B)
- create a debit operation (debiting money from account A to account B)
- the info are: (amount, date, operation type, account_id)

## security considerations (post pilot)
- we need authentication to ensure that the operation is successful

## dynamo db tables

### account table
- account_id
- balance
- name
- created_at
- account_type [optional]

### ledger table
- transaction_id
- amount
- operation_type [credit, debit]
- account

### settlement

So far we do the settlement online