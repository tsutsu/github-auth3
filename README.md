# github-auth3

`github-auth3` is a plugin for OpenBSD SSHD (specifically, an `AuthorizedKeysCommand`) which allows users to authenticate themselves to machines configured with it by supplying the usernames of their GitHub accounts, and then doing SSH pubkey auth against any public key attached to those GitHub accounts. Access is controlled by membership to a specified GitHub organization.

## Installation

1. Create a GitHub organization, or choose one you already have. *Any member of the configured organization will be able to log into the server.*

2. [Generate a GitHub access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/), with minimal grants, against a user (usually yourself, but this could be an isolated "machine user") who can "see into" the membership of the organization. (For most organizations, all members are publicly visible, so you can do this as any user, even one who is not a member of the organization.)

3. Create an `sshauthcmd` user:

```bash
#!/bin/sh
sudo adduser --system sshauthcmd
```

4. Add the following to your `/etc/ssh/sshd_config`:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -a YOUR_GITHUB_ACCESS_TOKEN -o YOUR_ORG_NAME -u %u
AuthorizedKeysCommandUser sshauthcmd
```

5. Restart the `sshd` service (`sudo systemctl restart sshd` or equivalent.)


## Enabling caching

`github-auth3` will optionally cache API responses according to the HTTP caching headers in GitHub's API responses. This doesn't matter much normally, but can avoid some pain if your instance is public-visible and gets DDoSed.

1. Create a credential-cache directory for `github-auth3` to use:

```bash
#!/bin/sh
sudo mkdir -p '/var/cache/github-auth3'
sudo chown sshauthcmd:root '/var/cache/github-auth3'
sudo chmod 0700 '/var/cache/github-auth3'
```

2. Add the `-cpath` flag to your `AuthorizedKeysCommand` in `/etc/ssh/sshd_config`:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -a YOUR_GITHUB_ACCESS_TOKEN -cpath /var/cache/github-auth3 -o YOUR_ORG_NAME -u %u
```

3. Restart the `sshd` service (`sudo systemctl restart sshd` or equivalent.)


## Hardening

If you're worried about having an access token embedded in `/etc/ssh/sshd_config` (despite this token not really being able to do anything much), you can provide a path to a token file instead.

You'll probably want to lock down access to the token file itself, but remember that it's the `AuthorizedKeysCommandUser`, not OpenSSH itself, that will need to access this file.

Steps:

1. Create the token file, and secure it:

```bash
#!/bin/sh
sudo vi /etc/ssh/github_access_token # or what-have-you
sudo chown sshauthcmd:root /etc/ssh/github_access_token
sudo chmod 0400 /etc/ssh/github_access_token
```

2. In `/etc/ssh/sshd_config`, replace your `AuthorizedKeysCommand`'s `-a` flag with `-apath`:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -apath /etc/ssh/github_access_token -o YOUR_ORG_NAME -u %u
```

3. Restart the `sshd` service (`sudo systemctl restart sshd` or equivalent.)
