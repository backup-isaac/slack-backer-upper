function populateChannels() {
  fetch("/channels").then((response) => {
    if (!response.ok) {
      document.getElementById("error").style.display = "block";
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
