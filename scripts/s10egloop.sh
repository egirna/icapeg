#!/bin/bash
exec 3>&1 4>&2
trap 'exec 2>&4 1>&3' 0 1 2 3
exec 1>eg.txt 2>&1
countmain=1
for i in $(seq $countmain); do
count=30
for i in $(seq $count); do
filerb="./reb$i.pdf"
touch "$filerb"  && rm "$filerb" && time  c-icap-client -i 192.168.1.9  -p 1344 -s gw_rebuild  -f ./test.pdf -o "$filerb" -v &
wait
done
wait
done