# cping

Tiny Go program that updates a CloudFlare DNS record with the local host's outward facing IP address. Useful for a dynamic DNS type setup.

Install with:

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
