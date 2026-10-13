# RDS Proxy to DB Instance

Opens the network path from an RDS proxy to the DB instance it fronts, then registers
that instance as the proxy's target.

Both halves are done here because the order between them matters. A target registered
before the path is open reports `UNAVAILABLE`, and the proxy serves nothing until it can
reach its own target.

Egress is authorised on the proxy's security group towards the instance and ingress on
the instance's group from the proxy, on the database port. Sharing a group is not enough as security groups do not permit traffic within a group by default, so a proxy and instance
in the same group still need a rule naming that group on both sides.

## Why there is no target group resource

`AWS::RDS::DBProxyTargetGroup` blocks its create until the registered target reports
healthy. Modelled as a resource it would have been created a whole phase before the link that writes
the rule making the target healthy, and spent its entire timeout waiting. RDS creates the
default target group along with the proxy, so nothing needs to create it.

## Port

Read from the instance's own port. Override it with the
`aws.rds.dbProxy.<targetInstance>.port` annotation on the proxy.

## Removal

Security group rules are tagged with the link that wrote them and destroying the link
revokes exactly those rules. The target is deregistered at the same time; a target that
is already gone is not treated as a failure, since the instance is commonly destroyed
before the link when a whole blueprint comes down.
