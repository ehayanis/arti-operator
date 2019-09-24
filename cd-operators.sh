#!/bin/sh

set -e
set -o nounset

NUMBER_OF_TRY=30

curl -kL -u ${AWX_LOGIN}:${AWX_PASSWORD} -H "Content-Type: application/json" https://awx.os.prodinfo.gca/api/v2/workflow_job_templates/?name="$1"

JOB_TEMPLATE=$(curl -kL -u ${AWX_LOGIN}:${AWX_PASSWORD} -H "Content-Type: application/json" https://awx.os.prodinfo.gca/api/v2/workflow_job_templates/?name="$1" | jq '.results | .[0].id')

JOB_RUNNING_ID=$(curl -kv -u ${AWX_LOGIN}:${AWX_PASSWORD} -X POST -H "Content-Type: application/json" https://awx.os.prodinfo.gca/api/v2/workflow_job_templates/$JOB_TEMPLATE/launch/ | jq .workflow_job)

for i in $(seq 1 $NUMBER_OF_TRY)
do
  RESULT=$(curl -kL -u ${AWX_LOGIN}:${AWX_PASSWORD} -H "Content-Type: application/json" "https://awx.os.prodinfo.gca/api/v2/workflow_jobs/$JOB_RUNNING_ID" | jq .status)
  if [ $RESULT != "\"running\"" ] && [ $RESULT != "\"pending\"" ]
  then
    echo "Le job $JOB_RUNNING_ID est terminé. Status: $RESULT"
    break
  fi
  sleep 30
done

if [ "$RESULT" = "\"successful\"" ]
then
  true
else
  false
fi