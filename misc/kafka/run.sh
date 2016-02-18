#!/usr/bin/env bash

set -e

cd $(dirname "$0")

gradle jar

echo "Running kafka..."
DIR=$(mktemp -d)
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
