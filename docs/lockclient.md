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

## Session management

## Lock Watching

## Lock Pouncing

