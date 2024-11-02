# readflow

Track your Kobo reads on Anilist and Hardcover using Calibre-Web and Calibre databases

> [!WARNING]
> This project is an early-stages WIP

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
