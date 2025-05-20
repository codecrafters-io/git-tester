#!/bin/sh

if [ "$1" = "init" ]
then
  ein init "@$"
fi

if [ "$1" = "cat-file" ]
then
  gix cat "@$"
fi

exec gix "$@"
