#!/usr/bin/env bash

# get linode details
echo "getting info from linode"
curl -s https://api.linode.com/v4/linode/types | json_pp > linode-types.json
curl -s https://api.linode.com/v4/regions | json_pp > linode-regions.json
curl -s https://api.linode.com/v4/images | json_pp > linode-images.json

echo "done"