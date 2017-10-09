#!/usr/bin/env bash

declare -a arr=( # http://stackoverflow.com/questions/8880603/loop-through-array-of-strings-in-bash-script
#	"api_common_dto"
#	"api_user"
#	"api_receipt"
	"api_counterparties"
#	"auth_google"
#	"api_transfer"
#	"api_transfers"
#    "api_bills"
)

for f in "${arr[@]}"
do
	echo "Removing ${f}_ffjson.go...."
	rm "${f}_ffjson.go"
done

for f in "${arr[@]}"
do
   echo "Regenerating ${f}..."
   ffjson "${f}.go"
done