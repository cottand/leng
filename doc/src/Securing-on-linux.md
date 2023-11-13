# Securing on Linux

The recommended way to harden access to the leng server is to only allow connections from clients you trust, mainly because public dns servers are hit by penetration testers and hackers regularly to scout for vulnerabilities.

## Installing Requirements

Let's grab ufw to allow for easy editing of iptables.

```
apt-get install ufw -y
```

## Firewall Setup

Now let's whitelist our dns clients IP address or range, and block access from everywhere else by default using ufw.

```
ufw deny 53
ufw allow from <ip or range> to any port 53
ufw reload
```

Now only the client(s) you whitelisted can access the dns server.