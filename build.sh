
#TODO
# create separate script to delete cluster.  needs ID in a file
# try and run k3sup on the VM.  does it work?
# move k3sup to a bash script
# try and deploy a container. does it work?
# try and map dns to node_ip - is port what we expect?
# now do the same with AWS & GCP & AZURE

APP_NAME=sample
# get app name from directory name or deploy file

if [ -f "deploy_state.json" ]; then
  # TODO should check if VM is still running. file could be old
  VM_ID=`cat deploy_state.json | jq '.vm_id'`
  DROPLET_DETAILS=`doctl compute droplet get $VM_ID -o json`
  DROPLET_STATUS=`echo $DROPLET_DETAILS| jq '.[0].status'`
  if [ $DROPLET_STATUS == '"active"' ]; then
    echo "app has already running on droplet $VM_ID.  need to destroy before can create new instance"
    exit 1
  fi
fi

# server name
# should we add branch name (eg sample-master-cluster)
# only if not 'master'
# check if repo has git (should!!!)
if [ -d ".git" ]; then
  GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
  if [ $GIT_BRANCH == 'master' ]; then
    BRANCH=''
  else
    BRANCH=${GIT_BRANCH}-
  fi
else
  BRANCH=''
fi
VM_NAME=${APP_NAME}-${BRANCH}cluster

# do setup
# need API key (or login using doctl)
# check token there.  otherwise doctl not setup yet
DO_ACCESS_TOKEN=`cat "${HOME}/Library/Application Support/doctl/config.yaml" | egrep access-token | cut -f 2 -d ' '`

# ssh key in $HOME/.ssh/id_rsa
SSH_KEYGEN=`ssh-keygen -l -E md5 -f $HOME/.ssh/id_rsa`
FINGERPRINT=`echo $SSH_KEYGEN | cut -f 2 -d ' '`
DO_FINGERPRINT=${FINGERPRINT#MD5:}
# need to check if has been uploaded to DO
DO_SSH_KEYS=`doctl compute ssh-key list`
if [[ ${DO_SSH_KEYS} != *${DO_FINGERPRINT}* ]]; then
  echo "Need to upload SSH key to DO"
  # new_key_name should be same as $HOME
  # doctl import new_key_name --public-key-file ~/.ssh/id_rsa.pub
fi

# do
VM_IMAGE=ubuntu-20-04-x64
# doctl compute image list --public | egrep 'ubunut-'
VM_SIZE=s-1vcpu-1gb
# doctl compute size list
VM_REGION=tor1
# doctl compute region list
# "nyc1", "sfo1","nyc2","ams2","sgp1","lon1","nyc3","ams3","fra1","tor1","sfo2","blr1","sfo3"
# select region based on country we are in.  if country has several do ping test
OWN_IP=`dig +short myip.opendns.com @resolver1.opendns.com`
GEO_INFO=`curl https://ipvigilante.com/$OWN_IP`
COUNTRY=`echo $GEO_INFO | jq '.data.country_name'`
echo $COUNTRY
# use country to map to a DO geo
# if country=USA, need to decide between NYC & SFO.  ping test?
# http://digitalocean.com/geo/google.csv has all their IPs
# SFO: 45.55.0.1
# NYC: 45.55.100.1
# DC_IP_ADDR=45.55.0.1
# PING_TIME=`ping -c 4 $DC_IP_ADDR | tail -1| awk '{print $4}' | cut -d '/' -f 2`
# do for both DCs and choose smaller number

# setup
# instal doctl
# doctl auth init

# create the VM
RESULT=`doctl compute droplet create $VM_NAME --image $VM_IMAGE --size $VM_SIZE --region $VM_REGION --ssh-keys $DO_FINGERPRINT -o json` 
ERROR=`echo $RESULT | jq '.errors'`
if [[ -z $ERROR ]]; then
  VM_ID=`echo $RESULT | jq '.[0].id'`
  echo $VM_ID
  echo "{ \"vm_id\": ${VM_ID} }" > "deploy_state.json"
else
  echo $RESULT
fi

# if [[ $(echo $RESULT | jq 'keys[0]') == 'errors' ]]; then
#   echo "could not create VM"
# fi

# VM_ID=zzz
# wait for it to be active
DROPLET_STATE=`doctl compute droplet get 208893003 -o json`
