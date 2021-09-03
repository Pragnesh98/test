FROM fluent/fluentd:v1.10-debian-1

# Use root account to use apt
USER root

# below RUN includes plugin as examples elasticsearch is not required
# you may customize including plugins as you wish
RUN buildDeps="sudo make gcc g++ libc-dev" \
 && apt-get update \
 && apt-get install -y --no-install-recommends $buildDeps \
 && apt-get install -y libxml2-dev libxslt-dev build-essential \
 && gem install bundler \
 && gem install fluentd --no-doc \
 && sudo gem install fluent-plugin-route \
 && sudo gem install fluent-plugin-azurestorage \
 && sudo gem sources --clear-all \
 && SUDO_FORCE_REMOVE=yes \
    apt-get purge -y --auto-remove \
                  -o APT::AutoRemove::RecommendsImportant=false \
                  $buildDeps \
 && rm -rf /var/lib/apt/lists/* \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

RUN mkdir -p /var/log/fluent && chown fluent /var/log/fluent

COPY fluent.conf /fluentd/etc/
COPY entrypoint.sh /bin/

ENV AZURE_ACCOUNT="yellowmessenger"
ENV AZURE_ACCESS_KEY="/qnaZM6BMIwCaeRvhbhbrfzb6t3npUBZDxVMSDgdnJSozd7O3zdurau+yNUXf/FO7XdP1zhQQCnX9/jeyIg2ew=="
ENV AZURE_CONTAINER="telephony-logs"

RUN ["chmod", "+x", "/bin/entrypoint.sh"]

USER fluent
