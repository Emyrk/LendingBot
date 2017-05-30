# Deployment

Notes regarding deployment

## Golang + Revel

Install golang 1.8.4

Get revel cli to launc

## Nginx

We have an Nginx proxy that handles the SSL and http --> https redirect. Revel serves on port 9000, and Nginx redirects port 80 to 443, then 443 to 9000


Nginx `/etc/nginx/sites-available/default`

```
# http --> https
server {
        listen 80 default_server;
        listen [::]:80 default_server;
        server_name _;
        return 301 https://$host$request_uri;
}
```


Nginx `/etc/nginx/sites-available/hodl.zone.conf`

```
# www.hodl.zone --> hodl.zone
server {
    server_name www.hodl.zone;
    return 301 $scheme://hodl.zone$request_uri;
}

# hodl.zone:443 --> 9000
server {
     server_name hodl.zone;

     location ~* ^/ {
       rewrite ^/(.*) /$1 break;
       #proxy_redirect http:///$host/  /$1/;
       proxy_set_header Host $host;
       proxy_pass http://localhost:9000;
       proxy_pass_request_headers on;
     }

    listen 443 ssl; # managed by Certbot
ssl_certificate /etc/letsencrypt/live/hodl.zone/fullchain.pem; # managed by Certbot
ssl_certificate_key /etc/letsencrypt/live/hodl.zone/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot

}
```


## User settings ~/.profile

```
# GoLang pathing
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin


# LendingBot App paths
export USER_DB=$HOME/database/users/UserDB.db
export SEC51_KEYPATH=$HOME/database/keys/
export LOG_PATH=$HOME/database/log/hodlzone.log

# Convenience path
export HODL=$GOPATH/src/github.com/Emyrk/LendingBot                                                          
```

## Running


```
cd $GOPATH/src/github.com/Emyrk/LendingBot
revel run prod
```
