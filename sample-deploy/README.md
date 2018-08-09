# Sample Deployment Files

These files can be used to deploy the sample adapter container.  You can
build that with `make sample-container`.  The Dockerfile describes the
container itself, while the [manifests](manifests) can be used to deploy
that container as a provider of the custom metrics and external metrics
APIs on the cluster.
