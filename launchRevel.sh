#!/bin/sh

# Arguments:
#	1 -> Prod,dev,etc
#	2 -> UserDB CipherKey
#	3 -> Mongo Revel Password
#	4 -> Mongo Balancer Pass

if [ "$1" = "prod" ]
	then
	export HODLZONE_KEY=$2
	export MONGO_REVEL_PASS=$3
	export MONGO_BAL_PASS=$4

else
	export HODLZONE_KEY="0000000000000000000000000000000000000000000000000000000000000000"
	export MONGO_REVEL_PASS=""
	export MONGO_BAL_PASS=""
fi

revel run $1