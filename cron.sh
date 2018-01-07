#!/bin/sh

while true; do
  bundle exec ruby app.rb > /var/www/masutaka.net/current/webroot/${SECRET}.atom
  sleep 900
done
