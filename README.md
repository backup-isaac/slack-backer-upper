# slack-backer-upper

Want access to your Slack message history, but can't pay for premium Slack? If so, this is the solution for you.

## Usage
```
-d string
      a directory to import
-z string
      a zip file to import
```
If a directory or zip file name is passed, the corresponding Slack backup is imported.
If neither option is provided, an HTTP server is started.

The server will feature:
- Simple front end for viewing message history
- API for retrieving messages
- API endpoint for manually adding a workspace export
- Automatic requests to the Slack API every day/week to make backups of new messages
