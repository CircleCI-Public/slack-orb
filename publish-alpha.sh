#!/bin/bash

set -e

circleci config pack src > orb.yml
circleci orb validate orb.yml
circleci orb publish orb.yml narrativescience/slack@0.1.0
rm -rf orb.yml
