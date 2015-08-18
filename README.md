# vcp-spam
A small tool used to feed spam assassin Bayesian learning filters on machines (currently only known working on Ubuntu 14.04) installed with Vesta Control Panel using the default configuration.

## Disclaimer

While I use this tool in my production on my Ubuntu 14.04 server with Vesta Control Panel installed, it is not guaranteed to work for you or at all. So please be considerate if you decide to employ this tool in your own environment. I make no claims that this software is quality, without bugs, or secure in any fashion.

## What does it do?

`vcp-spam` updates spam assassin filters and scans all of the user folders in /home (the default location that Vesta Control Panel allocates users). It then gets all mailboxes for each user and each domain for a special folder for user-marked junk mail, then feeds these locations into spamassassin's sa-learn commnad.

eg:

`/home/someuser/mail/somedomain.com/somemailuser/.Junk`

Or for mail that was erroneously delivered as Spam and manually unmarked by the user by dragging/copying the messages into the NotSpam folder (Ham):

`/home/someuser/mail/somedomain.com/somemailuser/.NotJunk`

These folders would be made in the roundcube web interface for IMAP.

## How to use it?

Copy it to `/usr/local/bin` with appropriate permissions and set up the cron. It will by default look for a configuration file @ `/etc/vcp-spam.conf.json` to specify home directories/users that should not be scanned by the tool:

##### Config JSON:

```JSON
{
    "skiplist": [
        "backup",
        "somenonvcpuser"
    ]
}
```

##### Cron settings:

```BASH
0 6,18 * * * /usr/local/bin/vcp-spam >/var/log/vcp-update.log 2>&1 # Twice-daily spam learning 6am/6pm
0 0 1 * * /usr/local/bin/vcp-spam --clean >/var/log/vcp-update.log 2>&1 # Once-monthly (1st) spam learning with junk box deletion
```

##### Commandline Variables:

`--clean`

This will delete the messages in the junk box folders after performing the learn actions. The CRON listed above is set to perform this action on the 1st of the month.

## Contribute?

Send me a pull request.