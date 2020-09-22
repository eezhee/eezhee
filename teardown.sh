
# read deploy_state.json
if [ -f "deploy_state.json" ]; then

  # get droplet ID
  VM_ID=`cat deploy_state.json | jq '.vm_id'`

  # should check if VM is still running. file could be old
  DROPLET_DETAILS=`doctl compute droplet get $VM_ID -o json`
  DROPLET_STATUS=`echo $DROPLET_DETAILS| jq '.[0].status'`
  if [ $DROPLET_STATUS == '"active"' ]; then
    doctl compute droplet delete $VM_ID
    rm deploy_state.json
    exit 0
  fi
else
  echo "app is not running"
fi
