#!/usr/bin/env bash

while getopts ":t:c:p:g:v:" opt; do
  case $opt in
    t) TARGETS="$OPTARG"
    ;;
    c) CONFIG_PARAMS="$OPTARG"
    ;;
    p) PROFILE="$OPTARG"
    ;;
    g) TAGS="$OPTARG"
    ;;
    v) VERBOSE="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    exit 1
    ;;
  esac

  case $OPTARG in
    -*) echo "Option $opt needs a valid argument"
    exit 1
    ;;
  esac
done

if ! command -v pulumi &> /dev/null; then
  echo "pulumi CLI not found. Pulumi needs to be installed on the system.
        See https://github.com/DataDog/test-infra-definitions/blob/main/README.md"
  exit
fi

# Defaults
if [[ -z "${TARGETS}" ]]; then
  TARGETS="./test/e2e"
fi

if [[ -z "${VERBOSE}" ]]; then
  VERBOSE="-v"
fi

if [[ -z "${PROFILE}" ]]; then
  PROFILE="local"
fi

TAGS="${TAGS}"

if [[ ! ("${PROFILE}" == "ci" || "${PROFILE}" == "local") ]]; then
  echo "Unknown profile: ${PROFILE}. Valid values are 'local' or 'ci'."
  exit
fi

if [[ "${PROFILE}" == "local" ]]; then
  msg="Profile is ${PROFILE}, but missing"
  if [[ -z $PULUMI_CONFIG_PASSPHRASE || -z $PULUMI_CONFIG_PASSPHRASE ]]; then
    echo "${msg} Pulumi passphrase environment variable. Set environment variable
          'PULUMI_CONFIG_PASSPHRASE' or 'PULUMI_CONFIG_PASSPHRASE_FILE' to continue."
    exit
  fi
  if [[ -z "${E2E_API_KEY}" ]]; then
       echo "${msg} 'E2E_API_KEY' environment variable. Set environment variable to continue."
       exit
  fi
  AWS_KEYPAIR_NAME=$USER
  if [[ -z ${AWS_KEYPAIR_NAME} ]]; then
    echo "${msg} AWS keypair name. Tried environment variable 'USER', but not found. Configure AWS keypair in
          AWS console and set USER environment variable to continue."
    exit
  fi
fi

declare -A parsed_params
parsed_params["ddinfra:aws/defaultKeyPairName"]="${AWS_KEYPAIR_NAME}"

if [[ $CONFIG_PARAMS ]]; then
  while read -d, -r config; do
    IFS='=' read -r key val <<<"$config"
    parsed_params[$key]=$val
  done <<<"$CONFIG_PARAMS,"
fi


config_json=$(jq -n --argjson n "${#parsed_params[@]}" '
  { config:
    (reduce range($n) as $i ({}; .[$ARGS.positional[$i]] = $ARGS.positional[$i+$n]))
  }' --args "${!parsed_params[@]}" "${parsed_params[@]}")

export PULUMI_CONFIGS=$config_json
export DD_TAGS=$TAGS
export DD_TEAM="container-ecosystems"

cmd="gotestsum --format pkgname --packages=${TARGETS} -- ${VERBOSE} -vet=off -timeout 1h -count=1"

$cmd
