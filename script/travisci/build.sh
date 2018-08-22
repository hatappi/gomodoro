#!/bin/sh -ex

make depend
make build-assets
rm -rf pkg
bin/crosscompile

