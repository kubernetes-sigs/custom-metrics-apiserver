# Release Process

The custom-metrics-apiserver library is released following Kubernetes releases.
The major and minor versions of custom-metrics-apiserver are in sync with
upstream Kubernetes and the patch version is reserved for changes in
custom-metrics-apiserver that we want to release without having to wait for the
next version of Kubernetes.

The process is as follow and should always be in sync with Kubernetes:

1. A new Kubernetes minor version is released
2. An issue is proposing a new release with a changelog containing the changes
   since the last minor release
3. At least one [approver](OWNERS) must LGTM this release
4. A PR that bumps Kubernetes dependencies to the latest version is created and
   merged. The major and minor version of the dependencies should be in sync
   with the version we are releasing.
5. An OWNER creates a draft GitHub release
6. An OWNER creates a release tag using `git tag -s $VERSION`, inserts the
   changelog and pushes the tag with `git push $VERSION`
7. An OWNER creates and pushes a release branch named `release-X.Y`
8. An OWNER publishes the GitHub release
9. An announcement email is sent to
   `kubernetes-sig-instrumentation@googlegroups.com` with the subject
   `[ANNOUNCE] custom-metrics-apiserver $VERSION is released`
10. The release issue is closed
