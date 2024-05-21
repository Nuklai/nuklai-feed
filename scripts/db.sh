#!/usr/bin/env bash
# Copyright (C) 2024, AllianceBlock. All rights reserved.
# See the file LICENSE for licensing terms.

DB_PATH=".nuklai-feed/db/feeds.db"

function get_feed_by_txid() {
  local txid="$1"
  echo "Getting feed with TxID: $txid"
  sqlite3 $DB_PATH "SELECT * FROM feeds WHERE txid='$txid';"
}

function get_all_feeds() {
  echo "Getting all feeds"
  sqlite3 $DB_PATH "SELECT * FROM feeds;"
}

function get_feeds_by_user() {
  local user_address="$1"
  echo "Getting feeds for user: $user_address"
  sqlite3 $DB_PATH "SELECT * FROM feeds WHERE address='$user_address';"
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
