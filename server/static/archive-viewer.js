function populateChannels() {
  fetch("/channels").then((response) => {
    if (!response.ok) {
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
  }).catch((error) => {
    document.getElementById("error").style.display = "block";
    document.getElementById("select-params").style.display = "none";
    console.log(error);
  });
}

function renderMessage(message) {
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
    let reaccContainer = document.createElement("div");
    for (let name of Object.keys(message.reacts)) {
      let reacc = document.createElement("span");
      reacc.innerText = `:${name}: (${message.reacts[name].length})`
      reacc.style.marginRight = "20px";
      reaccContainer.appendChild(reacc);
    }
    msgContainer.appendChild(reaccContainer);
  }
  return msgContainer;
}

function loadMessages(channel, from, to) {
  document.getElementById("loading").style.display = "";
  document.getElementById("select-params").style.display = "none";
  document.getElementById("nomessages").style.display = "none";
  fetch(`/messages?channel=${channel}&from=${from.getTime()}&to=${to.getTime()}`).then((response) => {
    if (!response.ok) {
      throw new Error(`GET /messages failed: ${response.status} ${response.statusText}`);
    }
    return response.json();
  }).then((messages) => {
    document.getElementById("loading").style.display = "none";
    document.getElementById("error").style.display = "none";
    for (let message of messages) {
      let msgContainer = renderMessage(message);
      if (message.thread) {
        let showThread = document.createElement("button");
        const showText = `Show ${message.thread.length} repl${message.thread.length === 1 ? "y" : "ies"}`;
        const hideText = `Hide ${message.thread.length} repl${message.thread.length === 1 ? "y" : "ies"}`;
        showThread.innerText = showText;
        let threadMessages = document.createElement("div");
        threadMessages.style.display = "none";
        showThread.onclick = () => {
          if (threadMessages.style.display === "none") {
            threadMessages.style.display = "block";
            if (threadMessages.children.length === 0) {
              for (let reply of message.thread) {
                let replyContainer = renderMessage(reply);
                replyContainer.style.marginBottom = "10px";
                threadMessages.appendChild(replyContainer);
              }
            }
            showThread.innerText = hideText;
          } else {
            threadMessages.style.display = "none";
            showThread.innerText = showText;
          }
        };
        showThread.style.marginLeft = "40px";
        showThread.style.marginBottom = "10px";
        threadMessages.style.marginLeft = "40px";
        msgContainer.appendChild(showThread);
        msgContainer.appendChild(threadMessages);
      }
      msgContainer.style.marginBottom = "20px";
      document.getElementById("messages").appendChild(msgContainer);
    }
    if (messages.length === 0) {
      document.getElementById("nomessages").style.display = "";
    }
  }).catch((error) => {
    document.getElementById("loading").style.display = "none";
    document.getElementById("error").style.display = "block";
    console.log(error);
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

function uploadZip() {
  const files = document.getElementById("upload").files;
  if (files.length === 0) {
    return;
  }
  const formData = new FormData();
  for (let i = 0; i < files.length; i++) {
    formData.append(i.toString(), files[i]);
  }
  document.getElementById("uploading").style.display = "";
  document.getElementById("upload-error").style.visibility = "hidden";
  fetch("/upload", {
    method: "POST",
    body: formData
  }).then((response) => {
    if (!response.ok) {
      throw new Error(`POST /upload failed: ${response.status} ${response.statusText}`);
    }
    document.getElementById("uploading").style.display = "none";
    document.getElementById("upload").value = "";
    window.location.reload()
  }).catch((error) => {
    document.getElementById("uploading").style.display = "none";
    document.getElementById("upload-error").style.visibility = "visible";
    console.log(error);
  });
}
