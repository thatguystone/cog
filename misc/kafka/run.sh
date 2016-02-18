#!/usr/bin/env bash

set -e

gradle jar

DIR=$(mktemp -d)

echo "Running kafka..."
CMD="java -jar build/libs/kafka.jar $DIR"

case "$1" in
travis)
	nohup $CMD > /dev/null &
	while ! nc -q 1 localhost 9092 < /dev/null; do sleep .1; done
	;;

*)
	trap "rm -r $DIR" EXIT
	$CMD
	;;
esac
