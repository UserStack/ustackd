# ustackd

Project to learn golang.org in the context of user lifecycle management.

## Description

This repository contains the heart of the UserStack project. The core domain
model is implemented at the ustackd. The daemon has multiple backend
implementations in which the data can be stored.

## Influence

We may stole ideas from:

* Apache Syncope
* ConnID

## Configure the daemon

This section describes the configuration of the ustackd.

    [Daemon]
    # comma separated list of interfaces to listen on
    interfaces = 0.0.0.0
    
    # default port for ustackd is 7654
    port = 7654
    
    # the realm send by the server after connect
    realm = ustackd $VERSION$
    
    # be default the daemon is in background, foreground at demand
    # by uncommenting foreground
    ; foreground
    
    # The backend to use
    backend = sqlite
    
    # If enabled, the ssl section will be used to allow an encrypted connection
    ; ssl
    
    # Secret that needs to be passed after connect
    ; secret = 42421da75756d69832de50c3ab34f68ab5118b53
    
    [logging]
    # which syslog facility should be used
    facility = 3 # (system daemons)
    
    # set the syslog log level
    # (Emergency, Alert, Critical, Error, Warning, Notice, Informational, Debug)
    level = Debug
    
    [security]
    # change root to this location after start
    ; chroot = /var/run/ustackd
    
    # change group to this location after start
    ; gid = ustack
    
    # change user to this location after start
    ; uid = ustack
    
    [ssl]
    # Port where the daemon should listen with ssl/tls enabled
    ; port = 8765
    
    # location of the private key in pem format 
    ; key = /etc/ustack/key.pem
    
    # location of the certificate in pem format
    ; cert = /etc/ustack/cert.pem
    
    # protocols to support
    ; protocol = SSLv3 TLSv1 TLSv1.1 TLSv1.2
    
    # ciphers to support
    ; ciphers = ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:ECDH+3DES:DH+3DES:RSA+AESGCM:RSA+AES:RSA+3DES:!aNULL:!MD5:!DSS
        
    [sqlite]
    url = ustack.db
    
    
## Daemon command line options

    ustackd [-f config file]
    
If now config file is passed, the file will be searched in the following 
locations in order:

* ./ustack.conf
* /etc/ustack.conf
* /usr/local/etc/ustack.conf

## Domain Model

    User
        * uid (int)
        * firstname (string)
        * lastname (string)
        * email (string)
        * password (string)
        * active (bool)
        has many Groups
        
    Group
        * gid (int)
        * name (string)
        has many Users

## Backends

The backends in ustackd are based on a plugin mechanism. This way, ustackd
should be able to communicate with all possible backends.

### sqlite

Capability: Customer, User

Sqlite 3 implementation of the backend.

### postgresql

Capability: Customer, User

PostgreSQL backend implementation.

### redis

Capability: Customer, User

Redis backend implementation.

### mongodb

Capability: Customer, User

MongoDB backend implementation.

### pam

Capability: User, Group

PAM backend implementation.

### unix

Capability: User, Group

Unix backend implementation.

## Protocol

This section describes the protocol that is used to interface with the daemon.

Following notation is used

    -> Client sends something to the server
    <- Server send something back to the client
    
CRLF "\r\n" is implicit for every line sent. If the request was ok the response
is prefixed with a "+" otherwise with a minus, followed by the response code.

### Login

If a secret is set, the client has to issue the login command. Consider SSL!

    -> login secret
    <- + OK

Return Codes:

    OK: Ok
    EPERM: no valid secret

### General

    -> stats
    <- logins: 13435
    <- err logins: 1123
    <- users: 651
    <- inactive users: 15
    <- groups: 4
    <- + OK
    

### User Commands

#### Create user

    -> user email password
    <- + OK 1

Return Codes:

    OK: Ok with the uid
    EEXIST: User already exists
    EINVAL: Parameter missing or invalid

#### Disable user

    -> disable <email|uid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown

#### Enable user

    -> enable <email|uid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown

#### Store data on the user object

    -> set <email,uid> <key> <value>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown
    EINVAL: Parameter missing or invalid

Recommended Keys:

    firstname
    lastname

#### Get stored user object data

    -> get <email,uid> <key>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown
    EINVAL: Parameter missing or invalid

#### Login

    -> login email password
    <- + OK 1

Return Codes:

    OK: Ok with the uid
    EPERM: email and password are not a valid combination

#### Change password

    -> change password <email|uid> <password> <newpassword>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown
    EPERM: email and password are not a valid combination
    EINVAL: Parameter missing or invalid

#### Change email

    -> change email <email|uid> <password> <newemail>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: email or uid unknown
    EPERM: email and password are not a valid combination
    EINVAL: Parameter missing or invalid

#### List all groups of a user

    -> user groups <email,uid>
    <- administrators:1
    <- sales:20
    <- engineering:10
    <- + OK

Format:

    List of groups with group id: <group>:<gid>

Return Codes:

    OK: Ok with the list of objects
    ENOENT: email or uid unknown
    EINVAL: Parameter missing or invalid
    
#### Delete user

    delete user <email,uid>

Return Codes:

    OK: Ok user deleted
    ENOENT: email or uid unknown

### Group Commands

#### Create Group

    -> group fooo
    <- + OK 1

Return Codes:

    OK: Ok with the gid
    EEXIST: Group already exists
    EINVAL: Parameter missing or invalid

#### Add user to group

    -> add <user|uid> to <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group or user doesn't exist
    
#### Remove user from group

    -> remove <user|uid> to <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group or user doesn't exist

### delete group, user, permission, role

    -> delete group <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group doesn't exist
    
### groups

    -> groups
    <- administrators:1
    <- sales:20
    <- engineering:10
    <- + OK

Return Codes:

    OK: Ok
