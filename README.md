# cping

Tiny Go program that updates a CloudFlare DNS record with the local host's outward facing IP address. Useful for a dynamic DNS type setup.

Install with:

Create `~/.cping` with something like:

```
Email = <email>
Token = <API secret token>
```

Then run from the command line:

```
cping --daemonize
```
