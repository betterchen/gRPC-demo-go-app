#!/bin/bash

for mod in api/proto/* ; do
  for ver in $mod/*; do
    for file in $ver/def/*; do
      OUT=$ver
      DIR=$ver/def
      FILE=${file##*/}

      mkdir -p $OUT/swagger

      protoc \
      -I /usr/local/include \
      -I $DIR \
      --go_out=plugins=gorpc:$OUT \
        $FILE

      protoc \
        -I /usr/local/include \
        -I $DIR \
        --grpc-gateway_out=logtostderr=true:$OUT \
        $FILE

      protoc \
        -I /usr/local/include \
        -I $DIR \
        --swagger_out=logtostderr=true:$OUT/swagger \
        $FILE

      echo "PB generated in $OUT: $FILE"

    done
  done
done
