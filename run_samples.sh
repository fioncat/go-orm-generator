#!/bin/bash

# This script is to run generating sample code.
# If runing this script and exit normally, the
# program is running normally.

if command -v go-gendb >/dev/null 2>&1; then
	go-gendb version
else
	echo 'command "go-gendb" is not exists'
	echo 'Did you forget to install go-gendb?'
	echo 'Please run:'
	echo '    go install'
	echo 'to install go-gendb in your match first before running this script.'
	exit 1
fi

# Check if the database connection is set
go-gendb conn-get test_samples > /dev/null 2>&1

if [ $? -ne 0 ]; then
	echo 'Now we need to set a database connection first.'
	echo 'This database must have the data table under '
	echo 'the samples help file (see samples/README.md).'
	echo ''
	echo 'You may not have such a database, you can exit '
	echo 'the program to create it, and then execute the '
	echo 'script. '
	echo ''
	read -p 'Do you want to continue script execution?(y/n): ' ACK

	if [ "$ACK" != "y" ]; then
		echo 'exits'
		exit 0
	fi

	echo ''
	read -p 'Please input database address: ' ADDR
	read -p 'Please input database user: ' USER
	read -p 'Please input database password: ' PASS
	read -p 'Please input databse name: ' DB

	go-gendb conn-set test_samples $ADDR --user $USER --pass $PASS --db $DB
	if [ $? -ne 0 ]; then
		echo 'set connection failed, exits.'
		exit 1
	fi

	echo ''
	echo 'Set connection done, you can run "go-gendb conn-set test_samples ..."'
	echo 'to overwrite existing configuration.'
	echo ''
fi

echo 'Begin to execute samples...'
go-gendb gen --cache --log --conn test_samples samples/quickstart/oper.go
go-gendb gen --log samples/dynamic/user.go


