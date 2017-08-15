#!/bin/sh

# Arguments:
#	1 -> Prod,dev,etc
#	2 -> UserDB CipherKey
#	3 -> Mongo Revel Password
#	4 -> Mongo Balancer Pass

if [ "$1" = "prod" ] ; then
	export HODLZONE_KEY=$2
	export MONGO_REVEL_PASS=$3
	export MONGO_BAL_PASS=$4
elif [ "$1" = "prodDev" ]; then
	#Setting is used for testing in prod mode on dev server
	export HODLZONE_KEY="0766f805c375d84f45554b835377744d92228708ac65b84284cfcb34901846355c"
	export MONGO_REVEL_PASS="MadeUpPass"
	export MONGO_BAL_PASS="MadeUpPass"
else
	export HODLZONE_KEY="0000000000000000000000000000000000000000000000000000000000000000"
	export MONGO_REVEL_PASS=""
	export MONGO_BAL_PASS=""
fi

revel run $1