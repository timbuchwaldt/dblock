#dblock
dblock is an ipset based fail2ban alternative with cluster support. It follows logs, parses them using regex and blocks attackers based on the rate of incidents. It supports IPv4 + IPv6, stores blocks in etcd for all other hosts on your network to also block the IPs. It supports whitelisting your own networks to prevent accidental blocking.
