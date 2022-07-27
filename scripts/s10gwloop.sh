#!/bin/bash
exec 3>&1 4>&2
trap 'exec 2>&4 1>&3' 0 1 2 3
exec 1>gw.txt 2>&1
countmain=1
for i in $(seq $countmain); do
count=30
for i in $(seq $count); do
filerb="./reb$i.pdf"
touch "$filerb"  && rm "$filerb" && time  c-icap-client -i 18.203.158.160  -p 1344 -s gw_rebuild  -f ./test.pdf -o "$filerb" -v &
done
wait
done