![logo](logo.png "logo")

Chaos Monkey randomly terminates instances. This helps  ensure that engineers
implement their services to be resilient to instance failures.

## Prerequisites

* [Spinnaker]
* MySQL

To use Chaos Monkey, you must be using [Spinnaker]. Spinnaker is the
continuous delivery platform that we use at Netflix.

Chaos Monkey also requires a MySQL database.

[Spinnaker]: http://www.spinnaker.io/


## Build

To build Chaos Monkey on your local machine (requires the Go
toolchain).

```
go install github.com/netflix/chaosmonkey/bin/chaosmonkey
```

This will install a `chaosmonkey` binary in your `$GOBIN` directory.

## How Chaos Monkey runs

Chaos Monkey does not run as a service. Instead, you set up a cron job
that calls Chaos Monkey once a weekday to create a schedule of terminations.

When Chaos Monkey creates a schedule, it creates another cron job to schedule terminations
during the working hours of the day.

## Deploy overview

To deploy Chaos Monkey, you need to:

1. Configure Spinnaker for Chaos Monkey support
1. Set up the MySQL database
1. Write a configuration file (chaosmonkey.toml)
1. Set up a cron job that runs Chaos Monkey daily schedule

## Configure Spinnaker for Chaos Monkey support

Spinnaker's web interface is called *Deck*. You need to be running Deck version
v.2839.0 or greater for Chaos Monkey support. Check which version of Deck you are
running by hitting the `/version.json` endpoint of your Spinnaker deployment.
(Note that this version information will not be present if you are running
Deck using a [Docker container hosted on Quay][quay]).

[quay]: https://quay.io/repository/spinnaker/deck

Deck has a config file named `/var/www/settings.js`. In this file there is a
"feature" object that contains a number of feature flags:

```
  feature: {
    pipelines: true,
    notifications: false,
    fastProperty: true,
    ...
```

Add the following flag:

```
chaosMonkey: true
```

If the feature was enabled successfully, when you create a new app with Spinnaker, you will see
a "Chaos Monkey: Enabled" checkbox in the "New Application" modal dialog. If it
does not appear, you may need to deploy a more recent version of Spinnaker.

![new-app](new-app.png "new application dialog")

For more details, see [Additional configuration files][spinconfig] on the
Spinnaker website.

[spinconfig]: http://www.spinnaker.io/docs/custom-configuration#section-additional-configuration-files



## Set up the MySQL database

Chaos Monkey uses a MySQL database as a backend to record daily termination
schedule and to enforce a minimum time between terminations. (By default, Chaos
Monkey will not terminate more than one instance per day per group).

Set up a MySQL database named chaosmonkey with the following schema:

```
CREATE TABLE IF NOT EXISTS schedules (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    entry TEXT NOT NULL,
    INDEX date_index (date)
)jj
ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS schedules (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    date         DATE NOT NULL,        -- date of termination schedule, in local time zone
    time         DATETIME NOT NULL,    -- time in UTC. Because of time difference, may differ from date
    app          VARCHAR(512) NOT NULL,
    account      VARCHAR(100) NOT NULL,
    region       VARCHAR(50)  NOT NULL, -- use blank string to indicate not present
    stack        VARCHAR(255) NOT NULL, -- use blank string to indicate not present
    cluster      VARCHAR(768) NOT NULL, -- use blank string to indicate not present
    INDEX date_index (date)
    )
ENGINE=InnoDB;
```

Note: Chaos Monkey does not currently include a mechanism for purging old data.
Until this function exists, it is the operator's responsibility to remove old
data as needed.

## Write a configuration file (chaosmonkey.toml)


### Config file location

Chaos Monkey will look for a file named `chaosmonkey.toml` in the following
locations:

 * `.` (current directory)
 * `/apps/chaosmonkey`
 * `/etc`
 * `/etc/chaosmonkey`


### Configuration file format

The config file is in [TOML] format. Here is an example configuration file:

[TOML]: https://github.com/toml-lang/toml


```
[chaosmonkey]
enabled = true
schedule_enabled = true
leashed = false
accounts = ["production", "test"]

[database]
host = "dbhost.example.com"
name = "chaosmonkey"
user = "chaosmonkey"
encrypted_password = "securepasswordgoeshere"

[spinnaker]
endpoint = "http://spinnaker.example.com:8084"
```

Note that while the field is called "encrypted_password", you should put the
unencrypted version of your password here. Chaos Monkey currently only ships
with a no-op (do nothing) password decryptor.


The following example shows all of the default values:

