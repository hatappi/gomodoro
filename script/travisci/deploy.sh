#!/bin/sh -ex
VERSION=`echo $TRAVIS_TAG | sed -e "s/v//g"`

echo "*** Compression start ***"

ls pkg | grep -v tar.gz | xargs -I{} tar -zcvf pkg/{}-${VERSION}.tar.gz pkg/{}

echo "*** $VERSION deploy start ***"

export GITHUB_TOKEN=$GITHUB_TOKEN
ghr $TRAVIS_TAG pkg/

echo "*** deploy end ***"

