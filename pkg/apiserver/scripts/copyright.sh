#!/usr/bin/env bash
# Copyright 2023 Linkall Inc.

ff=$(find ./handlers -type f -name "*.go" -print)

for f in ${ff}; do
	c=$(sed -n 1p $f |grep -ci Copyright)
	if [ $c -ne 0 ] ;then
		echo 'had add copyright :' $f
        else
		echo 'add copyright: ' $f
        sed -i '1i \// Copyright 2023 Linkall Inc.' $f
	fi
done
