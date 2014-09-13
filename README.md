# cping

Tiny Go program that updates a CloudFlare DNS record with the local host's outward facing IP address. Useful for a dynamic DNS type setup.

Install with:

```
# Linux
$ curl -L http://gobuild.io/github.com/brandur/cping/master/linux/amd64 -o cping.zip
$ unzip cping.zip cping -d /usr/local/bin

# OSX
$ curl -L http://gobuild.io/github.com/brandur/cping/master/darwin/amd64 -o cping.zip
$ unzip cping.zip cping -d /usr/local/bin

# or if you have go
go get -u github.com/brandur/cping
```

Create `~/.cping` with something like:

```
[cloudflare]
email = <email>
name = home.example.com
token = <API secret token>
zone = example.com
```

Then run from the command line:

```
cping
```
