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
    # Interface and port where the daemon should listen
    listen = 0.0.0.0:7654
    ; listen = 127.0.0.1:7654

    # the realm send by the server after connect
    realm = ustackd $VERSION$
    
    # be default the daemon is in background, foreground at demand
    # in foreground mode the syslog is disabled and logging appears on STDOUT
    ; foreground = yes
    
    # The backend to use
    backend = sqlite

    # path where to store the pid file
    pid = ./ustackd.pid

    [syslog]
    # which syslog facility should be used
    facility = 3 # (system daemons)

    # set the syslog log level
    # (Emergency, Alert, Critical, Error, Warning, Notice, Informational, Debug)
    level = Debug
    
    [client]
    # client that is allowed to issue all commands (e.g. web gui)
    auth = 42421da75756d69832d:deny://
        
    # client that is restricted to certain commands (e.g. auth server)
    auth = 6d95e4ac638daf4b786:allow:/^(login|set|get|change (password|name))/
    
    # client that can manage everything, but is secure from data stealing
    auth = 04d6eb93ab5d30f7bb0:deny:/^(users|groups|group users)/
    
    [security]
    # change root to this location after start
    ; chroot = /var/run/ustackd
        
    # change user to this location after start
    # the same is used for the gid, so you need to have the user only in one group
    # with the same name
    ; uid = ustack
    
    [ssl]
    # status
    enabled = yes
    
    # Interface and Port where the daemon should listen with ssl/tls enabled
    ; listen = 0.0.0.0:8765
    
    # location of the private key in pem format 
    ; key = /etc/ustack/key.pem
    
    # location of the certificate in pem format
    ; cert = /etc/ustack/cert.pem
    
    # protocols to support
    # SSL v3.0 = 768
    # TLS v1.0 = 769
    # TLS v1.1 = 770
    # TLS v1.2 = 771
    ; protocol-min = 771
    ; protocol-max = 771
    
    # ciphers to support
    ; ciphers = ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:ECDH+3DES:DH+3DES:RSA+AESGCM:RSA+AES:RSA+3DES:!aNULL:!MD5:!DSS
        
    [sqlite]
    url = ustack.db
    
## Daemon command line options

    ustackd [-c config file] [-f|--foreground]
    
If now config file is passed, the file will be searched in the following 
locations in order:

* ./ustack.conf
* /etc/ustack.conf
* /usr/local/etc/ustack.conf

## Start hacking

Simply download the dependencies and start the server:

    make prepare
    go run ustackd.go -f

## Domain Model

    User
        * uid (int)
        * firstname (string)
        * lastname (string)
        * name (string)
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

    [Daemon]
    backend = sqlite
    
    [sqlite]
    url = /var/run/ustack.db
    
Or to use a memory database for testing
    
    [sqlite]
    url = :memory:

### proxy

Capability: Customer, User

Proxy backend implementation connects to a different ustackd and proxies requests.

    [Daemon]
    backend = sqlite
    
    [proxy]
    host = 127.0.0.1:7543
    ssl = yes

### nil

Capability: Customer, User

Nil backend implementation is a dummy implementation, that always returns ok.

    [Daemon]
    backend = nil

## Protocol

This section describes the protocol that is used to interface with the daemon.

Following notation is used

    -> Client sends something to the server
    <- Server send something back to the client
    
CRLF "\r\n" is implicit for every line sent. If the request was ok the response
is prefixed with a "+" otherwise with a minus, followed by the response code.

### Login

If a secret is set, the client has to issue the client auth command in order
to get access to the system. Depending on the secret the capabilities may
change. This is useful, to for example not allow apps to list all users.
Generally consider use of SSL/TLS!

    -> client auth <secret>
    <- + OK (user group admin)

Return Codes:

    OK: Ok with a list of privileges
    EPERM: no valid secret

### General

#### Stats

Return stats of the server.

    -> stats
    <- logins: 13435
    <- err logins: 1123
    <- users: 651
    <- inactive users: 15
    <- groups: 4
    <- + OK

#### Start tls/ssl

Upgrades the current connection into a ssl connection.

    -> starttls
    <SSL connection from know on>

### User Commands

#### Create user

    -> user <name> <password>
    <- + OK 1

Return Codes:

    OK: Ok with the uid
    EEXIST: User already exists
    EINVAL: Parameter missing or invalid

#### Disable user

    -> disable <name|uid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name or uid unknown

#### Enable user

    -> enable <name|uid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name or uid unknown

#### Store data on the user object

    -> set <name|uid> <key> <value>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name or uid unknown
    EINVAL: Parameter missing or invalid

Recommended Keys:

    firstname
    lastname

#### Get stored user object data

    -> get <name,uid> <key>
    <- <value>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name, uid or key unknown
    EINVAL: Parameter missing or invalid

#### Login

    -> login <name> <password>
    <- + OK 1

Return Codes:

    OK: Ok with the uid
    EPERM: name and password are not a valid combination

#### Change password

    -> change password <name|uid> <password> <newpassword>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name or uid unknown
    EPERM: name and password are not a valid combination
    EINVAL: Parameter missing or invalid

#### Change name

    -> change name <name|uid> <password> <newname>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: name or uid unknown
    EPERM: name and password are not a valid combination
    EINVAL: Parameter missing or invalid

#### List all groups of a user

    -> user groups <name|uid>
    <- administrators:1
    <- sales:20
    <- engineering:10
    <- + OK

Format:

    List of groups with group id: <group>:<gid>

Return Codes:

    OK: Ok with the list of objects
    ENOENT: name or uid unknown
    EINVAL: Parameter missing or invalid
    
#### Delete user

    delete user <name|uid>

Return Codes:

    OK: Ok user deleted
    ENOENT: name or uid unknown
    EINVAL: Parameter missing or invalid
    
#### All users

    -> users
    <- foo@bar.com:1
    <- bar@example.com:2
    <- mr@bean.com:3
    <- + OK

Format:

    List of names with user id: <name>:<uid>

Return Codes:

    OK: Ok

### Group Commands

#### Create Group

    -> group <name>
    <- + OK 1

Return Codes:

    OK: Ok with the gid
    EEXIST: Group already exists
    EINVAL: Parameter missing or invalid

#### Add user to group

    -> add <name|uid> <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group or user doesn't exist
    
#### Remove user from group

    -> remove <name|uid> <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group or user doesn't exist

#### Delete group, user, permission, role

    -> delete group <group|gid>
    <- + OK

Return Codes:

    OK: Ok
    ENOENT: Group doesn't exist
    
#### Groups

    -> groups
    <- administrators:1
    <- sales:20
    <- engineering:10
    <- + OK

Format:
    
    List of groups with group id: <group>:<gid>

Return Codes:

    OK: Ok

#### Users of a group

    -> group users <group|gid>
    <- foo@bar.com:1
    <- bar@example.com:2
    <- mr@bean.com:3
    <- + OK

Format:

    List of names with user id: <name>:<uid>

Return Codes:

    OK: Ok
    ENOENT: Group doesn't exist
