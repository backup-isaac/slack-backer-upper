function populateChannels() {
  fetch("/channels").then((response) => {
    if (!response.ok) {
      document.getElementById("error").style.display = "block";
      document.getElementById("select-params").style.display = "none";
      throw new Error(`GET /channels failed: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }).then((channels) => {
    for (let channel of channels) {
      let option = document.createElement("option");
      option.value = channel;
      option.text = `#${channel}`;
      document.getElementById("channel").appendChild(option);
    }
  });
}

function loadMessages(channel, from, to) {
  fetch(`/messages?channel=${channel}&from=${from.getTime()}&to=${to.getTime()}`).then((response) => {
    if (!response.ok) {
      document.getElementById("error").style.display = "block";
      document.getElementById("select-params").style.display = "none";
      throw new Error(`GET /messages failed: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }).then((messages) => {
    document.getElementById("error").style.display = "none";
    document.getElementById("select-params").style.display = "none";
    for (let message of messages) {
      let msgContainer = document.createElement("div");
      let msgTime = document.createElement("span");
      const time = new Date(message.timestamp * 1000);
      msgTime.innerText = time.getFullYear().toString().padStart(4, "0") + "-"
        + (time.getMonth() + 1).toString().padStart(2, "0") + "-"
        + time.getDate().toString().padStart(2, "0") + " "
        + time.getHours().toString().padStart(2, "0") + ":"
        + time.getMinutes().toString().padStart(2, "0");
      msgContainer.appendChild(msgTime);
      let msgUser = document.createElement("strong");
      msgUser.style.marginLeft = "20px";
      msgUser.innerText = message.user;
      msgContainer.appendChild(msgUser);
      let msgBody = document.createElement("p");
      msgBody.innerText = message.text;
      msgContainer.appendChild(msgBody);
      if (message.attachments) {
        for (let attachment of message.attachments) {
          let div = document.createElement("div");
          let attach;
          if (attachment.from_url) {
            attach = document.createElement("a");
            attach.innerText = attachment.title || attachment.from_url;
            attach.href = attachment.from_url;
          } else {
            attach = document.createElement("p");
            attach.innerText = attachment.fallback;
          }
          div.appendChild(attach);
          div.style.marginLeft = "40px";
          msgContainer.appendChild(div);
        }
      }
      if (message.reacts) {
        for (let name of Object.keys(message.reacts)) {
          let reacc = document.createElement("span");
          reacc.innerText = `:${name}: (${message.reacts[name].length})`
          reacc.style.marginRight = "20px";
          msgContainer.appendChild(reacc);
        }
      }
      msgContainer.style.marginBottom = "20px";
      document.getElementById("messages").appendChild(msgContainer);
    }
  });
}

let selectedChannel = "";
let selectedFrom = new Date(0);
let selectedTo = new Date(0);

function tryLoadMessages() {
  const channel = document.getElementById("channel").value;
  if (!channel) {
    return;
  }
  const fromStr = document.getElementById("from").value;
  if (!fromStr) {
    return;
  }
  const toStr = document.getElementById("to").value;
  if (!toStr) {
    return;
  }
  const fromDay = fromStr.split("-");
  const from = new Date(fromDay[0], fromDay[1] - 1, fromDay[2]);
  const toDay = toStr.split("-");
  const to = new Date(toDay[0], toDay[1] - 1, toDay[2]);
  if (from.getTime() >= to.getTime()
    || (channel === selectedChannel
      && from.getTime() === selectedFrom.getTime()
      && to.getTime() === selectedTo.getTime())) {
    return;
  }
  document.getElementById("messages").textContent = "";
  loadMessages(channel, from, to);
  selectedChannel = channel;
  selectedFrom = from;
  selectedTo = to;
}
