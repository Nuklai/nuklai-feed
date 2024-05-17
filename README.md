# Nuklai Feed

## Build & Run from Source

- Build

  ```bash
  ./scripts/build.sh
  ```

- Run

  ```bash
  ./build/nuklai-feed ./config.json
  ```

## Build & Run with Docker

- Build

  ```bash
  docker build -t nuklai-feed .
  ```

- Run

  ```bash
  docker container rm -f nuklai-feed;
  docker run -d -p 10592:10592 -v ./config.json:/app/config.json --name nuklai-feed nuklai-feed;
  ```

- Read the logs

  ```bash
  docker container logs -f nuklai-feed
  ```
