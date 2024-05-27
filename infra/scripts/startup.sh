#!/bin/bash
APP_DIR="/app"

echo "NUKLAI_RPC="$NUKLAI_RPC"" >> ${APP_DIR}/.env
echo "RECIPIENT="$RECIPIENT"" >> ${APP_DIR}/.env
echo "ADMIN_TOKEN="$ADMIN_TOKEN"" >> ${APP_DIR}/.env

echo "${@}" | xargs -I % sh -c '%'