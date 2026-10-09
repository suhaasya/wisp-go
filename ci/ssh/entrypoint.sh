#!/bin/sh
set -eu
mkdir -p /run/sshd /home/wisp/.ssh
cp /mnt/hostkeys/ssh_host_* /etc/ssh/
chmod 600 /etc/ssh/ssh_host_*_key
chmod 644 /etc/ssh/ssh_host_*_key.pub
cp /mnt/authorized_keys /home/wisp/.ssh/authorized_keys
chown -R wisp:wisp /home/wisp/.ssh
chmod 700 /home/wisp/.ssh
chmod 600 /home/wisp/.ssh/authorized_keys
exec /usr/sbin/sshd -D -e
