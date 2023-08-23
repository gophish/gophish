#!/bin/bash

# set config for admin_server
if [ -n "${ADMIN_LISTEN_URL+set}" ] ; then
    jq -r \
        --arg ADMIN_LISTEN_URL "${ADMIN_LISTEN_URL}" \
        '.admin_server.listen_url = $ADMIN_LISTEN_URL' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${ADMIN_USE_TLS+set}" ] ; then
    jq -r \
        --argjson ADMIN_USE_TLS "${ADMIN_USE_TLS}" \
        '.admin_server.use_tls = $ADMIN_USE_TLS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${ADMIN_CERT_PATH+set}" ] ; then
    jq -r \
        --arg ADMIN_CERT_PATH "${ADMIN_CERT_PATH}" \
        '.admin_server.cert_path = $ADMIN_CERT_PATH' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${ADMIN_KEY_PATH+set}" ] ; then
    jq -r \
        --arg ADMIN_KEY_PATH "${ADMIN_KEY_PATH}" \
        '.admin_server.key_path = $ADMIN_KEY_PATH' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${ADMIN_TRUSTED_ORIGINS+set}" ] ; then
    jq -r \
        --arg ADMIN_TRUSTED_ORIGINS "${ADMIN_TRUSTED_ORIGINS}" \
        '.admin_server.trusted_origins = ($ADMIN_TRUSTED_ORIGINS|split(","))' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

# set config for phish_server
if [ -n "${PHISH_LISTEN_URL+set}" ] ; then
    jq -r \
        --arg PHISH_LISTEN_URL "${PHISH_LISTEN_URL}" \
        '.phish_server.listen_url = $PHISH_LISTEN_URL' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${PHISH_USE_TLS+set}" ] ; then
    jq -r \
        --argjson PHISH_USE_TLS "${PHISH_USE_TLS}" \
        '.phish_server.use_tls = $PHISH_USE_TLS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${PHISH_CERT_PATH+set}" ] ; then
    jq -r \
        --arg PHISH_CERT_PATH "${PHISH_CERT_PATH}" \
        '.phish_server.cert_path = $PHISH_CERT_PATH' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi
if [ -n "${PHISH_KEY_PATH+set}" ] ; then
    jq -r \
        --arg PHISH_KEY_PATH "${PHISH_KEY_PATH}" \
        '.phish_server.key_path = $PHISH_KEY_PATH' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

# set contact_address
if [ -n "${CONTACT_ADDRESS+set}" ] ; then
    jq -r \
        --arg CONTACT_ADDRESS "${CONTACT_ADDRESS}" \
        '.contact_address = $CONTACT_ADDRESS' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

# db_name has to be changed to mysql for mysql connection to work
if [ -n "${DB_NAME+set}" ] ; then
    jq -r \
        --arg DB_NAME "${DB_NAME}" \
        '.db_name = $DB_NAME' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

if [ -n "${DB_FILE_PATH+set}" ] ; then
    jq -r \
        --arg DB_FILE_PATH "${DB_FILE_PATH}" \
        '.db_path = $DB_FILE_PATH' config.json > config.json.tmp && \
        cat config.json.tmp > config.json
fi

echo "Runtime configuration: "
cat config.json

# start gophish
./gophish
