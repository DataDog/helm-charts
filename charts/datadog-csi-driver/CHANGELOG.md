# Changelog

## 0.4.3

* Fix AllowlistSynchronizer helper

## 0.4.2

* Add gke AllowlistSynchronizer

## 0.4.1

* Mount `apm-socket` and `dsd-socket` to CSI node server container in readonly mode.
* Mount `plugins-dir` to node registrar container in readonly mode.

## 0.4.0

* Set node server image tag to `1.0.0`.

## 0.3.4

* Remove `hostNetwork: true` from csi driver daemonset.

## 0.3.3

* Fix bug that caused to pass the socket's parent directory to the start command arguments instead of the full socket path.

## 0.3.2

* Add option to configure CSI registrar image

## 0.3.1

* Fix image pull secrets of the CSI driver daemonset.

## 0.3.0

* Support configuring different host socket paths for apm and dogstatsd sockets. 
 
## 0.2.0

* Support configuring apm and dogstatsd sockets hostpaths. 

## 0.1.0

* Initial version
