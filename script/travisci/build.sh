#!/bin/sh -ex

make depend
make bindata
rm -rf pkg
bin/crosscompile

