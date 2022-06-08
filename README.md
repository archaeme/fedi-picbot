# fedi-picbot

An image bot for Mastodon and Pleroma.

## Features
- Fetch images from URLs
- Provides a link back to the image source
- Can mark images as sensitive

## Usage
Registering the bot on the instance:
```bash
./fedi-picbot register -server="https://instance.tld"
```

Making a post:
```bash
./fedi-picbot post [-dir=<path to dir with sources.txt and config.ini>]
```

**Note**: You can also specify the locations of `sources.txt` and `config.ini` separately using the `-sources` and `-config` flags respectively.

### sources.txt format
Each line in sources.txt is as follows:
```
<image-url> <sensitive-bool> <url-to-source> 
```

Notes:
- `<image-url>` can be an HTTP(S) url or a local filename
- Each component is separated by tabs

## Todo
- Add duplicate detection
