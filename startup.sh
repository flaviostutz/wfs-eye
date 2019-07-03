#!/bin/bash

wfs-eye \
  --loglevel="$LOG_LEVEL" \
  --wfs-url="$WFS3_API_URL" \
  --mongo-dbname="$MONGO_DBNAME" \
  --mongo-address="$MONGO_ADDRESS" \
  --mongo-username=$MONGO_USERNAME \
  --mongo-password=$MONGO_PASSWORD

