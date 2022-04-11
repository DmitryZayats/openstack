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