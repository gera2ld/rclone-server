export DOCKER_PREFIX := 'gera2ld'

default:
  just --list

build:
  docker build -t $DOCKER_PREFIX/rclone-server .

push:
  docker push $DOCKER_PREFIX/rclone-server
