#!/usr/bin/env bash


echo "nvm use"
./noodles "nvm use" |wc -l
echo "rvm use"
./noodles "rvm use" |wc -l
echo "fuck"
./noodles "fuck" | wc -l
