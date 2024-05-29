#!/usr/bin/env bash

# Check if .env file exists
if [ ! -f .env ]; then
  echo ".env file not found!"
  exit 1
fi

# Source the .env file to load environment variables
source .env

# Set default values for environment variables if not set
DB_HOST="${POSTGRES_HOST}"
DB_PORT="${POSTGRES_PORT}"
DB_USER="${POSTGRES_USER}"
DB_PASSWORD="${POSTGRES_PASSWORD}"
DB_NAME="${POSTGRES_DBNAME}"

# Function to check if psql is installed and install it if not
function check_and_install_psql() {
  if ! command -v psql &> /dev/null; then
    echo "psql is not installed. Attempting to install..."

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
      if command -v apt-get &> /dev/null; then
        sudo apt-get update
        sudo apt-get install -y postgresql-client
      elif command -v yum &> /dev/null; then
        sudo yum install -y postgresql
      else
        echo "Unsupported Linux package manager. Please install the PostgreSQL client manually."
        exit 1
      fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
      if command -v brew &> /dev/null; then
        brew install postgresql
      else
        echo "Homebrew is not installed. Please install Homebrew first or install the PostgreSQL client manually."
        exit 1
      fi
    else
      echo "Unsupported OS. Please install the PostgreSQL client manually."
      exit 1
    fi

    # Ensure psql is available after installation
    if ! command -v psql &> /dev/null; then
      echo "Error: psql installation failed. Please install the PostgreSQL client manually."
      exit 1
    fi
  fi
}

# Check and install psql if necessary
check_and_install_psql

function get_feed_by_txid() {
  local txid="$1"
  echo "Getting feed with TxID: $txid"
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT * FROM feeds WHERE txid='$txid';"
}

function get_all_feeds() {
  echo "Getting all feeds"
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT * FROM feeds;"
}

function get_feeds_by_user() {
  local user_address="$1"
  echo "Getting feeds for user: $user_address"
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT * FROM feeds WHERE address='$user_address';"
}

function usage() {
  echo "Usage: $0 {get-feed-by-txid|get-all-feeds|get-feeds-by-user} [args]"
}

# Ensure at least one argument is provided
if [ $# -eq 0 ]; then
  usage
  exit 1
fi

case "$1" in
  get-feed-by-txid)
    if [ -z "$2" ]; then
      echo "TxID is required"
      usage
      exit 1
    fi
    get_feed_by_txid "$2"
    ;;
  get-all-feeds)
    get_all_feeds
    ;;
  get-feeds-by-user)
    if [ -z "$2" ]; then
      echo "User address is required"
      usage
      exit 1
    fi
    get_feeds_by_user "$2"
    ;;
  *)
    usage
    ;;
esac
