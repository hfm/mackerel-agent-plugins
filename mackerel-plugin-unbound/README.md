mackerel-plugin-unbound
=====================

Unbound custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-unbound [-command=<command> ][-ip=<ipaddr>] [-port=<port>] [-conf=<config file>] [-enable_extended] [-metric-key-prefix=<prefix>]
```

## Example of mackerel-agent.conf

```
[plugin.metrics.unbound]
command = "/path/to/mackerel-plugin-unbound -conf=/path/to/unbound.conf -enable_extended"
```

### Enable extended statistics.

If you run on an ec2-instance and the instance is associated with an appropriate IAM Role, you probably don't have to specify `-access-key-id` & `-secret-access-key` in unbound.conf

```yaml
server:
    statistics-interval: 0
    extended-statistics: yes
    statistics-cumulative: no

remote-control:
    control-enable: yes
```

## References

- https://www.unbound.net/documentation/howto_statistics.html
