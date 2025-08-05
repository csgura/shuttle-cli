U=$(echo "$1" | cut -d@ -f1 | cut -d: -f1)
HOSTNAME=$(echo "$1" | cut -d@ -f2 | cut -d: -f1)
PORT=$(echo "$1" | cut -d: -f2)
if [[ $U = $1 ]]; then #no user specified
	U=$(USER)
fi
if [[ $PORT = $1 ]]; then #no port specified
	PORT="22"
fi	
PASSWD=$(security find-generic-password -l $HOSTNAME -g 2>&1 1>/dev/null | cut -d'"' -f2)
if [[ ! `echo $PASSWD | grep 'The specified item could not be found in the keychain.'` = '' ]]; then
echo "Could not found password for host $HOSTNAME in the keychain"
echo -n "Password:"
read -s password
	security add-generic-password -a $U -l $HOSTNAME -s ssh -w $password
	PASSWD=$password
echo "Password for $HOSTNAME succesfully added to keychain."
fi
SSHPASS=$PASSWD sshpass -e ssh $1

if [ "$?" = 5 ]; then 
	echo "Password incorrect"
	echo -n "Password:"
	read -s password
	security delete-generic-password -a $U -l $HOSTNAME -s ssh
	security add-generic-password -a $U -l $HOSTNAME -s ssh -w $password
	
	PASSWD=$password
	echo "Password for $HOSTNAME succesfully added to keychain."
	SSHPASS=$PASSWD sshpass -e ssh $1

fi
