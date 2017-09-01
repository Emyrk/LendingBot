# LendingBot

Add route to prod.

## Running

```
export USER_DB=$HOME/database/users/UserDB.db
export SEC51_KEYPATH=$HOME/database/keys/
export LOG_PATH=$HOME/database/log/hodlzone.log
export USER_STATS_DB=$HOME/database/users/UserStatistics.db
export REVEL_LOG=$HOME/database/
export INVITE_DB=$HOME/database/InviteCodes.db
```


Run this separtely 

```
export HODLZONE_KEY=KEY
export COINBASE_ACCESS_KEY=API-KEY
export COINBASE_SECRET_KEY=SEC-KEY
export MONGO_ADMIN_PASS=$1
export MONGO_REVEL_PASS=$2
export MONGO_BAL_PASS=$3
export MONGO_BEE_PASS=$4
``

# Welcome to Revel

A high-productivity web framework for the [Go language](http://www.golang.org/).


### Start the web server:

   revel run myapp

### Go to http://localhost:9000/ and you'll see:

    "It works"

## Code Layout

The directory structure of a generated Revel application:

    conf/             Configuration directory
        app.conf      Main app configuration file
        routes        Routes definition file

    app/              App sources
        init.go       Interceptor registration
        controllers/  App controllers go here
        views/        Templates directory

    messages/         Message files

    public/           Public static assets
        css/          CSS files
        js/           Javascript files
        images/       Image files

    tests/            Test suites


## Help

* The [Getting Started with Revel](http://revel.github.io/tutorial/gettingstarted.html).
* The [Revel guides](http://revel.github.io/manual/index.html).
* The [Revel sample apps](http://revel.github.io/examples/index.html).
* The [API documentation](https://godoc.org/github.com/revel/revel).

