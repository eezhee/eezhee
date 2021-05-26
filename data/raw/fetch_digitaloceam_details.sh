#!/usr/bin/env bash

# note, you need to set environment variables for the various cloud apis before running the script

# get digitalocean details
if [ -z ${DO_API_KEY} ]; then
  echo "DO_API_KEY is not set"
else
  echo "getting info from digitalocean"
  DO_API_URL="https://api.digitalocean.com/v2"
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/images?page=1&per_page=500" | json_pp > digitalocean-images.json   
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/regions?page=1&per_page=100" | json_pp > digitalocean-regions.json
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/sizes?page=1&per_page=200" | json_pp > digitalocean-sizes.json

fi

echo "done"