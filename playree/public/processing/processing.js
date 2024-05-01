window.addEventListener("DOMContentLoaded", (_) => {
	let websocket = new WebSocket("ws://" + window.location.host + "/start-processing");
	let room = document.getElementById("status");
	let path = window.location.pathname;
  
	websocket.addEventListener("message", function (e) {
	  let data = e.data;
  
	  if (/^PLAYLIST URL:\s*(.+)$/.test(data)) {
		websocket.close();
		let url = data.match(/PLAYLIST URL:\s*(.+)$/)[1];
		window.location.href = url;
		return;
	  }
  
	  let p = document.createElement("p");
	  p.innerHTML = `<strong>${data}</strong>`;
	  room.append(p);
	  room.scrollTop = room.scrollHeight;
	});
});
