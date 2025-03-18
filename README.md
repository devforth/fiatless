# fiatless
OpenSource crypto payment gateway


# Setting up

You are responsible for generating and storing securely seed phrases for your wallet. For example, you can generate BIP-0039 seed phrase using next command:

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


| Environment Variable                | Blockchain  | Acceptable Values                       | Example                               | Default Value           |
|-------------------------------------|------------|------------------------------------------|---------------------------------------|-------------------------|
| `MODE`                              | All        | `mainnet`, `testnet`                     | `mainnet`                            | `testnet`               |
| `BITCOIN_WALLET_SEED_PHRASE`        | Bitcoin    | (Any BIP-0039 seed phrase)               | `word1 word2 ... word12`             | _(None, must be set)_   |
| `BINANCE_WALLET_SEED_PHRASE`        | Binance    | (Any BIP-0039 seed phrase)               | `word1 word2 ... word12`             | _(None, must be set)_   |
| `SOLANA_WALLET_SEED_PHRASE`         | Solana     | (Any BIP-0039 seed phrase)               | `word1 word2 ... word12`             | _(None, must be set)_   |
| `ETHEREUM_WALLET_SEED_PHRASE`       | Ethereum   | (Any BIP-0039 seed phrase)               | `word1 word2 ... word12`             | _(None, must be set)_   |
| `TRON_WALLET_SEED_PHRASE`           | Tron       | (Any BIP-0039 seed phrase)               | `word1 word2 ... word12`             | _(None, must be set)_   |
| `TRON_NODE`                         | Tron       | `TRONGRID`                               | `TRONGRID`                           | `TRONGRID`               |
| `TRON_TRONGRID_API_KEY`             | Tron       | (Any valid TronGrid API key)             | `your-trongrid-api-key`              | _(None, must be set)_   |


### Roadmap

Just an example of env var extending.

#### Custom blockchains


| Environment Variable                | Blockchain  | Acceptable Values                       | Example                               | Default Value           |
|-------------------------------------|------------|------------------------------------------|---------------------------------------|-------------------------|
| `TRON_NODE`                         | Tron       | `TRONGRID` | `CUSTOM`                    | `TRONGRID`                           | `TRONGRID`               |
| `TRON_CUSTOM_RPC_URL`               | Tron       | (Any valid RPC URL)                      | `TRONGRID`                           |  _(None, must be set)_              |
| `TRON_CUSTOM_API_KEY`               | Tron       | (Any valid API key)                      | `123141541`                          |  _(None, must be set)_               |
| `TRON_CUSTOM_API_KEY_HEADER`        | Tron       | (Any valid header name)                   | `Authorization`                      |  _(None, must be set)_              |


### QuickNode

| Environment Variable                | Blockchain  | Acceptable Values                       | Example                               | Default Value           |
|-------------------------------------|------------|------------------------------------------|---------------------------------------|-------------------------|
| `TRON_NODE`                         | Tron       | `TRONGRID` | `QUICKNODE`                    | `QUICKNODE`                           | `TRONGRID`               |
| `TRON_QUICKNODE_API_KEY`            | Tron       | (Any valid QUICKNODE API key)                      | `123141541`                     |  _(None, must be set)_               |


