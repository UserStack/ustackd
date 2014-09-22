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
    # default port for ustackd is 7654
    port = 7654
    
    # be default the daemon is in background, foreground at demand
    # by uncommenting foreground
    ; foreground
    
    # The backend to use
    backend = sqlite
    
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

    Customer
        * email (string)
        has many Users
        has many Groups
       
    User
        * firstname (string)
        * lastname (string)
        * email (string)
        * username (string)
        * password (string)
        * active (bool)
        has many Roles
        has many Permissions
        has many Groups
        
    Group
        * name (string)
        has many Permissions
        has many Roles
        
    Role
        * name (string)
        has many Permissions
        
    Permission
        * name (string)

## Backends

The backends in ustackd are based on a plugin mechanism. This way, ustackd
should be able to communicate with all possible backends.

### sqlite

Capability: Customer, User, Group, Role, Permission

Sqlite 3 implementation of the backend.

### postgresql

Capability: Customer, User, Group, Role, Permission

PostgreSQL backend implementation.

### redis

Capability: Customer, User, Group, Role, Permission

Redis backend implementation.

### mongodb

Capability: Customer, User, Group, Role, Permission

MongoDB backend implementation.

### pam

Capability: User, Group

PAM backend implementation.

### unix

Capability: User, Group

Unix backend implementation.

## Protocol

This section describes the protocol that is used to interface with the daemon.
