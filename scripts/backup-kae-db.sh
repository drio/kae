#!/bin/bash

set -e

sqlite3 /data/kae/kae.sqlite "VACUUM INTO '/tmp/db.kae'"
gzip /tmp/db.kae

aws s3 cp /tmp/db.kae.gz s3://drio-kae-backup/backup-`date +%d%H`.gz

curl https://kae.driohq.net/hb/vzndxvgbtlzqlkjrfkkz &> /dev/null
rm -f /tmp/db.kae /tmp/db.kae.gz
