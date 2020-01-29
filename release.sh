#!/bin/bash
# Publishes and releases the version of the Orb to the CCI repo. Reads in Semantic version for the arg.
# You must have a valid CCI token setup to run this script.

if [ $# -eq 0 ]
  then
    echo "The semantic version must be provided as an argument. I.E. '1.0.0'"
else
  circleci config pack src > orb.yml
  circleci orb validate orb.yml
  circleci orb publish orb.yml "narrativescience/slack@$1"
fi

