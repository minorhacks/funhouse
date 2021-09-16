# funhouse_server

## Print Handler

The `print` handler simply prints out the JSON request, for debugging and
testcase-capturing purposes. For a server running at
`funhouse.minorhacks.cloud`, issue a request like:

```
xh \
  --follow \
  --verify=no \
  POST \
  https://funhouse.minorhacks.cloud/hook/print \
  ref=foo \
  before=bar \
  after=baz
```

and then look for the request payload in the logs:

```
{Ref:        "foo",
 Before:     "bar",
 After:      "baz",
 Repository: nil}
```

## Mirror Handler

The `mirror` handler clones the named repository if it is not cloned already,
and pulls the latest on the specified branch.

```
xh \
  --follow \
  --verify=no \
  POST \
  https://funhouse.minorhacks.cloud/hook/mirror \
  ref=master \
  repository:='{"url": "https://github.com/minorhacks/advent_2020"}'
```