#!/bin/bash

if [ -x "$(command -v docker-compose)" ]; then
  DOCKER_COMPOSE="docker-compose"
elif docker compose version > /dev/null 2>&1; then
  DOCKER_COMPOSE="docker compose"
else
  echo 'Error: neither docker-compose nor docker compose is installed.' >&2
  exit 1
fi

echo "Using command: $DOCKER_COMPOSE"

domains=(dockr.ru)
rsa_key_size=4096
data_path="./certbot"
email="developerinspb@gmail.com"
staging=1

if [ -d "$data_path" ]; then
  read -p "Existing data found for $domains. Continue and replace existing certificate? (y/N) " decision
  if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
    exit
  fi
fi


if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "### Downloading recommended TLS parameters ..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
  echo
fi

echo "### Creating dummy certificate for $domains ..."
path="/etc/letsencrypt/live/$domains"
mkdir -p "$data_path/conf/live/$domains"
$DOCKER_COMPOSE -f docker-compose.prod.yaml run --rm --entrypoint "\
  openssl req -x509 -nodes -newkey rsa:$rsa_key_size -days 1\
    -keyout '$path/privkey.pem' \
    -out '$path/fullchain.pem' \
    -subj '/CN=localhost'" certbot
echo


echo "### Starting nginx (with init config) ..."
$DOCKER_COMPOSE -f docker-compose.prod.yaml -f <(echo "
services:
  nginx:
    volumes:
      - ./nginx/init.conf:/etc/nginx/conf.d/default.conf
") up --force-recreate -d nginx
echo

echo "### Deleting dummy certificate ..."
$DOCKER_COMPOSE -f docker-compose.prod.yaml run --rm --entrypoint "\
  rm -Rf /etc/letsencrypt/live/$domains && \
  rm -Rf /etc/letsencrypt/archive/$domains && \
  rm -Rf /etc/letsencrypt/renewal/$domains.conf" certbot
echo


echo "### Requesting Let's Encrypt certificate ..."
domain_args=""
for domain in "${domains[@]}"; do
  domain_args="$domain_args -d $domain"
done

# Select appropriate email arg
case "$email" in
  "") email_arg="--register-unsafely-without-email" ;;
  *) email_arg="-m $email" ;;
esac

# Enable staging mode if needed
if [ $staging != "0" ]; then staging_arg="--staging"; fi

$DOCKER_COMPOSE -f docker-compose.prod.yaml run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    $email_arg \
    $domain_args \
    --rsa-key-size $rsa_key_size \
    --agree-tos \
    --force-renewal" certbot
echo

echo "### Stopping nginx ..."
$DOCKER_COMPOSE -f docker-compose.prod.yaml stop nginx
echo
