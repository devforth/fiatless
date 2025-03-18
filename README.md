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
