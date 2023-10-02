#!/bin/sh

if [ "$1" = "write-tree" ]
then
  git add .
fi

exec git "$@"
