FROM alpine:3.20
RUN apk add --no-cache openssh-server openssh-client \
 && adduser -D -h /home/wisp -s /bin/sh wisp \
 && mkdir -p /run/sshd /mnt/hostkeys
COPY sshd_config /etc/ssh/sshd_config
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
EXPOSE 22
ENTRYPOINT ["/entrypoint.sh"]
