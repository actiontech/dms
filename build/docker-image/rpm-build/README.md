## Reference
https://hub.docker.com/r/jitakirin/rpmbuild

## Usage
Typical usage:
```
docker run --rm --volume=/path-to-source:/src --volume=/path-to-spec-dir:/spec --volume=/path-to-output:/out \
  rpmbuild-image dms.spec
```

You can also specify rpmbuild args:
```
docker run --rm --volume=/path-to-source:/src --volume=/path-to-spec-dir:/spec --volume=/path-to-output:/out \
  --env=RPM_ARGS="-bb" \
  rpmbuild-image dms.spec
```

If your package requires something from a non-core repo to build, you can add that repo using a PRE_BUILDDEP hook. It is an env variable that should contain an inline script or command to add the repo you need. E.g. for EPEL do:
```
docker run --rm --volume=/path-to-source:/src --volume=/path-to-spec-dir:/spec --volume=/path-to-output:/out \
  --env=PRE_BUILDDEP="yum install -y epel-release" \
  rpmbuild-image dms.spec
```

## Debugging
Set VERBOSE option in the environment (with -e VERBOSE=1 option to docker run) which will enable verbose output from the scripts and rpmbuild. 
```
docker run -it -e VERBOSE=1 --volume=/path-to-source:/src --volume=/path-to-spec-dir:/spec --volume=/path-to-output:/out \
  rpmbuild-image dms.spec
```
