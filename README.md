# flexassistant

[Flexpool.io](https://www.flexpool.io/) is a famous cryptocurrency mining or farming pool supporting
[Ethereum](https://ethereum.org/en/) and [Chia](https://www.chia.net/) blockchains. As a miner, or a farmer, we like to
get **notified** when a **block** is mined, or farmed. We also like to keep track of our **unpaid balance** and our
**transactions** to our personal wallet.

*flexassistant* is a tool that parses the Flexpool API and sends notifications via [Telegram](https://telegram.org/).

<p align="center">
    <img src="static/screenshot.jpg" width="300" />
</p>


## Installation

*Note: this guide has been written with Linux x86_64 in mind.*

### Binaries

Go to [Releases](https://github.com/jouir/flexassistant/releases) to download the binary in the version you like (latest
is recommended) into a `bin` directory.

Write checksum information to a local file:

```
echo checksum > flexassistant-VERSION-Linux-x86_64.sha256sum
```

Verify checksums to avoid binary corruption:

```
sha256sum -c flexassistant-VERSION-Linux-x86_64.sha256sum
```

### Compilation

You will need to install [Go](https://golang.org/dl/), [Git](https://git-scm.com/) and a development toolkit (including
[make](https://linux.die.net/man/1/make)) for your environment.

Then, you'll need to download and compile the source code:

```
git clone https://github.com/jouir/flexassistant.git
cd flexassistant
make
```

The binary will be available under the `bin` directory:

```
ls -l bin/flexassistant
```

## Configuration

*flexassistant* can be configured using a YaML file. By default, the `flexassistant.yaml` file is used but it can be
another file provided by the `-config` argument.

As a good start, you can copy the configuration file example:

```
cp -p flexassistant.yaml.example flexassistant.yaml
```

Then edit this file at will.

Reference:
* `database-file` (optional): file name of the database file to persist information between two executions (SQLite
   database)
* `max-blocks` (optional): maximum number of blocks to retreive from the API
* `max-payments` (optional): maximum number of payments to retreive from the API
* `pools` (optional): list of pools
    * `coin`: coin of the pool (ex: `eth`, `xch`)
    * `enable-blocks` (optional): enable block notifications for this pool (disabled by default)
* `miners` (optional): list of miners and/or farmers
    * `address`: address of the miner or the farmer registered on the API
    * `enable-balance` (optional): enable balance notifications (disabled by default)
    * `enable-payments` (optional): enable payments notifications (disabled by default)
    * `enable-offline-workers` (optional): enable offline/online notifications for associated workers (disabled by
       default)
* `telegram`: Telegram configuration
    * `token`: token of the Telegram bot
    * `chat-id` (optional if `channel-name` is present): chat identifier to send Telegram notifications
    * `channel-name` (optional if `chat-id` is present): channel name to send Telegram notifications
* `notification-templates` (optional): path to [template](https://pkg.go.dev/text/template) files for each notification
   type
    * `balance` (optional): path to template file to format balance notifications
    * `payment` (optional): path to template file to format payment notifications
    * `block` (optional): path to template file to format block notifications
    * `offline-worker` (optional): path to template file to format offline worker notifications

## Templating

Notifications can be customized with [templating](https://pkg.go.dev/text/template).

The following **functions** are available to templates:
* `upper(str string)`: convert string to upper case
* `lower(str string)`: convert string to lower case
* `convertCurrency(coin string, value int64)`: convert the smallest unit of a coin to a human readable unit
* `formatBlockURL(coin string, hash string)`: return the URL on the explorer website of the coin of the block
   identified by its hash
* `formatTransactionURL(coin string, hash string)`: return the URL on the explorer website of the coin of the
   transaction identified by its hash

The following **data** is available to templates:
* balance: `.Miner`
* payment: `.Miner`, `.Payment`
* block: `.Pool`, `.Block`
* offline-worker: `.Worker`

Default templates are available in the [templates](templates) directory.

Custom template files can be used with the `notification-templates` settings (see _Configuration_ section).

## Usage

```
Usage of ./flexassistant:
  -config string
        Configuration file name (default "flexassistant.yaml")
  -debug
        Print even more logs
  -quiet
        Log errors only
  -verbose
        Print more logs
  -version
        Print version and exit
```
