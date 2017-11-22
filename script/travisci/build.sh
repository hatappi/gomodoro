#!/bin/sh -ex

make depend
rm -rf pkg
bin/crosscompile

