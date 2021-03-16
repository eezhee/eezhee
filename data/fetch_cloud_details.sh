#!/usr/bin/env bash

# note, you need to set environment variables for the various cloud apis before running the script


# get digitalocean details
if [ -z ${DO_API_KEY}]; then
  echo "DO_API_KEY is not set"
else
  echo "getting info from digitalocean"
  DO_API_URL="https://api.digitalocean.com/v2"
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/images?page=1&per_page=500" | json_pp > digitalocean-images.json   
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/regions" | json_pp > digitalocean-regions.json
  curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DO_API_KEY" \
  "$DO_API_URL/sizes" | json_pp > digitalocean-sizes.json

fi

# get linode details
echo "getting info from linode"
curl -s https://api.linode.com/v4/linode/types | json_pp > linode-types.json
curl -s https://api.linode.com/v4/regions | json_pp > linode-regions.json
curl -s https://api.linode.com/v4/images | json_pp > linode-images.json

# get vultr details
echo "getting info from vultr"
curl -s --location --request GET 'https://api.vultr.com/v2/plans' | json_pp > vultr-plans.json
# curl --location --request GET 'https://api.vultr.com/v2/plans-metal'
curl -s --location --request GET 'https://api.vultr.com/v2/regions' | json_pp > vultr-regions.json
curl -s --location --request GET 'https://api.vultr.com/v2/os' | json_pp > vultr-os.json

echo "done"