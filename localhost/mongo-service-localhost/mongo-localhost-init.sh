#!/bin/bash
set -e

# dbUser is the userName used from application code to interact with databases and dbPwd is the password for this user.
# MONGO_INITDB_ROOT_USERNAME & MONGO_INITDB_ROOT_PASSWORD is the config for db admin.
# admin user is expected to be already created when this script executes. We use it here to authenticate as admin to create
# dbUser and databases.

echo ">>>>>>> trying to create database and users"
if [ -n "${MONGO_INITDB_ROOT_USERNAME:-}" ] && [ -n "${MONGO_INITDB_ROOT_PASSWORD:-}" ] && [ -n "${dbUserGoSession:-}" ] && [ -n "${dbPwdGoSession:-}" ] && [ -n "${dbUserEmailServer:-}" ] && [ -n "${dbPwdEmailServer:-}" ]; then
mongo -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD<<EOF
db=db.getSiblingDB('gosession');
use gosession;
db.users.createIndex( { email: 1 }, { unique: true } )
db.createUser({
  user:  '$dbUserGoSession',
  pwd: '$dbPwdGoSession',
  roles: [{
    role: 'readWrite',
    db: 'gosession'
  }]
});
db=db.getSiblingDB('emailserver');
use emailserver;
db.users.createIndex( { email: 1 }, { unique: true } )
db.createUser({
  user:  '$dbUserEmailServer',
  pwd: '$dbPwdEmailServer',
  roles: [{
    role: 'readWrite',
    db: 'emailserver'
  }]
});
EOF
else
    echo "MONGO_INITDB_ROOT_USERNAME,MONGO_INITDB_ROOT_PASSWORD,dbUserGoSession,dbPwdGoSession must be provided. Some of these are missing, hence exiting the database and user creation process"
    exit 403
fi