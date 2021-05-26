#!/usr/bin/env bash

# get vultr details
echo "getting info from vultr"
curl -s --location --request GET 'https://api.vultr.com/v2/plans' | json_pp > vultr-plans.json
# curl --location --request GET 'https://api.vultr.com/v2/plans-metal'
curl -s --location --request GET 'https://api.vultr.com/v2/regions' | json_pp > vultr-regions.json
curl -s --location --request GET 'https://api.vultr.com/v2/os' | json_pp > vultr-os.json

echo "done"