#!/usr/bin/env bash

echo "==> Generating certificate..."

if ! which lego > /dev/null; then
    echo "==> Installing lego..."
    go install github.com/go-acme/lego/v4/cmd/lego@latest
fi

##############################################################################
# functions
##############################################################################

usage()
{
	cat << EOF

Usage: create-certificate [options...] <domain> [ <domain> ... ]
Options:
  -e, --email EMAIL         Your email from your scaleway account
  -h, --help                This help
  -t, --test                Use staging API of Let's Encrypt for testing the script
  -d, --debug               Debug mode, print additional debug output
  -a, --action              The action to execute [run, renew, list]


The first domain parameter should be your main domain name with the subdomains following after it.

Example: $0 -e me@example.com example.com www.example.com

EOF
}

# general log messages
log()
{
	echo "#### ${1}"
}

# error messages
error()
{
	echo "ERROR: ${1}" >&2
}

# debug messages
debug()
{
	if [ "${VERBOSE}" = "true" ];
	then
		# do not output to stdout, else debug output from api_request would
		# become part of the function response
		echo "${1}" >&2
	fi
}

# last command
on_exit()
{
	debug "EXIT ${?}"
	exit
}

##############################################################################
# main
##############################################################################


# stop on error
set -e
trap on_exit EXIT

# defaults
# ACME API to use
API="https://acme-v02.api.letsencrypt.org/directory"
API_STAGING="https://acme-staging-v02.api.letsencrypt.org/directory"
ACTION="run"
CONTACT_EMAIL="hashicorp@scaleway.com"
DOMAINS=()
VERBOSE="false"

# arg handling
if [ ${#} -lt 1 ];
then
	error "Missing parameter"
	usage
	exit 1
fi

while [ ${#} -gt 0 ];
do
	ARG="${1}"
	case "${ARG}" in
		-e|--email)
			shift
			CONTACT_EMAIL="${1}"
			;;
		-h|--help)
			usage
			exit
			;;
		-t|--test)
			# use staging API for testing
			log "Using staging API"
			API="${API_STAGING}"
  		;;
    -a|--action)
      shift
      ACTION_LEGO="${1}"
      ;;
    -d|--debug)
     export TF_LOG="DEBUG"
     export SCW_DEBUG=1
     ;;
		*)
			X="${ARG/-*/}"
			if [ -z "${X}" ];
			then
				error "Unknown option"
				usage
				exit 1
			else
				DOMAINS[${#DOMAINS[@]}]="${ARG}"
			fi
	esac
	# shift the option flag, option flag values (if any) are shifted in case block
	shift
done

if [ ${#DOMAINS[@]} -eq 0 ];
then
	error "Domain missing"
	usage
	exit 1
fi

DOMAIN="${DOMAINS[0]}"

api_request()
{
  MAIL="${1}"
	DOMAINS_ARRAY="${2}"
	ACTION_ARG="${3}"
  LEGO_OUT="$(SCALEWAY_API_TOKEN=${SCW_SECRET_KEY} lego --server ${API} --accept-tos --email ${MAIL} --pem --dns scaleway --domains ${DOMAINS_ARRAY} ${ACTION_ARG})"
 	echo "${LEGO_OUT}"
 	return 0
}

api_request ${CONTACT_EMAIL} ${DOMAINS} ${ACTION_LEGO}

log "Finished."