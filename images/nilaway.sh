cd $1

ls . | grep -q go.mod 
if [[ "$?" != "0" ]]
then
	echo "::skip no go.mod found"
	exit 0
fi
files=$(find . | grep '.go' | wc -l)
if [ "$files" -gt "500" ]
then
	echo "::skip too many files $files"
	exit 0
fi
mod=$(cat go.mod | grep '^module ' | grep -o '[^ ]*$')
result=$(nilaway -pretty-print=false $mod 2>&1 | perl -pe "s/^([^\s]+):(\d+):(\d+):/::error file=\1,line=\2::/g")
replaced=${result/$(pwd)\//}
echo "$replaced"
