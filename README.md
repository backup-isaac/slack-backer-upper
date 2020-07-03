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

Work in progress:
- API endpoint for manually adding a workspace export
- Automatic requests to the Slack API every day/week to make backups of new messages

## API

### `GET /channels`
Retrieves the list of channel names sorted in alphabetical order.

#### URL Parameters
None.

#### Response
Field | Data type | Description
-|-|-
top level field | String array | The channel names

#### Example
```json
GET /channels
200 OK
["general", "random", "software"]
```

### `GET /messages`
Retrieves messages from a channel sorted in chronological order.

#### URL Parameters
Name | Data type | Required
-|-|-|-
channel | String | yes
from | UNIX millisecond timestamp | yes
to | UNIX millisecond timestamp | yes

#### Response
Field | Data type | Description
-|-|-
top level field | `ParentMessage` array | The messages from the channel

#### Example
```json
GET /messages?channel=announcements&from=1588392000000&to=1593662400000
200 OK
[{
  "attachments": null,
  "reacts": {
    "peak-performance": ["Ryan Babaie"],
    "sad-solar-boi": ["Ryan Babaie"],
    "sr3": [
      "Matthew Marting",
      "Steven Licciardello",
      "Arvin Ajmani",
      "Tara Chan",
      "Ryan Babaie"
    ],
    "yeet": ["Joshua Hoffman"]
  },
  "text": "Hello solar raycers! If you are one of our wonderful new  graduates, please reacc to this!",
  "thread": null,
  "timestamp": 1588412758,
  "user": "AJ Wasserman"
}, {
  "attachments": null,
  "reacts": null,
  "text": "<!channel> hey all! If you are a new grad and didn't reacc to my message above, please do!",
  "thread": [{
    "attachments": null,
    "reacts": null,
    "sent": false,
    "text": "@ryan.babaie @Joshua Hoffman @Tara Chan",
    "timestamp": 1588557488,
    "user": "Matthew Marting"
  }, {
    "attachments": null,
    "reacts": null,
    "sent": false,
    "text": "Ty for the save. Wasnt paying attention May 2nd.",
    "timestamp": 1588676894,
    "user": "Joshua Hoffman"
  }],
  "timestamp": 1588477847,
​​  "user": "AJ Wasserman"
}]
```

### Data Types

#### `Attachment`
Field | Data type | Description
-|-|-
fallback | String | Text to display if the URL can't be reached
from_url | String | The URL of the attached file or link
title | String | The title of the attached file or link

#### `ParentMessage`
Field | Data type | Description
-|-|-
attachments | `null` or `Attachment` array | Files or links attached to the message
reacts | `null` or `Reacts` object | Reactions to the message
text | String | The text body of the message
thread | `null` or `ThreadMessage` array | Thread replies to the message in sorted chronological order
timestamp | UNIX second timestamp | The time when the message was sent
user | String | The user who sent the message

#### `Reacts`
Field | Data type | Description
-|-|-
\<react name> | String array | The users who reacted with \<react name>

#### `ThreadMessage`
Field | Data type | Description
-|-|-
attachments | `null` or `Attachment` array | Files or links attached to the message
reacts | `null` or `Reacts` object | Reactions to the message
sent | Boolean | Whether or not the message was also sent to the channel
text | String | The text body of the message
timestamp | UNIX second timestamp | The time when the message was sent
user | String | The user who sent the message
