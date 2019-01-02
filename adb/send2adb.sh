#!/bin/bash
#

set -e


while read -p "@? " CMD
do
	FULLCMD=$(printf "%04x%s" ${#CMD} "${CMD}")
	echo -n "$FULLCMD"
done

exit 0
#echo "SEND: $FULLCMD"
echo -n "${PREFIX}$FULLCMD" | nc localhost 5037
echo ""
