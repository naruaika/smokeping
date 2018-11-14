Go Smokeping
Its reincarnation of Tobias Oetiker SmokePing daemon, written on Go.

Usage: go-smokeping -config CONFIGFILE.toml [-verbose]
CONFIGFILE.toml format:

[global_tags]
project = "myproject"                          # Project tag to add to records       
output ="influx"                               # output db format

[agent]
hostname = "myagent"                           # Agent tag to add to records

[database.influx] 
host = "https://influxdb.influxdb.test:8086"   # Connect string to influx DB
db = "ping"                                    # Database name
user = "ping"                                  # Db user
pass = "pong"                                  # Db password
step = 60                                      # Step to write to Db in seconds

[probe.icmp]                                   # Probe host with internal icmp mech
retries = 30                                   # Number of icmp packets
step = 60                                      # Step between probes in seconds
                                               # Host is pinged by retries packets(30), every step interval

[probe.fping]                                  # Probe host with external fping command
step = 60                                      # Step betwrrn probes in seconds
cmd = "/usr/bin/fping"                         # Path to binary
args = ["-c", "10", "-q"]                      # Additional arguments, see man fping to additional info

[[group]]                                      # Group of host
name = "google"                                # Group tag name to add to records
  [[group.host]]                               # Host definition
    fqdn = "ns1.google.com"                    # Host name
    ip = "8.8.8.8"                             # Ip address of host
    probe = "fping"                            # Probe for host

  [[group.host]]                               # Another definition
    fqdn = "ns2.google.com"            
    ip = "8.8.4.4"
    probe = "fping"

[[group]]                                     # Another group
name = "cloudflare"                           
  [[group.host]]                              # Another host definition
    fqdn = "ns1.cloudflare.com"            
    ip = "1.1.1.1"
    probe = "icmp"


Currently go-smokeping support only 2 probes: internal icmp and external fping. Output supported only to influxdb.
To run go-smokeping on systemd based systems put go-smokeping.service file to /etc/systemd/system and run:
systemctl daemon-reload
systemctl enable go-smokeping
systemctl start go-smokeping

For viewing stats in grafana import dashboard from file Smokeping.json
