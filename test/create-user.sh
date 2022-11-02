#!/bin/bash

# This script creates a new user against the Megaport Staging API
# There is currently no error checking with the user registration.

if ! command -v jq &> /dev/null; then
    echo "jq (json cli) not installed"
    exit 1
fi

CHARSET="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345678"
CHARSETSYMBOL="${CHARSET}!@+="
ENDPOINT="https://api-staging.megaport.com/"
CREDENTIALFILE=".mpt_test_credentials"

genRandomString() {
    local length=$1
    local charset=$2


    for i in $(seq 1 ${length}); do
        result+="${charset:$(( RANDOM % ${#charset} )):1}"
    done

    echo $result
}

usernamePostfix=$(genRandomString 10 ${CHARSET})
username="golib${usernamePostfix}@sink.megaport.com"
password=$(genRandomString 20 ${CHARSETSYMBOL})

echo "Registering User"
regResp=$(curl -s -X POST -H "Content-Type: multipart/form-data" -H "Accept: application/json" -H "User-Agent: Go-Megaport-Library/0.1" \
    -F "firstName=Go" -F "lastName=Testing" -F "email=${username}" -F "password=${password}" -F "companyName=Go Testing Company" \
    "${ENDPOINT}/v2/social/registration")

token=$(echo $regResp | jq -r '.data.session')

cat > ${CREDENTIALFILE} <<EOF
export MEGAPORT_USERNAME="${username}"
export MEGAPORT_PASSWORD="${password}"
EOF

read -r -d '' companyEnable <<-EOC
{
    "tradingName": "Go Testing Company"
}
EOC

read -r -d '' market <<-EOC
{
    "currencyEnum": "AUD",
    "language": "en",
    "companyLegalIdentifier": "ABN987654",
    "companyLegalName": "Go Testing Company",
    "billingContactName": "Go Testing",
    "billingContactPhone": "0730000000",
    "billingContactEmail": "${username}",
    "address1": "Level 3, 825 Ann St,  QLD 4006",
    "city": "Fortitude Valley",
    "state": "QLD",
    "postcode": "4006",
    "country": "AU",
    "firstPartyId": 808
}
EOC

echo $compResp
echo "Registering Marketplace"
marketResp=$(curl -X POST -H "X-Auth-Token: ${token}" -H "Content-Type: application/json" -H "Accept: application/json" -H "User-Agent: Go-Megaport-Library/0.1" \
    -d "${market}" \
    "${ENDPOINT}/v2/market/" 2>/dev/null)
echo $marketResp
