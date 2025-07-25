#!/usr/bin/env bash
set -euo pipefail

DO_SPACES_ACCESS_KEY="${DO_SPACES_ACCESS_KEY:-}"
DO_SPACES_SECRET_KEY="${DO_SPACES_SECRET_KEY:-}"
DO_SPACES_ENDPOINT="${DO_SPACES_ENDPOINT:-sfo3.digitaloceanspaces.com}"
DO_SPACES_BUCKET_NAME="${DO_SPACES_BUCKET_NAME:-pictoria}"
S3CFG_LOCATION="${S3CFG_LOCATION:-/root/.s3cfg}"
SYNC_WAIT_PERIOD_SECONDS="${SYNC_WAIT_PERIOD_SECONDS:-60}"
SYNC_SOURCE_LOCATION="${SYNC_SOURCE_LOCATION:-/sync-source}"

if [ -z "${DO_SPACES_ACCESS_KEY}" ]; then
    echo "DO_SPACES_ACCESS_KEY Must be defined"
    exit 1
fi

if [ -z "${DO_SPACES_SECRET_KEY}" ]; then
    echo "DO_SPACES_SECRET_KEY Must be defined"
    exit 1
fi

echo "Using \$DO_SPACES_ENDPOINT=${DO_SPACES_ENDPOINT}"

sed -ie "s/^access_key = .*$/access_key = ${DO_SPACES_ACCESS_KEY}/" "${S3CFG_LOCATION}"
sed -ie "s|^secret_key = .*$|secret_key = ${DO_SPACES_SECRET_KEY}|" "${S3CFG_LOCATION}"
sed -ie "s/^host_base = .*$/host_base = ${DO_SPACES_ENDPOINT}/" "${S3CFG_LOCATION}"
# This tends to look something like 
# host_bucket = %(bucket)s.sfo3.digitaloceanspaces.com
# sed -ie "s/^host_bucket = .*$/host_bucket = ${DO_SPACES_ENDPOINT}/" "${S3CFG_LOCATION}"

# If you wanna leak secrets :)
# -e "secret_key = " \
grep \
    -e "access_key = " \
    -e "host_base = " \
    -e "host_bucket = " \
    "${S3CFG_LOCATION}"

while true; do
    echo "Running S3 Copy Operation"

    s3cmd sync "${SYNC_SOURCE_LOCATION}" "s3://${DO_SPACES_BUCKET_NAME}/uploads/"

    echo "Waiting ${SYNC_WAIT_PERIOD_SECONDS} second(s) before starting sync again"
    sleep "${SYNC_WAIT_PERIOD_SECONDS}"
done
