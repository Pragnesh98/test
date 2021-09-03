#!/bin/sh

#source vars if file exists
DEFAULT=/etc/default/fluentd

if [ -r $DEFAULT ]; then
    set -o allexport
    . $DEFAULT
    set +o allexport
fi

# If the user has supplied only arguments append them to `fluentd` command
if [ "${1#-}" != "$1" ]; then
    set -- fluentd "$@"
fi

# If user does not supply config file or plugins, use the default
if [ "$1" = "fluentd" ]; then
    if ! echo $@ | grep ' \-c' ; then
       set -- "$@" -c /fluentd/etc/${FLUENTD_CONF}
    fi

    if ! echo $@ | grep ' \-p' ; then
       set -- "$@" -p /fluentd/plugins
    fi
fi

HOSTNAME=`hostname`
if [ ! -z $AZURE_ACCOUNT ]; then
  sed -i "s%azure_storage_account.*%azure_storage_account $AZURE_ACCOUNT%" /fluentd/etc/fluent.conf
  sed -i "s%azure_storage_access_key.*%azure_storage_access_key $AZURE_ACCESS_KEY%" /fluentd/etc/fluent.conf
  sed -i "s%azure_container.*%azure_container $AZURE_CONTAINER%" /fluentd/etc/fluent.conf
  sed -i "s%path yellowmessenger.com.*%path $HOSTNAME/%" /fluentd/etc/fluent.conf
fi

cat /fluentd/etc/fluent.conf

exec "$@"
