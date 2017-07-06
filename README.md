# gobot
## PhD work ported to Go

![botmover logo](mstile-310x310.png)

`gobot` is a simplified prototype of a cluster of intelligent agents roaming a virtual world (theoretically either in Second Life or OpenSimulator). Users can visually programme the paths/activities such agents are supposed to follow by __visually__ dragging simple markers around the virtual world, without the need of learning any computer language. The principle is similar, and based on, several computer games where the player also gives visual commands to an AI controlling the players' units (like in Command & Conquer, StarCraft, or The Sims).

Right now, the Go application just deals with visualising the support database and adds some tools to give manual commands to the in-world markers (known simply as 'cubes' in `gobot`); a prototype run of the engine is being ported from previous code originally in PHP. Go, because of its very simple and intuitive approach to multithreading, will be used to allow each agent to get their own thread running the AI engine.
 
One day, the whole thing will be fully described here!
For now, all you need to know is that this works only on [OpenSimulator](http://opensimulator.org) since Second Life does not allow intelligent agents to be created and run in the same way as OpenSimulator does.

## Server-Side Configuration

- Install `gobot` as any other Go application (`go get github.com/GwynethLlewelyn/gobot` should do the trick)
- Create a database in SQLite3 using `database/schema.sql`
- Create a new user in SQLite3 with an email address and a MD5-hashed password (something like `insert into Users (Email, Password) Values ('valid@email.address', '4405c5984441a1b86bec717dc063ca46');`), you'll need at least one user to login to the backoffice; use `echo "password"|md5sum` or an online MD5 generator to get a valid password hash; afterwards, you can add more users manually 
- Remember the installation path and change `config.toml` accordingly! (you should also set an URL to grab a map tile from your OpenSimulator environment)
- Note:
 - The directories `lib/` and `templates/` only have static content, so either you configure `config.toml` to  point to the right directories (if running `gobot` as a standalone Go application) or you get these directories directly served by your webserver/reverse proxy/whatever
- Point your browser to the URL of the `gobot` appplication, login with the email/password, and try things out on the menus

If you're placing `gobot` behind a nginx server, [this is the configuration you'll need](nginx-config.md). Note that Go is wonderful as it includes its own webserver, so running it behind a 'real' web server is not necessary, although a real web server should be able to provide things like caching and direct serving of static content (images, JS, CSS...) for the backoffice, to make it even faster.

## Dependencies

These will be installed, *but* you *should* make sure you have the latest versions of them! (use git pull on the directories)

- [agGrid](https://www.ag-grid.com/)
- [Bootstrap](http://getbootstrap.com/) and [Bootstrap Dialog](https://nakupanda.github.io/bootstrap3-dialog/)
- StartBootstrap's [SB Admin 2](https://startbootstrap.com/template-overviews/sb-admin-2/) template
- [Leaflet.js](http://leafletjs.com/) (called remotely, no need to worry about it)
- [Leaflet.vector-markers](https://github.com/hiasinho/Leaflet.vector-markers) (to get cute markers for Leaflet)
- [Prism.js](http://prismjs.com/) (called remotely, no need to worry about it)
- Gravatar Hovercard support (also called remotely except for one file)

## Configuration on the Virtual World

You're on your own, I haven't written it yet, but here's a bit of what you need to do.

Move over to the _LSL scripts_ menu option. These should pre-generate correctly configured LSL scripts for the many in-world scripts. (Currently just one, I'll be adding the next few in a jiffy.) Copy & paste them inside the LSL editor on the SL Viewer Application of your choice, and drop them inside the appropriate cubes.

You should at least place on OpenSimulator one box with a 'Bot Controller' LSL script (you can have multiple Bot Controllers if you wish, spread out over several sims or grids, and control them all from a single interface). The Bot Controller item will only launch new agents, it does not do anything else (all AI logic is actually done in Go, and the actual commands given to the agents are directly made from `gobot` to each individual agent).

Agents will try to use the Energy, Happiness, and Money cubes; you can have as many of those as you wish. It's not obvious, but the Description field will tell which class of agent should use that particular cube. I think that the default is 'peasant' but you will have to figure it out on your own. The three types of cubes will have hovertext to show its current values, and you can set them up as you wish (there is just one script for all three types anyway).

The tough bit will be configuring Agents. This requires a special notecard with the avatar appearance and lots of blah blah blah that will one day be explained here.

More to come...

## Hidden Easter Eggs

Sending a SIGHUP will reload the configuration file (as expected)

SIGUSR1 will send a randomly generated female name to appear on the Engine page; SIGUSR2 will place a random country. Why? Well, this was the only way to get *something* to appear there while testing the code (and without creating a new backoffice just for testing purposes). Now if I only came up with pipes and more esoteric stuff... 😉

## License

This is currently copyrighted until I finish my PhD, then you can copy it at will. If you wish to use my code **now**, then please ask me before! Thank you!