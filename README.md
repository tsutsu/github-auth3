# github-auth3

`github-auth3` is a plugin for OpenBSD SSHD (specifically, an `AuthorizedKeysCommand`) which allows users to authenticate themselves to machines configured with it by supplying the usernames of their GitHub accounts, and then doing SSH pubkey auth against any public key attached to those GitHub accounts. Access is controlled by membership to a specified GitHub organization.

## Usage

1. Create a GitHub organization, or choose one you already have. *Any member of the configured organization will be able to log into the server.*

2. [generate a Github access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/), with minimal grants, against a user (usually yourself, but this could be an isolated "machine user") who can "see into" the membership of the organization. (For most organizations, all members are publicly visible, so you can do this as any user, even one who is not a member of the organization.)

3. Add the following to `/etc/ssh/sshd_config`:

```
AuthorizedKeysCommand /usr/local/bin/github-auth3 -a YOUR_GITHUB_ACCESS_TOKEN -o YOUR_ORG_NAME -u %u
```

4. Restart the `sshd` service (`sudo systemctl restart sshd` or equivalent.)

## Hardening

If you're worried about having an access token embedded in `/etc/ssh/sshd_config`, you can create a wrapper script like the following:

```bash
#!/bin/sh
exec /usr/local/bin/github-auth3 -a "$(cat /etc/ssh/github_access_token)" "$@"
```

And use an `sshd_config` line like this:

```
AuthorizedKeysCommand /usr/local/bin/your-wrapper-script -o YOUR_ORG_NAME -u %u
```

You can then create your `/etc/ssh/github_access_token` and `chmod 400 /etc/ssh/github_access_token` to protect it from prying eyes.
