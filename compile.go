package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

type ContractStateItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ContractState struct {
	Models []ContractStateItem `json:"models"`
}

type Export struct {
	AppState AppState `json:"app_state"`
}

type AppState struct {
	Bank BankState `json:"bank"`
}

type BankState struct {
	Balances []BankBalance `json:"balances"`
}

type BankBalance struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type Coin struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type Order struct {
	Owner        string `json:"owner"`
	OfferDenom   Denom  `json:"offer_denom"`
	OfferAmount  string `json:"offer_amount"`
	FilledAmount string `json:"filled_amount"`
}

type Claim struct {
	Amount string `json:"amount"`
}

type Denom struct {
	Native string `json:"native"`
}

var (
	Snapshot = make(map[string]uint64)
	AQLA     = "factory/kujira1xe0awk5planmtsmjel5xtx2hzhqdw5p8z66yqd/uaqla"
	FIN      = "kujira1nswv58h3acql85587rkusqx3zn7k9qx3a3je8wqd3xnw39erpwnsddsm8z"
	STAKING  = "kujira1la5qzckfzvhl3adqscj7l7l4dy42fevk9gdatkguqsm8qmnsy0psfmxl8q"
)

func main() {
	// Compile bank balances
	data, err := os.ReadFile("bank.json")
	if err != nil {
		panic(err)
	}
	var export Export
	if err := json.Unmarshal(data, &export); err != nil {
		panic(err)
	}

	for _, entry := range export.AppState.Bank.Balances {
		if entry.Address == FIN || entry.Address == STAKING {
			continue
		}
		for _, coin := range entry.Coins {

			if coin.Denom == AQLA {
				i, err := strconv.ParseUint(coin.Amount, 10, 64)
				if err != nil {
					panic(err)
				}

				Snapshot[entry.Address] = i
			}
		}

	}

	// Add staking balances
	data, err = os.ReadFile("staking.json")
	if err != nil {
		panic(err)
	}
	var staking ContractState
	if err := json.Unmarshal(data, &staking); err != nil {
		panic(err)
	}

	for _, entry := range staking.Models {
		keyBytes, err := hex.DecodeString(entry.Key)
		if err != nil || len(keyBytes) < 2 {
			continue
		}

		keyBytes = keyBytes[2:]
		keyStr := string(keyBytes)
		if strings.HasPrefix(keyStr, "stake") {
			finalKey := keyStr[5:] // Drop "stake" prefix
			decodedValue, err := base64.StdEncoding.DecodeString(entry.Value)
			if err != nil {
				panic(err)
			}

			i, err := strconv.ParseUint(strings.Trim(string(decodedValue), "\""), 10, 64)
			if err != nil {
				panic(err)
			}

			Snapshot[finalKey] += i
		}
		if strings.HasPrefix(keyStr, "claims") {
			finalKey := keyStr[6:] // Drop "claims" prefix

			var claims []Claim
			decodedValue, err := base64.StdEncoding.DecodeString(entry.Value)
			if err != nil {
				panic(err)
			}

			if err := json.Unmarshal(decodedValue, &claims); err != nil {
				continue
			}
			for _, claim := range claims {
				i, err := strconv.ParseUint(claim.Amount, 10, 64)
				if err != nil {
					panic(err)
				}

				Snapshot[finalKey] += i

			}

		}

	}

	// Add open orders
	data, err = os.ReadFile("fin.json")
	if err != nil {
		panic(err)
	}
	var fin ContractState
	if err := json.Unmarshal(data, &fin); err != nil {
		panic(err)
	}

	for _, entry := range fin.Models {
		keyBytes, err := hex.DecodeString(entry.Key)
		if err != nil || len(keyBytes) < 2 {
			continue
		}

		keyBytes = keyBytes[2:]
		keyStr := string(keyBytes)
		if strings.HasPrefix(keyStr, "order") {

			decodedValue, err := base64.StdEncoding.DecodeString(entry.Value)
			if err != nil {
				continue
			}

			var order Order
			if err := json.Unmarshal(decodedValue, &order); err != nil {
				continue
			}

			if order.OfferDenom.Native == AQLA {
				i, err := strconv.ParseUint(string(order.OfferAmount), 10, 64)
				if err != nil {
					panic(err)
				}

				Snapshot[order.Owner] += i

			} else {
				i, err := strconv.ParseUint(string(order.FilledAmount), 10, 64)
				if err != nil {
					panic(err)
				}

				Snapshot[order.Owner] += i

			}
		}
	}

	var checksum uint64

	for _, v := range Snapshot {
		checksum += v
	}

	log.Printf("Snapshot Total: %s", strconv.FormatUint(checksum, 10))

	// Create a CSV file
	file, err := os.Create("snapshot.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"Address", "Balance"})

	// Write each key-value pair to the CSV
	for k, v := range Snapshot {
		err := writer.Write([]string{k, strconv.FormatUint(v, 10)})
		if err != nil {
			panic(err)
		}
	}
}
