#!/bin/sh

while true; do
  bundle exec ruby app.rb
  cp ${SECRET}.atom /var/www/masutaka.net/current/webroot
  sleep 900
done
