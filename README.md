# END OF LIFE NOTICE

> **Flexpool.io will officially wind down its operations on November 1, 2023**

[See the full
announcement](https://www.reddit.com/r/Flexpool/comments/16q72ul/action_required_flexpoolio_shutdown_notice_nov_1/).

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

### Telegram

Follow [this procedure](https://core.telegram.org/bots#3-how-do-i-create-a-bot) to create a bot `token`.

Then you have two possible destinations to send messages:
* channel using a `channel_name` (string)
* chat using a `chat_id` (integer)

For testing purpose, you should store the token in a variable for next sections:
```
read -s TOKEN
```

#### Chat

To get the chat identifier, you can send a message to your bot then read messages using the API:

```
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/getUpdates" | jq -r ".result[].message.chat.id"
```

You can test to send messages to a chat with:

```
read CHAT_ID
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/sendMessage?chat_id=${CHAT_ID}&text=hello" | jq
```

#### Channel

Public channel names can be used (example: `@mychannel`). For private channels, you should use a `chat_id` instead.

You can test to send messages to a channel with:

```
read CHANNEL_NAME
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/sendMessage?chat_id=${CHANNEL_NAME}&text=hello" | jq
```

Don't forget to prefix the channel name with an `@`.


### flexassistant

*flexassistant* can be configured using a YaML file. By default, the `flexassistant.yaml` file is used but it can be
another file provided by the `-config` argument.

As a good start, you can copy the configuration file example:

```
cp -p flexassistant.example.yaml flexassistant.yaml
```

Then edit this file at will.

Reference:
* `database-file` (optional): file name of the database file to persist information between two executions (SQLite
   database)
* `max-blocks` (optional): maximum number of blocks to retreive from the API
* `max-payments` (optional): maximum number of payments to retreive from the API
* `pools` (optional): list of pools
    * `coin`: coin of the pool (ex: `etc`, `eth`, `xch`)
    * `enable-blocks` (optional): enable block notifications for this pool (disabled by default)
    * `min-block-reward` (optional): send notifications when block reward has reached this minimum threshold in crypto
       currency unit (ETH, XCH, etc)
* `miners` (optional): list of miners and/or farmers
    * `address`: address of the miner or the farmer registered on the API
    * `coin` (optional): coin of the miner (ex: `etc`, `eth`, `xch`) (deduced by default, can be wrong for `etc` coin)
    * `enable-balance` (optional): enable balance notifications (disabled by default)
    * `enable-payments` (optional): enable payments notifications (disabled by default)
    * `enable-offline-workers` (optional): enable offline/online notifications for associated workers (disabled by
       default)
* `telegram`: Telegram configuration
    * `token`: token of the Telegram bot
    * `chat-id` (optional if `channel-name` is present): chat identifier to send Telegram notifications
    * `channel-name` (optional if `chat-id` is present): channel name to send Telegram notifications
* `notifications` (optional): Notifications configurations
    * `balance` (optional): balance notifications settings
        * `template` (optional): path to [template](https://pkg.go.dev/text/template) file
        * `test` (optional): send a test notification
    * `payment` (optional): payment notifications settings
        * `template` (optional): path to [template](https://pkg.go.dev/text/template) file
        * `test` (optional): send a test notification
    * `block` (optional): block notification settings
        * `template` (optional): path to [template](https://pkg.go.dev/text/template) file
        * `test` (optional): send a test notification
    * `offline-worker` (optional): offline workers notification settings
        * `template` (optional): path to [template](https://pkg.go.dev/text/template) file
        * `test` (optional): send a test notification

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

Custom template files can be used with the `template` settings (see _Configuration_ section).

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
