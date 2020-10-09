# LocKey

We like to describe LocKey as a software that can be individually deployed and used as a distributed lock service. It is made of the following major components, the client and the service.

## Components
### Lock Client
The lock client is a library that enables the user to completely setup a lock service on a desired domain and use it. The client looks like this:
```
type Client interface {
  StartService() error
  Connect() session.Session
  Acquire(lockDescriptor) error
  Release(lockDescriptor) error
}
```

Any implementation of a client should involve the above function implementations to display its features.


### Lock Service

This is the core of LocKey that maintains the locks. This can be deployed in a distributed manner and the client can adopt to this based on some configuration changes.
The service looks like this:
```
type LockService interface {
  Acquire(descriptors) error
  Release(descriptors) error
  CheckAcquired(descriptors) error
  CheckReleased(descriptors) error
}
```
## High Level Design

Following is a high level design of the system:

![LocKey](https://github.com/SystemBuilders/LocKey/blob/master/docs/LocKey.png?raw=true)
