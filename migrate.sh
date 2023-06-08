#!/bin/sh
# Description: migration for sqlite
# Dependencies: sqlite3
# Shell: POSIX compliant

DEFAULT_DIR="./migrations"
HELP_MESSAGE="migrate: [file] [up|down]"
NO_FILE_ERR="migrate: create database file first"

die() {
	echo "$1"
	exit
}

migrate_sqlite() {
	for f in "$DEFAULT_DIR"/*."$2".sql; do
		sqlite3 "$1" ".read $f"
	done
}

[ -z "$1" ] || [ -z "$2" ] && die "$HELP_MESSAGE"
[ "$1" = "--help" ] && die "$HELP_MESSAGE"

case $2 in 
	"up")
	migrate_sqlite "$1" "$2"
	;;
	"down")
	[ ! -f "$1" ] && die "$NO_FILE_ERR"
	migrate_sqlite "$1" "$2"
	;;
esac
