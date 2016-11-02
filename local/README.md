# Local development

This directory contains scripts for running a self-contained Spinnaker
deployment with a Kubernetes backend, using Docker.

These scripts are based on the [Spinnaker Docker Compose][spin-compose]
scripts.


This documentation assumes you are running on macOS.

## Prereqs

### Install software

You must have installed the following softwarwe.

- [Docker for Mac][docker4mac]
- [corectl.app][corectl]
- [Kubernetes Solo cluster for macOS][kube-solo]

### Copy your kubeconfig file to config directory

Assuming your working directory is the location of this file, do:

```bash
cp ~/kube-solo/kube/kubeconfig ./config/kubeconfig
```

## Bringing up a deployment

Make sure that Docker, corectl.app, and Kuberentes solo are running.

### Start the containers

Use docker-compose to bring everything up.

```
make up
```

### Initialize the database tables

Create the tables in the mysql database

```
chaosmonkey migrate
```

### Upload a Docker container image to the registry

```
docker pull nginx && docker tag nginx localhost:5001/nginx
docker push localhost:5001/nginx
```

### Create an app in Spinnaker

1. Navigate to <http://localhost:9000>
1. Click the "Actions" button and choose "Create Application" from the drop-down

### Verify chaosmonkey sees eligible instaces

```
chaosmonkey eligible nginx default
```







[spin-compose]: https://github.com/spinnaker/spinnaker/tree/master/experimental/docker-compose
[docker4mac]: https://docs.docker.com/engine/installation/mac/
[corectl]: https://github.com/TheNewNormal/corectl.app
[kube-solo]: https://github.com/TheNewNormal/kube-solo-osx
