#!/bin/bash
SERVICE=$1
DIAG_PATH="$(pwd)/diagnostics/${SERVICE}"

mkdir -p ${DIAG_PATH}
DIAG_PATH=$(realpath $DIAG_PATH)
echo "saving diagnostics of $SERVICE in $DIAG_PATH"

rm ${DIAG_PATH}/${SERVICE}.log; for i in {1..1000}; do curl -H "Authorization: validToken" -ks http://localhost:8080/${SERVICE}/service | jq | grep "instance" >> ${DIAG_PATH}/${SERVICE}.log; done

cat ${DIAG_PATH}/${SERVICE}.log | sort | uniq -c > ${DIAG_PATH}/${SERVICE}_results.log