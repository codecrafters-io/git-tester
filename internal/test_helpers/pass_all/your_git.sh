#!/bin/sh

/tmp/test_helpers/git config --global user.email "you@example.com"
/tmp/test_helpers/git config --global user.name "Your Name"
/tmp/test_helpers/git config --global init.defaultBranch "main"

if [ "$1" = "write-tree" ]
then
  /tmp/test_helpers/git add .
fi

exec /tmp/test_helpers/git "$@"
