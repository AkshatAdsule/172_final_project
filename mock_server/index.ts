import { WebSocketServer } from 'ws'

const wss = new WebSocketServer({port: 8080})


wss.on("connection", (ws) => {
  let lat = 38.5417957
  let lon = -121.7756125
  const speed = 0.0001

  setInterval(() => {
    const angle = ((Math.PI * Math.random()) / 2) // 0-pi/2
    const dlat = speed * Math.sin(angle)
    const dlon = speed * Math.cos(angle)
    
    lat += dlat;
    lon += dlon;

    ws.send(JSON.stringify({
      lat,
      lon
    }))
  }, 1000);
})
