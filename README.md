# ssm-cosmovisor

**You probably don't want to use this and do so at your own risk.**

This is *very* experimental and completely untested. It will likely: set your house on fire, kick your dog, and then make fun of your children ... right before getting you slashed for equivocation. 

This wraps cosmovisor to offer slightly more secure handling of a consensus private key and to make warm-failover easier.

ssm-cosmovisor is an experiment that uses encrypted parameters in AWS's SSM Parameter Store to hold a 
tendermint private key and to provide the key to tendermint upon startup via a named-pipe instead of 
persisting the private key on the filesystem. It is expected that a "safe" priv_validator_key.json file 
already exists on the system that is **not** the consensus key. This allows controlling whether the node is an 
active validator via a single environment variable, and when un-set it returns to normal operation.

It does not handle any authentication to the AWS API, since it uses the standard AWS go library, all the
standard methods (including metadata v2 roles) should just work.

Additional ENV vars expected for this to work:

```shell
# parameters for getting the key
AWS_REGION=your-region-here
AWS_PARAMETER=/some/parameter
USE_SSM_KEY=true-or-false # if false, it will not try to get the consensus key from SSM.

# still need the standard cosmovisor vars too....
DAEMON_HOME=....
DAEMON_NAME=....
DAEMON_ALLOW_DOWNLOAD_BINARIES=false
DAEMON_RESTART_AFTER_UPGRADE=true
DAEMON_LOG_BUFFER_SIZE=512
UNSAFE_SKIP_BACKUP=true
```

Example of log output:

```
Nov 05 00:58:07 node systemd[1]: Started Xxxxx Daemon.
Nov 05 00:58:07 node ssm-cosmovisor[172898]: ssm.go:27: attempting to fetch consensus key from AWS parameter store
Nov 05 00:58:07 node ssm-cosmovisor[172898]: ssm.go:56: retrieved public key: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx from parameter store
Nov 05 00:58:07 node ssm-cosmovisor[172898]: file.go:88: backing up original key
Nov 05 00:58:07 node ssm-cosmovisor[172898]: file.go:142: original key saved
Nov 05 00:58:07 node ssm-cosmovisor[172898]: 12:58AM INF Configuration is valid:
...
Nov 05 00:58:07 node ssm-cosmovisor[172898]: 12:58AM INF running app args=["start"] module=cosmovisor path=/var/lib/xxxxx/.xxxxx/cosmovisor/genesis/bin/xxxxx
Nov 05 00:58:19 node ssm-cosmovisor[172898]: file.go:43: writing key to named pipe
Nov 05 00:58:19 node ssm-cosmovisor[172898]: file.go:49: key was read from named pipe
Nov 05 00:58:19 node ssm-cosmovisor[172898]: file.go:82: writing stripped key file
...
Nov 05 01:03:10 node systemd[1]: Stopping Xxxxx Daemon...
Nov 05 01:03:10 node ssm-cosmovisor[172898]: file.go:150: attempting to restore original key file
Nov 05 01:03:10 node ssm-cosmovisor[172898]: file.go:170: restored original key file
Nov 05 01:03:10 node systemd[1]: xxxxxx.service: Succeeded.
Nov 05 01:03:10 node systemd[1]: Stopped Xxxxx Daemon.
...
```