```
[chaosmonkey]
enabled = false                    # if false, won't terminate instances when invoked
leashed = true                     # if true, terminations are only simulated (logged only)
schedule_enabled = false           # if true, will generate schedule of terminations each weekday
accounts = []                      # list of Spinnaker accounts with chaos monkey enabled, e.g.: ["prod", "test"]

start_hour = 9                     # time during day when starts terminating
end_hour = 15                      # time during day when stops terminating

time_zone = "America/Los_Angeles"  # time zone used by start.hour and end.hour

term_account = "root"              # account used to run the term_path command

max_apps = 2147483647              # max number of apps Chaos Monkey will schedule terminations for

# location of command Chaos Monkey uses for doing terminations
term_path = "/apps/chaosmonkey/chaosmonkey-terminate.sh"

# cron file that Chaos Monkey writes to each day for scheduling kills
cron_path = "/etc/cron.d/chaosmonkey-daily-terminations"

# decryption system for encrypted_password fields for spinnaker and database
decryptor = ""

# event tracking systems that records chaos monkey terminations
trackers = []

# metric collection systems that track errors for monitoring/alerting
error_counter = ""

# outage checking system that tells chaos monkey if there is an ongoing outage
outage_checker = ""

[database]
host = ""                # database host
port = 3306              # tcp port that the database is lstening on
user = ""                # database user
encrypted_password = ""  # password for database auth, encrypted by decryptor
name = ""                # name of database that contains chaos monkey data

[spinnaker]
endpoint = ""           # spinnaker api url
certificate = ""        # path to p12 file when using client-side tls certs
encrypted_password = "" # password used for p12 certificate, encrypted by decryptor
user = ""               # user associated with terminations, sent in API call to terminate

# For dynamic configuration options, see viper docs
[dynamic]
provider = ""   # options: "etcd", "consul"
endpoint = ""   # url for dynamic provider
path = ""       # path for dynamic provider
```

Note that many of these configuration parameters (decryptor, trackers,
error_counter, outage_checker) currently only have no-op implementations.

### Verifying Chaos Monkey is configured properly

Chaos Monkey supports a number of command-line arguments that are useful for
verifying that things are working properly.

#### Spinnaker

You can verify that Chaos Monkey reach Spinnaker by fetching the Chaos Monkey
configuration for an app:

```
chaosmonkey config <appname>
```

If successful, you'll see output that looks like:

```
(*chaosmonkey.AppConfig)(0xc4202ec0c0)({
 Enabled: (bool) true,
 RegionsAreIndependent: (bool) true,
 MeanTimeBetweenKillsInWorkDays: (int) 2,
 MinTimeBetweenKillsInWorkDays: (int) 1,
 Grouping: (chaosmonkey.Group) cluster,
 Exceptions: ([]chaosmonkey.Exception) {
 }
})
```

If it fails, you'll see an error message.

#### Database

You can verify that Chaos Monkey can reach the database by attempting to
retrieve the termination schedule for the day.

```
chaosmonkey fetch-schedule
```

If successful, you should see output like:

```
[69400] 2016/09/30 23:41:03 chaosmonkey fetch-schedule starting
[69400] 2016/09/30 23:41:03 Writing /etc/cron.d/chaosmonkey-daily-terminations
[69400] 2016/09/30 23:41:03 chaosmonkey fetch-schedule done
```

(Chaos Monkey will write an empty file to
`/etc/cron.d/chaosmonkey-daily-terminations` since the database does not contain
any termination schedules yet).

If Chaos Monkey cannot reach the database, you will see an error. For example:

```
[69668] 2016/09/30 23:43:50 chaosmonkey fetch-schedule starting
[69668] 2016/09/30 23:43:50 FATAL: could not fetch schedule: failed to retrieve schedule for 2016-09-30 23:43:50.953795019 -0700 PDT: dial tcp 127.0.0.1:3306: getsockopt: connection refused
```



### Optional: Dynamic properties (etcd, consul)

Chaos Monkey supports changing the following configuration properties dynamically:

* chaosmonkey.enabled
* chaosmonkey.leashed
* chaosmonkey.schedule_enabled
* chaosmonkey.accounts

These are intended to allow an operator to make certain changes to Chaos
Monkey's behavior without having to redeploy.

Note: the configuration file takes precedence over dynamic provider, so do
not specify these properties in the config file if you want to set them
dynamically.

To take advantage of dynamic properties, you need to keep those properties in
either [etcd] or [Consul] and add a `[dynamic]` section that contains the
endpoint for the service and a path that returns a JSON file that has each of
the properties you want to set dynamically.

Chaos Monkey uses the [Viper][viper] library to implement dynamic configuration, see the
Viper [remote key/value store support][remote] docs for more details.


