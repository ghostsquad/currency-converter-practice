#!/bin/bash

# TODO why isn't this a go test?? 🤦
set -euxo pipefail

HOSTNAME="${1}"
HOSTPORT="${2}"

# TODO keep your bash script libraries and stuff open source
# This is a rewrite of stuff you've done in the past, but don't have access to
# Because it was committed to private repositories
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

TestGet () {
  local url="$1"
  shift
  local extraArgs=$@
  success="false"

  n=0
  until [ "$n" -ge 5 ]
  do
    if http --check-status --ignore-stdin --timeout=1 ${extraArgs} GET "${url}"; then
      success="true"
      echo 'OK!' 1>&2 && break
    else
      case $? in
        2) echo 'Request timed out!' 1>&2 ;;
        3) echo 'Unexpected HTTP 3xx Redirection!' 1>&2;;
        4) echo 'HTTP 4xx Client Error!' 1>&2;;
        5) echo 'HTTP 5xx Server Error!' 1>&2;;
        6) echo 'Exceeded --max-redirects=<n> redirects!' 1>&2;;
        *) echo 'Other Error!' 1>&2;;
      esac
    fi
    n=$((n+1))
    sleep 0.1
  done

  if [ "${success}" != "true" ]; then
    printf "${RED}TEST FAILURE${NC}\n"
    exit 1
  fi
}

TestGet "http://${HOSTNAME}:${HOSTPORT}/ping" 1>&2
testResult=$(TestGet "http://${HOSTNAME}:${HOSTPORT}/convert/eur/1/jpy" --body)

TestResultKey () {
  local key="$1"
  echo "${testResult}" | jq -r --arg key "${key}" '.[$key]'
}

rate="$(TestResultKey "rate")"
toAmount="$(TestResultKey "toAmount")"

if [ "${rate}" != "${toAmount}" ]; then
  echo "expected ${rate}, got ${toAmount}" 1>&2
  printf "${RED}TEST FAILURE${NC}\n"
  exit 1
fi

printf "${GREEN}SUCCESS!${NC}\n" 1>&2

exit 0
