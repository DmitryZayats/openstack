# openstack
Openstack client re-implementation in go.  

Tested on :  
- devstack Newton  
- microstack ussuri  
- RedHat OSP 16.1  

The goal is to gradually keep adding functionality that is not available in [github.com/gophercloud/gophercloud](https://github.com/gophercloud/gophercloud)

## Using library for adding security rules to a group with --remote-group

Below code is equivalent to following openstack cli commands  

```
openstack security group create --description "Group that allows communication within cluster" Internal_Cluster 
openstack security group create --description "Default NetAct security group" NetAct_default
openstack security group rule create InternalCluster --protocol udp --remote-group NetAct_default
openstack security group rule create InternalCluster --protocol udp --remote-group NetAct_default --ethertype IPv6
openstack security group rule create InternalCluster --protocol tcp --remote-group NetAct_default
openstack security group rule create InternalCluster --protocol tcp --remote-group NetAct_default --ethertype IPv6
openstack security group rule create InternalCluster --protocol icmp --remote-group NetAct_default
openstack security group rule create InternalCluster --protocol icmp --remote-group NetAct_default --ethertype IPv6
```

```
package main

import (
	"github.com/DmitryZayats/openstack"
)

func main() {
	openstack.Get_token()
	openstack.Create_security_group("Internal_Cluster", "Group that allows communication within cluster")
	openstack.Create_security_group("NetAct_default", "Default NetAct security group")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "tcp")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "udp")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv4", "icmp")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "tcp")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "udp")
	openstack.Add_remote_group_security_rule("Internal_Cluster", "NetAct_default", "IPv6", "icmp")
}
```

Result of running above code on a fresh install of microstack is shown below where we can see 6 ingress rules added   

```
(openstack) dmitry@dmitry-Standard-PC-Q35-ICH9-2009:~/Dev/Golang/TestOS$ openstack security group list --insecure
+--------------------------------------+------------------+------------------------------------------------+----------------------------------+------+
| ID                                   | Name             | Description                                    | Project                          | Tags |
+--------------------------------------+------------------+------------------------------------------------+----------------------------------+------+
| 6cdd45fe-2563-4bad-9570-caacd0153ec7 | NetAct_default   | Default NetAct security group                  | 4ddb973d4d564e4ebb68136bb81fc14f | []   |
| 7d4209e4-3a4b-4b67-ad11-219f84b19f60 | default          | Default security group                         | 4ddb973d4d564e4ebb68136bb81fc14f | []   |
| a773f92d-11b1-4c45-bed9-8c9d94f2042c | Internal_Cluster | Group that allows communication within cluster | 4ddb973d4d564e4ebb68136bb81fc14f | []   |
+--------------------------------------+------------------+------------------------------------------------+----------------------------------+------+

(openstack) dmitry@dmitry-Standard-PC-Q35-ICH9-2009:~/Dev/Golang/TestOS$ openstack security group rule list Internal_Cluster --insecure
+--------------------------------------+-------------+-----------+-----------+------------+-----------+--------------------------------------+----------------------+
| ID                                   | IP Protocol | Ethertype | IP Range  | Port Range | Direction | Remote Security Group                | Remote Address Group |
+--------------------------------------+-------------+-----------+-----------+------------+-----------+--------------------------------------+----------------------+
| 0d2842f7-718e-4336-852f-f96cc7b5a274 | ipv6-icmp   | IPv6      | ::/0      |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
| 1633a051-05b9-44ef-b9ae-18003c73ffab | tcp         | IPv6      | ::/0      |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
| 1ff2550d-7ccd-4845-8809-be93b4485a4c | None        | IPv4      | 0.0.0.0/0 |            | egress    | None                                 | None                 |
| 438a9a47-edbf-4ab8-8ba5-42c8cc93871d | icmp        | IPv4      | 0.0.0.0/0 |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
| 5d4fdb56-7db1-40c7-a1b5-d6ec29b0398f | None        | IPv6      | ::/0      |            | egress    | None                                 | None                 |
| afc0c7e0-2884-4a46-8739-2daa99d7a80e | udp         | IPv4      | 0.0.0.0/0 |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
| cc828180-3a02-4397-b5dd-e745e898195f | tcp         | IPv4      | 0.0.0.0/0 |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
| f16e5bfe-6b7c-42bd-94b4-363e199cff89 | udp         | IPv6      | ::/0      |            | ingress   | 6cdd45fe-2563-4bad-9570-caacd0153ec7 | None                 |
+--------------------------------------+-------------+-----------+-----------+------------+-----------+--------------------------------------+----------------------+

```