# Lock Service

This document describes the design and functionality of the lock service.

## High Level Functionality
The lock service implements a synchronization mechanism that can be used by other applications. It implements functions for acquiring and releasing locks to objects present in the application. The lock to an object is leased for a certain duration to the caller. This is done to prevent starvation of other requests to the lock. When the entity that holds the lock is done using the corresponding object, it can release the lock. If the lock's lease expires then it will be released and assigned to the next waiting entity that wishes to acquire it.

Clients communicate with the lock server via HTTP. `node` contains code for a HTTP server. When this is started, incoming requests are routed to their corresponding functions in the lock service using route handlers defined in `route`.

## Low-level implementation
The lockservice uses the following data structure to keep track of lock acquisitions: 
```go
type SafeLockMap struct {
	LockMap map[string]string
	Mutex   sync.Mutex
}
```
On an `Acquire()` or `Release()`, the mutex of `SafeLockMap` is locked. The LockMap stores a mapping of the object that is locked to the sessionID that currently owns that object.

## Acquire
When a client wishes to acquire a lock, it sends an HTTP request to the lock service (an HTTP server) at the `/acquire` endpoint with a JSON encoded `LockRequest` struct. 

```go
type LockRequest struct {
	FileID string `json:"FileID"`
	SessionID string `json:"SessionID"`
}
```
The request contains information of 'what' (FileID) needs to be acquired and 'who' (SessionID) wishes to acquire it. The `SessionID` is important because if the object does end up being locked, then the lock service maps the objects to the sessionID that is leasing the lock in `SafeLockMap`. This is to ensure that only the process that acquired the lock has the ability to release it. Since the `SessionID` is unique to each session and is never exposed to a client process, it is unlikely that it can be spoofed. The server then routes this request to the `Acquire()` method defined in the lock service using a route handler. This method updates the lockmap with the acquisition if the lock is not already acquired. If the method is successful, a response with status code 200 is sent to the client that requested the lock

## Check Acquire
TODO

## Release
Either when a client wishes to release its lock or a session of a client expires, `Release()` is called for all corresponding locks. As in the case of `Acquire()`, a marshaled JSON of the `LockRequest` struct is sent to the lock server via HTTP to the `/release` endpoint. This struct would contain information of both the `object` that has to be released and the `sessionID`. The `sessionID` is important because it is used to ensure that only the process that requested the lock can release it. The condition for checking sessionID before performing a release would be: 
```go
if request.sessionID == SafeLockMap.LockMap[object] {
	//release
}
```
If the release condition is met, the `object:sessionID` mapping is deleted from the `SafeLockMap`

## Check Release

Is this needed?

## Lock Leasing (Expiry)
We implement a 'lazy' approach to determine when a lock expires. When acquiring a lock, the service notes the timestamp in a map that mirrors `Lockmap`.
```go
TimestampMap map[string]Time
duration Duration
```
It maps the object being locked to the timestamp at which it was locked. When the lockservice is required to verify if an entity posesses a lock or if a new entity wishes to acquire this lock, it can perform the following check:

```go
if _ ,ok := LockMap[object]; !ok || time.Now() > (TimestampMap[object] + duration)
```

If the condition is satisfied, then the lock can be acquired. The if statement first checks if the object has ever been acquired. If not, it need not evaluate the second condition and the new entity can acquire the lock directly. However, if it has been acquired some time in the past and is present in the LockMap, then an additional check is performed using the timestamp that was recorded when the lock was acquired.  

