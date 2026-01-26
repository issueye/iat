export class WSClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.handlers = new Map();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
          console.log('WS Connected');
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data);
            const handler = this.handlers.get(msg.type) || this.handlers.get('*');
            if (handler) handler(msg);
          } catch (e) {
            console.error('WS Parse Error', e);
          }
        };

        this.ws.onclose = () => {
          console.log('WS Closed');
          this.attemptReconnect();
        };

        this.ws.onerror = (err) => {
          console.error('WS Error', err);
          reject(err);
        };
      } catch (e) {
        reject(e);
      }
    });
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.pow(2, this.reconnectAttempts) * 1000;
      console.log(`WS Reconnecting in ${delay}ms...`);
      setTimeout(() => this.connect(), delay);
    }
  }

  on(type, handler) {
    this.handlers.set(type, handler);
  }

  send(msg) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  close() {
    if (this.ws) {
      this.ws.close();
    }
  }
}
