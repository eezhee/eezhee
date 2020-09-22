

APP_NAME=sample
VM_NAME=${APP_NAME}-cluster
# should we add branch name (eg sample-master-cluster)

# do setup
# need API key (or login using doctl)
# token stored in 'access-token' ${HOME}/Library/Application Support/doctl/config.yaml

# ssh key in $HOME/.ssh/id_rsa
SSH_KEYGEN=`ssh-keygen -l -E md5 -f $HOME/.ssh/id_rsa`
FINGERPRINT=`echo $SSH_KEYGEN | cut -f 2 -d ' '`
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

# setup
# instal doctl
# doctl auth init

# create the VM
# TODO add --ssh-keys
RESULT=`doctl compute droplet create $VM_NAME --image $VM_IMAGE --size $VM_SIZE --region $VM_REGION -o json` 
ERROR=`echo $RESULT | jq '.errors'`
if [[ -z $ERROR ]]; then
  VM_ID=`echo $RESULT | jq '.[0].id'`
  echo $VM_ID
else
  echo $RESULT
fi

# if [[ $(echo $RESULT | jq 'keys[0]') == 'errors' ]]; then
#   echo "could not create VM"
# fi

# VM_ID=zzz

# doctl compute droplet delete $VM_ID