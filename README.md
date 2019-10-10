# github-auth3

`github-auth3` is a plugin for OpenBSD SSHD (specifically, an `AuthorizedKeysCommand`) which allows users to authenticate themselves to machines configured with it by supplying the usernames of their GitHub accounts, and then doing SSH pubkey auth against any public key attached to those GitHub accounts. Access is controlled by membership to a specified GitHub organization.

## Usage

1. Create a GitHub organization, or choose one you already have. *Any member of the configured organization will be able to log into the server.*

2. [generate a Github access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/), with minimal grants, against a user (usually yourself, but this could be an isolated "machine user") who can "see into" the membership of the organization. (For most organizations, all members are publicly visible, so you can do this as any user, even one who is not a member of the organization.)

3. Add the following to `/etc/ssh/sshd_config`:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -a YOUR_GITHUB_ACCESS_TOKEN -o YOUR_ORG_NAME -u %u
AuthorizedKeysCommandUser sshauthcmd
```

4. Create the `sshauthcmd` user, and a credential-cache directory for `github-auth3` to use:

```bash
#!/bin/sh
sudo adduser --system sshauthcmd
sudo mkdir -p '/var/cache/github-auth3'
sudo chown sshauthcmd:root '/var/cache/github-auth3'
sudo chmod 0700 '/var/cache/github-auth3'
```

4. Restart the `sshd` service (`sudo systemctl restart sshd` or equivalent.)

## Hardening

If you're worried about having an access token embedded in `/etc/ssh/sshd_config` (despite this token not really being able to do anything much), you can provide a path instead:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -apath /etc/ssh/github_access_token -o YOUR_ORG_NAME -u %u
```

You'll probably want to lock down access to the token file itself, but remember that it's `github-auth3`, not OpenSSH itself, that will need to access this file. The file will need to be owned by the `AuthorizedKeysCommandUser` (`sshauthcmd` above):

```bash
#!/bin/sh
sudo touch /etc/ssh/github_access_token
sudo chown sshauthcmd:root /etc/ssh/github_access_token
sudo chmod 0400 /etc/ssh/github_access_token
```
