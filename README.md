# gosukebot
A bot for slack

This was sort of a half-abandoned attempt to clean up the earlier [gogobot](https://github.com/KOMON/gogobot) project, which was a rewrite-it-in-go version of the even earlier [jojobot](https://github.com/KOMON/jojobot) written in Ruby.

The main focus of this cleanup was to make the logic of some of the magic handlers (especially stats) a little less eye-gouging to look at, and to make the handlers a bit more modular overall.

Another reason for the hard fork was to straighten up the API of the handlers to make it easier for my friends to write their own handlers for Jojo, but life got in the way, the amount of deckbrewing we were doing on the slack channel slowed down, and they didn't have much investment in writing new functionality for a little slack bot that primarily searched for magic cards.

Gosukebot still lives on, although I've always had plans to fix him up a little bit. He's a little underused in the slack channel, but life goes in cycles and he may see the light of day soon enough.

There was an attempt to rewrite him in Racket scheme, see [giornobot](https://github.com/KOMON/giornobot) for that, although it never got very far.
