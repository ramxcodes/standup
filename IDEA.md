## Idea 

A Simple TUI based terminal CMD that helps you out prepare quickly for standup.

### Working 

Just type

```bash
git stand-up
```

And it will return 

```shell
short-commit-hash - A long commit text but not too long <3 Days ago> Ramxcodes 

AI GENERATED SUMMARY - 

I've created XYZ yesterday and working on abc will try to comlete it today.
```

You can run this command inside a specific repo folder. Or root.

If specific repo folder then it will show git logs for that repo only. Otherwise it will show logs for all the repos (where git is initialized) and group logs under the repo.

After the commits will automatically will generate an AI generated summary 

```bash
Repo 1
short-commit-hash - A long commit text but not too long <3 Days ago> Ramxcodes 
short-commit-hash - A long commit text but not too long <3 Days ago> Ramxcodes 

Repo 2

short-commit-hash - A long commit text but not too long <3 Days ago> Ramxcodes 
short-commit-hash - A long commit text but not too long <3 Days ago> Ramxcodes 

AI GENERATED SUMMARY - 

I've created XYZ yesterday and working on abc will try to comlete it today.

```

### Tech Stack

- Go
- Bubble tea
- Gemini AI Model for summary

### Flags

```bash
standup 
standup -d <number> [can take days from current time, default is current time to last date]
standup -a <name> [filter logs by author, default is current author]
standup --set-api-key <string> [can set gemini api key]
standup --set-model-name <string> [can set model name]
standup --disable-ai
standup --enable-ai
standup -h or --help
standup -v or --version
```
