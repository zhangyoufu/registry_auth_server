# registry_auth_server

Auth server for Docker registry.

Use htpasswd-format text file (username:bcrypt_hash) for authentication.
Write `github.com/ory/ladon` [policies](https://pkg.go.dev/github.com/ory/ladon#readme-policies) in YAML format for access control.

## Example Policies

```yaml
id: allow-user-pull-alpine
effect: allow
subjects:
- user
actions:
- pull
resources:
- repository:alpine
```

```yaml
id: allow-admin-do-anything
effect: allow
subjects:
- admin
actions:
- <.*>
resources:
- <.*>
```

```yaml
id: allow-trusted-ip-do-anything
effect: allow
subjects:
- <.*>
actions:
- <.*>
resources:
- <.*>
conditions:
  ip:
    type: CIDRCondition
    options:
      cidr: 192.0.2.0/24
```
