#dblock
dblock is an ipset based fail2ban alternative with cluster support. It follows logs, parses them using regex and blocks attackers based on the rate of incidents. It supports IPv4 + IPv6, stores blocks in etcd for all other hosts on your network to also block the IPs. It supports whitelisting your own networks to prevent accidental blocking.


## Architecture
* A goroutine is started for each log file. It tails the log and analyzes the lines
* The inicdentstore package reads all incidents picked up by the log followers and counts them
* The sync package reads the block requests issued by the incidentstore and writes them to etcd
* The sync package also watches the etcd directories containing the block messages, forwarding all changes to the blocker
* The blocker blocks and unblocks IPs based on the sync engines input

