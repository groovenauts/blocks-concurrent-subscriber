#! /bin/sh

# Trigger Docker Hub automated build
# https://docs.docker.com/docker-hub/builds/#remote-build-triggers

curl -H "Content-Type: application/json" \
     --data "{\"source_type\": \"Tag\", \"source_name\": \"${TRAVIS_TAG}\"}" \
     -X POST \
     https://registry.hub.docker.com/u/groovenauts/blocks-concurrent-subscriber/trigger/${DOCKER_HUB_TRIGGER_TOKEN}/

