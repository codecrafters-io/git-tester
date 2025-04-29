#!/bin/sh

git config --global user.email "you@example.com"
git config --global user.name "Your Name"

if [ "$1" = "write-tree" ]
then
  git add .
fi

exec git "$@"
