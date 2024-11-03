# readflow

Track your Kobo reads on Anilist and Hardcover using Calibre-Web and Calibre databases

> [!WARNING]
> This project is an early-stages WIP

## Pre-Requisites for this to actually be useful

Admittedly this is quite a niche tool. It's only really useful in the following scenario:

1. You own a Kobo eReader
2. You store all of your books in a [Calibre](https://calibre-ebook.com/) library
3. You run a [Calibre-Web](https://github.com/janeczku/calibre-web) server
4. You have configured your Kobo eReader to use the Calibre-Web
[Kobo Integration](https://github.com/janeczku/calibre-web/wiki/Kobo-Integration)
as the API endpoint


## Installation

1. Download the [GitHub Release binaries](https://github.com/RobBrazier/readflow/releases/latest)
compiled for MacOS, Linux and Windows (arm64, amd64 and i386)
2. Install with `go install`

    ```bash
    go install github.com/RobBrazier/readflow@latest
    ```

## Setup

Once installed, you'll need to configure the CLI.
This can be done by following the instructions in the below command:

```bash
readflow setup
```

This will take you through a guided form to get all the information required
for the application

## Sync

Once setup has been completed, you can run

```bash
readflow sync
```

And this will pull recent reads and sync them to the providers configured

> [!IMPORTANT]
> Currently you are required to set calibre identifiers for the providers in the
> books you want to sync. Any books missing these will be skipped.
>
> e.g. [hardcover:pride-and-prejudice](https://hardcover.app/books/pride-and-prejudice)
or [anilist:53390](https://anilist.co/manga/53390/Attack-on-Titan/)

## Running on a Schedule

This is a `oneshot` CLI tool, so if you want to run it frequently, you'll need
to configure a cron job

On Linux systems this can be done with

```bash
crontab -e
```

As an example, the cron job I use is:

```crontab
0 * * * * /usr/local/bin/readflow sync 2>> /var/log/readflow.log
```
