#!/bin/bash

set -e

for i in $@; do
  useradd $i
done
