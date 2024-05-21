# Nuklai Feed

## Build & Run from Source

- Build

  ```bash
  ./scripts/build.sh
  ```

- Run

  ```bash
  cp .env.example .env;
  ./build/nuklai-feed
  ```

- Database Operations

  You can use the scripts/db.sh script to interact with the SQLite database.

  - Get All Feeds:

    ```bash
    ./scripts/db.sh get-all-transactions
    ```

  - Get Feed by TxID:

    ```bash
    ./scripts/db.sh get-feed-by-txid <TxID>
    ```

- Get Feeds by user:

  ```bash
  ./scripts/db.sh get-feeds-by-user <WalletAddress>
  ```

## Build & Run with Docker

- Build

  ```bash
  docker build -t nuklai-feed .
  ```

- Run

  ```bash
  ./scripts/run_docker.sh
  ```

- Stop the docker container

  ```bash
  docker container stop nuklai-feed
  ```
