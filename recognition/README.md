Recognize plugin
----------------

This plugin listens for something like:

    !recognize @person1 for [some great reasons]
    !recognize @person1 and @person2 for [other reasons]
    !recognize @person3, @person5 and @person22 for [many great feats]

The bot will then announce the recognition in the configured
#recognitions channel. It will then listen on reactions for that
announcement and keep track of who reacted.

In the end, another process can take those numbers and recognitions,
their sender/initiator, list of recipients, and decide to feed it to a
concrete recognition (based on bonus or whatnot) system.

This is a nice example of the use of reactions as buttons.

Configuration
-------------

Config keys for this plugin look like:

    {
      ...
      "Recognition": {
        "domain_restriction": "@yourdomain.com",
        "channel": "#recognitions"
      },
      ...
    }

`domain_restriction` checks that users voting have an e-mail ending
with the given string.

`channel` is the channel where recognitions will be broadcast.