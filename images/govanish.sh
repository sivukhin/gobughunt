ls $1 | grep -q go.mod 
if [[ "$?" != "0" ]]
then
	echo "::skip no go.mod found"
	exit 0
fi
govanish -format github -path $1
