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
      console.error(`GET /messages failed: ${response.status} ${response.statusText}`);
      return response.text();
    }
    return response.json();
  }).then(console.log);
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
  selectedChannel = channel;
  selectedFrom = from;
  selectedTo = to;
  loadMessages(channel, from, to);
}
