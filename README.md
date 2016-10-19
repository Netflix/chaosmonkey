![logo](logo.png "logo")

Chaos Monkey randomly terminates virtual machine instances and containers that
run inside of your production environment. Exposing engineers to
failures more frequently incentivizes them to build resilient services.

### Requirements

This version of Chaos Monkey is fully integrated with [Spinnaker], the
continuous delivery platform that we use at Netflix. You must be managing your
apps with Spinnaker to use Chaos Monkey to terminate instances.

Chaos Monkey should work with any backend that Spinnaker supports (AWS, GCP,
Azure, Kubernetes, Cloud Foundry). It has been tested with AWS and Kubernetes.

### Install locally

To install the Chaos Monkey binary on your local machine:

```
go install github.com/netflix/chaosmonkey/bin/chaosmonkey
```

### How to deploy

See the [wiki](https://github.com/Netflix/chaosmonkey/wiki) for instructions on how to configure and deploy Chaos Monkey.

[Spinnaker]: http://www.spinnaker.io/

### Support

[Simian Army Google group](http://groups.google.com/group/simianarmy-users).
