# fiatless
OpenSource crypto payment gateway


# Setting up

You are responsible for generating and storing securely seed phrases for your wallet. For example, you can generate seed phrase using next command:

```
docker run --rm python:3-alpine sh -c "pip install mnemonic && python -c 'from mnemonic import Mnemonic; print(Mnemonic(\"english\").generate())'"
```

Then you need to set seed phrase in environment variable:


```
docker run -d --name fiatless \
  -p 8000:8000 \
  --restart always \
  -e TRON_WALLET_SEED_PHRASE="your-seed-phrase-here" \
  devforth/fiatless:latest
```

Same in compose:

```
version: '3.8'

services:
  fiatless:
    image: devforth/fiatless:latest
    container_name: fiatless
    restart: always
    environment:
      TRON_WALLET_SEED_PHRASE: "your-seed-phrase-here"
    ports:
      - "8000:8000"
```

## Supported env variables


| Environment Variable                | Blockchain  | Acceptable Values                      | Example                               |
|--------------------------------------|------------|----------------------------------------|---------------------------------------|
| `BITCOIN_WALLET_SEED_PHRASE`        | Bitcoin    | (Any valid seed phrase)               | `word1 word2 ... word12`             |
| `BITCOIN_RPC_URL`                   | Bitcoin    | (Any valid RPC URL)                   | `https://mainnet.bitcoin.org`        |
| `BITCOIN_NETWORK`                    | Bitcoin    | `mainnet`, `testnet`, `signet`, `regtest` | `mainnet`                            |
| `BINANCE_WALLET_SEED_PHRASE`        | Binance    | (Any valid seed phrase)               | `word1 word2 ... word12`             |
| `BINANCE_RPC_URL`                   | Binance    | (Any valid RPC URL)                   | `https://bsc-dataseed.binance.org/`  |
| `BINANCE_NETWORK`                   | Binance    | `mainnet`, `testnet`                  | `mainnet`                            |
| `SOLANA_WALLET_SEED_PHRASE`         | Solana     | (Any valid seed phrase)               | `word1 word2 ... word12`             |
| `SOLANA_RPC_URL`                    | Solana     | (Any valid RPC URL)                   | `https://api.mainnet-beta.solana.com` |
| `SOLANA_NETWORK`                    | Solana     | `mainnet-beta`, `devnet`, `testnet`   | `mainnet-beta`                       |
| `ETHEREUM_WALLET_SEED_PHRASE`       | Ethereum   | (Any valid seed phrase)               | `word1 word2 ... word12`             |
| `ETHEREUM_RPC_URL`                  | Ethereum   | (Any valid RPC URL)                   | `https://mainnet.infura.io/v3/...`   |
| `ETHEREUM_NETWORK`                  | Ethereum   | `mainnet`, `goerli`, `sepolia`, `holesky` | `mainnet`                            |

