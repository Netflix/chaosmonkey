![logo](logo.png "logo")

Chaos Monkey randomly terminates virtual machine instances and containers that
run inside of your production environment. Exposing engineers to
failures more frequently incentivizes them to build more resilient services.

This version of Chaos Monkey is fully integrated with [Spinnaker], the
continuous delivery platform that we use at Netflix. You must be managing your
apps with Spinnaker to be able to use Chaos Monkey.

Chaos Monkey should work with any backend that Spinnaker supports (AWS, GCP,
Azure, Kubernetes, Cloud Foundry). It has been tested with AWS and Kubernetes.

For instructions on how to deploy Chaos Monkey, see the [wiki](https://github.com/Netflix/chaosmonkey/wiki).

[Spinnaker]: http://www.spinnaker.io/
