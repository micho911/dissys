# dissys

## MICHEL YILDIRIM NIKOLAJ MIKKELSEN

Ledger keeps the balance of a number of accounts
An account is named by string
Account has a balance
balances are integers, and can be negative
Init balances are 0 for all accounts
Ledger is simply a map from Strings to Ints
Each Peer holds the same Ledger.
The system executes Transactions.
Transactions has two names From and To and an Amount
Transaction moves Amount from account From to account To
When the system executes it floods the peer to peer network.