const apiUrl = `http://${window.location.host}/api/v1`
const gkbResponse = document.getElementById('gkbResponse')

const sendCommand = function() {
    const cmd = document.getElementById('cmd').value
    const data = { cmd: cmd };

    fetch(apiUrl + "/commands", {
      method: 'POST', // or 'PUT'
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
      .then((response) => response.json())
      .then((data) => {
        printGkbResponse(gkbResponse, cmd, data)
      })
      .catch((error) => {
        printGkbResponse(gkbResponse, cmd, data)
      });
    
}

function printGkbResponse(spanEl, cmd, jsonData) {
  gkbResponse.innerText = cmd + " => " + JSON.stringify(jsonData)
}

const sendBtn = document.getElementById('send');
sendBtn.addEventListener('click', sendCommand)

