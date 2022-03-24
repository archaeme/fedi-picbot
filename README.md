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
./fedi-picbot post [-config=path/to/config.ini] [-sources=path/to/sources.txt]
```

### sources.txt format
Each line in sources.txt is as follows:
```
<image-url> <sensitive-bool> <url-to-source> 
```

Each component is seprated with tabs, so it can be thought of as a TSV file.

## Todo
- Add duplicate detection
- Maybe change sources.txt to a sqlite database or similar