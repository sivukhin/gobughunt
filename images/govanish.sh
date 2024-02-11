ls $1 | grep -q go.mod 
if [[ "$?" != "0" ]]
then
	echo "::skip no go.mod found"
	exit 0
fi
files=$(find $1 | grep '.go' | wc -l)
if [ "$files" -gt "500" ]
then
	echo "::skip too many files $files"
	exit 0
fi
govanish -format github -path $1
