# chaos-galago-smoke-tests

### Overview

`chaos-galago-smoke-tests` is a project to enable smoke tests for `https://github.com/FidelityInternational/chaos-galago`.

### Useage

```
git clone https://github.com/FidelityInternational/chaos-galago-smoke-tests.git
cd chaos-galago-smoke-tests
CF_USERNAME='an_admin_user' \
CF_PASSWORD='an_admin_password' \
CF_DOMAIN='system_domain.example.com' \
CF_HOME='temp_dir' \
go test -v ./...
```
