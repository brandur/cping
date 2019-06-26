# cping [![Build Status](https://travis-ci.org/brandur/cping.svg?branch=master)](https://travis-ci.org/brandur/cping)

Tiny Go program that updates a CloudFlare DNS record with the local host's outward facing IP address. Useful for a dynamic DNS type setup.

Install with:

``` sh
go get -u github.com/brandur/cping
cp $GOPATH/bin/cping /usr/local/bin/
```

Create `~/.cping` with something like:

```
[cloudflare]
email = <email>
name = site.example.com
token = <API secret token>
zone = example.com
```

Then run from the command line:

``` sh
$ cping
```

Install in your Crontab with something like:

```
*/5 * * * * /usr/local/bin/cping -v 2>&1 | tee -a ~/.cping.log
```
