#!/bin/sh

if [ "$1" = "cat-file" ]
then
  gix cat "@$"
fi

exec gix "$@"
