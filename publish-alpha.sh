#!/bin/bash

set -e

circleci config pack src > orb.yml
circleci orb validate orb.yml
circleci orb publish orb.yml narrativescience/slack@dev:alpha5
rm -rf orb.yml
