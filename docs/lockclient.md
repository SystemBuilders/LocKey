# The Lock Client

This document describes the design of the Lock Client with distinctions on seperable features.

## Begin the Lock Service
The Lock Client(LC) has a method which can be used to start the Lock Service(LS) on providing appropriate configuration. This method can be skipped and the LS can be started separately and then the config can be provided to the LC when creating its instance.

## Lock Acquire
User processes can use the LC's `Acquire` method to acquire a lock in the LS on providing appropriate descriptors. This method starts a private session for the user process and further makes a HTTP call to the LS with a combination of the parameters of the user process and the session parameters. The unique combination of the paramters is what ensures the security of the lock in the LS. The `Acquire` method involves a pre-check of session parameters and the user process that ensures no HTTP call is made without ensuring the authenticity of the user. Once the LS has allocated a lock to this user process an appropriate response is returned. 
  The lock lasts until the session and once the session expires, the LS removes the entry for the lock requested by this user either by a notification from the client or lazily when the next user requests for the lock.

### Existance of a Cache
If a cache exists in the client, the HTTP call can be saved. On a user process checking for the status of the lock, the LC can return on checking from the cache. No session needs to be established before the checking from the cache because, the lock might not be available and thus might result in wasting of the instantiation of a session.
  On lock expiration, the entry on the cache about the lock must also be invalidated.

## Lock Release
User processes can use the LC's `Release` method to release a lock in the LS on providing the appropriate descriptors. A user process can release a lock only if it has an active session - because the session provides neccessary authentication to the LC to release the said lock. If the session has expired, the lock would've been released anyway. The user process can query the LC for the same using the same method and receive appropriate responses. The lock release happens using a HTTP call to the LS.

### Existance of a Cache
If a cache exists in an LC, the release must handle cache invalidations when the lock has been release by the cache. These two actions of removing the entry of the lock acquisition in the LC and the LS cannot occur lazily(one instantly and other lazily, taken any combination) to protect consistency of the LS.

## Session management
Session management is key factor to the security of the locks in the LS. A session has to be established when any user process has to access the LS. The creation of a session will create the possibility of a session space with its own session parameters. These session parameters are a validation check for the user process to ensure that ONLY that user process has access to the locks it wishes to acquire and operate on.
  On creation of a session, the session parameters that exist are, the `sessionID`, the `clientID` and a `userID`. These three parameters together ensure that the locks acquired by this particular user process is protected from other user processes. The `sessionID` will be passed on to the user process on `connecting` to the LC and this `sessionID` must be used in the future by that process. 
   When a session is active on a lock, this lock (the lock descriptor) cannot engage in another session with a different user process. This is the executive check that uses the validation concept. The same user process can create new sessions with the LC with existing sessions or use the same session to obtain different locks whilst keeping in mind that once the session ends, all locks of that session will be discarded.

## Lock Watching

## Lock Pouncing

