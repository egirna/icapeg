#!/bin/bash


test_file () {
# get hash of original file
HashCode=$(sha256sum $1 | cut -d " " -f 1)

# run test code to get file back 

c-icap-client -i 127.0.0.1  -p 1344 -s echo -f "./$1" -o ./resultFile -no204

# check the hash

result=$(echo "$HashCode resultFile" | sha256sum --check | cut -d ":" -f 2)
#  delete hashed file 
rm -r resultFile

x="$1 ==> Result: $result , Expected: $2" 
if ( [ "$result" = " OK" ] && [ "$2" = "OK" ] ) || ( [ "$result" = " FAILED" ] && [ "$2" = "FAILED" ] ); then
    printf "\033[1;32m ✔ $x\n\033[0m"
    ((OK+=1))
else
    printf "\033[1;31m ✘ $x\n\033[0m"
    ((FAILED+=1))
fi 


}

OK=0
FAILED=0
while read line
do
   fileName=$(echo "$line" | cut -d "," -f 1)
   expectedResult=$(echo "$line" | cut -d "," -f 2)
   test_file "./testing/$fileName" "$expectedResult"
done < ./testing/input.csv
Total=$(($OK+$FAILED))
printf "\n\033[1;35m ------------------ Result -----------------------\n"

printf "   \033[1;33m Total: $Total\n   \033[1;32m ✔ Passed: $OK\n   \033[1;31m ✘ not passed: $FAILED  \n\n"

if [ $FAILED -gt 0 ]
then
    exit 50
fi
# c-icap-client -i 127.0.0.1  -p 1344 -s echo -f "./testing/somebook.pdf" -o ./resultFile 