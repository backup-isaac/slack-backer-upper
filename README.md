# slack-backer-upper

Want access to your Slack message history, but can't pay for premium Slack? If so, this is the solution for you.

Code is disgusting right now, will be cleaned up 🔜™️

Currently this is a command line program that takes a workspace export downloaded from Slack and loads its contents into a SQLite database. Next it's going to be turned into a web server featuring:
- API endpoint for manually adding a workspace export
- API for retrieving messages
- Simple front end for viewing message history
- Automatic requests to the Slack API every day/week to make backups of new messages