[etcd]: https://coreos.com/etcd/docs/latest/
[consul]: https://www.consul.io/
[viper]: https://github.com/spf13/viper
[remote]: https://github.com/spf13/viper#remote-keyvalue-store-support


## Set up a cron job that runs Chaos Monkey daily schedule

### Create /apps/chaosmonkey/chaosmonkey-schedule.sh

For the remainder if the docs, we assume you have copied the chaosmonkey binary
to `/apps/chaosmonkey`, and will create the scripts described below there as
well. However, Chaos Monkey makes no explicit assumptions about the location of
these files.


Create a file called `chaosmonkey-schedule.sh` that invokes `chaosmonkey
schedule` and writes the output to a logfile.

Note that because this will be invoked from cron, the PATH will likely not include the
location of the chaosmonkey binary so be sure to specify it explicitly.

/apps/chaosmonkey/chaosmonkey-schedule.sh:
```bash
#!/bin/bash
/apps/chaosmonkey/chaosmonkey schedule >> /var/log/chosmonkey-schedule.log 2>&1
```

### Create /etc/cron.d/chaosmonkey-schedule


Once you have this script, create a cron job that invokes it once a day. Chaos
Monkey starts terminating at `chaosmonkey.start_hour` in
`chaosmonkey.time_zone`, so it's best to pick a time earlier in the day.

The example below generates termination schedules each weekday at 12:00 system
time (which we assume is in UTC).

/etc/cron.d/chaosmonkey-schedule:
```bash
# Run the Chaos Monkey scheduler at 5AM PDT (4AM PST) every weekday
# This corresponds to: 12:00 UTC
# Because system clock runs UTC, time change affects when job runs

# The scheduler must run as root because it needs root permissions to write
# to the file /etc/cron.d/chaosmonkey-daily-terminations

# min  hour  dom  month  day  user  command
    0    12    *      *  1-5  root  /apps/chaosmonkey/chaosmonkey-schedule.sh
```

### Create /apps/chaosmonkey/chaosmonkey-terminate.sh

When Chaos Monkey schedules terminations, it will create cron jobs that call the
path specified by `chaosmonkey.term_path`, which defaults to /apps/chaosmonkey/chaosmonkey-terminate.sh

/apps/chaosmonkey/chaosmonkey-terminate.sh:
```
#!/bin/bash
/apps/chaosmonkey/chaosmonkey terminate "$@" >> /var/log/chaosmonkey-terminate.log 2>&1
```


## Configuring Chaos Monkey behavior via Spinnaker

Through the Spinnaker web UI, you can configure how often Chaos Monkey
terminates instances for each application.

Click on the "Config" tab in Spinnaker. There should be a "Chaos Monkey"
widget where you can enable/disable Chaos Monkey for the app, as well as
configure its behavior.

![config](config.png "config dialog")

### Termination frequency

By default, Chaos Monkey is configured for a *mean time between terminations* of
two (2) days, which means that on average Chaos Monkey will terminate an
instance every two days for each group in that app.

Chaos Monkey also has a *minimum time between terminations*, which defaults to
one (1) day. This means that Chaos Monkey is guaranteed to never kill more often
than once a day for each group. Even if multiple Chaos Monkeys are deployed, as
long as they are all configured to use the same database, they will obey the
minimum time between terminations.

### Grouping

Chaos Monkey operates on *groups* of instances. Every work day, for every
(enabled) group of instances, Chaos Monkey will flip a biased coin to determine
whether it should kill from an instance from a group. If so, it will randomly
select an instance from the group.

Users can configure what Chaos Monkey considers a group.  The three options are:

* app
* stack
* cluster

If grouping is set to "app", Chaos Monkey will terminate up to one instance per
app each day, regardless of how these instances are organized into clusters.

If the grouping is set to "stack", Chaos Monkey will terminate up to one instance per
stack each day. For instance, if an application has three stacks defined, then
Chaos Monkey may kill up to three instances in this app per day.

If the grouping is set to "cluster", Chaos Monkey will terminate up to one
instance per cluster each day.

By default, Chaos Monkey treats each region separately. However, if the "regions
are independent" option is unchecked, then Chaos Monkey will not terminate
instances that are in the same group but in different regions. This is intended
to support databases that replicate across regions where simultaneous
termination across regions is undesirable.

### Exceptions

You can opt-out combinations of account, region, stack, and detail. In the
example config shown above, Chaos Monkey will not terminate instances in the
prod account in the us-west-2 region with a stack of "staging" and a blank
detail field.

The exception field also supports a wildcard, `*`, which matches everything. IN
the example above, Chaos Monkey will also not terminate any instances in the
test account, regardless of region, stack or detail.
