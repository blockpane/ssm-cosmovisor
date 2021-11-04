# ssm-cosmovisor

**You probably don't want to use this and do so at your own risk.**

This is *very* experimental and completely untested. It will likely: set your house on fire, kick your dog, and then make fun of your children ... right before getting you slashed for equivocation. 

This wraps cosmovisor to offer slightly more secure handling of a consensus private key and to make warm-failover easier.

ssm-cosmovisor is an experiment that uses encrypted parameters in AWS's SSM Parameter Store to hold a 
tendermint private key and to provide the key to tendermint upon startup via a named-pipe instead of 
persisting the private key on the filesystem. It is expected that a "safe" priv_validator_key.json file 
exists on the system that is **not** the consensus key. This allows controlling whether the node is an 
active validator via a single environment variable.

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
