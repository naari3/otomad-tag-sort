#!/bin/bash

set -eu

index_url=$1
index_url_base=$(dirname $index_url)

dest_dir=$2

curl ${index_url} | grep jsonl | while read -r line ; do
    if [[ ${line} =~ ^(.+?)\>([0-9]+\.jsonl)(.+?)$ ]]; then
        all=${BASH_REMATCH[0]}
        jsonl_name=${BASH_REMATCH[2]}

        echo ${jsonl_name}  # 4
        jsonl_url="${index_url_base}/${jsonl_name}"
        curl -o ${dest_dir}/${jsonl_name} ${jsonl_url}
    fi
done
