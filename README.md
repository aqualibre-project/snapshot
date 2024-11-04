# AQLA Snapshot

Snapshot Balance: 279,897,705.369792. Includes wallet balances, staked balances, pending unstaking, un-filled limit orders.

Shortfall of 102,294.630208 from 280m total supply due to FIN filled order amounts missing (requires live query on the smart contract)

Data can be queried with a node with correct history:

```
kujirad query wasm contract-state all kujira1la5qzckfzvhl3adqscj7l7l4dy42fevk9gdatkguqsm8qmnsy0psfmxl8q  --height 24585406 -o json > staking.json

kujirad query wasm contract-state all kujira1nswv58h3acql85587rkusqx3zn7k9qx3a3je8wqd3xnw39erpwnsddsm8z --height 24585406 -o json > fin.json

kujirad export --modules-to-export bank --height 24585406 --output-document bank.json
```

And then running `go run compile.go`
